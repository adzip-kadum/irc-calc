package postgres

import (
	"context"
	"github.com/jackc/pgx/v4"
	"sync"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/log/zapadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"

	"github.com/adzip-kadum/irc-calc/errs"
	"github.com/adzip-kadum/irc-calc/log"
	"github.com/adzip-kadum/irc-calc/worker"
)

type ClientConfig struct {
	DSN                   string        `yaml:"dsn" jsonschema:"required"`
	PoolMaxConns          int32         `yaml:"poolMaxConns"`
	PoolMinConns          int32         `yaml:"poolMinConns"`
	PoolMaxConnLifetime   time.Duration `yaml:"poolMaxConnLifetime"`
	PoolMaxConnIdleTime   time.Duration `yaml:"poolMaxConnIdleTime"`
	PoolHealthCheckPeriod time.Duration `yaml:"poolHealthCheckPeriod"`
	PoolLazyConnect       bool          `yaml:"poolLazyConnect"`
	LogLevel              string        `yaml:"logLevel"`
}

type PgxPool struct {
	pool   *pgxpool.Pool
	closer *worker.Closer
	sync.RWMutex
}

func NewPgxPool(conf ClientConfig) (*PgxPool, error) {
	for {
		pool, err := newPgxPool(conf)
		if err != nil {
			log.Error(err)
			time.Sleep(2 * time.Second)
			continue
		}
		if err = pool.Pool().Ping(context.Background()); err != nil {
			log.Error(err)
			time.Sleep(2 * time.Second)
			continue
		}
		return pool, nil
	}
}

func (p *PgxPool) Begin(ctx context.Context, opts pgx.TxOptions) (context.Context, error) {
	tx, err := p.Pool().BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return WithTx(ctx, tx), nil
}

func (p *PgxPool) Commit(ctx context.Context) error {
	tx := GetTx(ctx)
	if tx == nil {
		return errors.New("no transaction in context")
	}
	return tx.Commit(ctx)
}

func (p *PgxPool) Rollback(ctx context.Context) error {
	tx := GetTx(ctx)
	if tx == nil {
		return errors.New("no transaction in context")
	}
	return tx.Rollback(ctx)
}

// TODO: commit, rollback

func (p *PgxPool) Pool() *pgxpool.Pool {
	p.RLock()
	defer p.RUnlock()
	return p.pool
}

func (p *PgxPool) Close() {
	p.closer.Close()
	p.pool.Close()
}

type PostgresError struct {
	code int
	err  error
}

func (e *PostgresError) Internal() error {
	return e.err
}

func (e *PostgresError) Error() string {
	return e.err.Error()
}

func (e *PostgresError) Code() int {
	return e.code
}

func NewPostgresError(code int, message string) *PostgresError {
	return &PostgresError{
		code: code,
		err:  errors.New(message),
	}
}

type txKeyType int

const txKey txKeyType = 9876

func WithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

func GetTx(ctx context.Context) pgx.Tx {
	tx, ok := ctx.Value(txKey).(pgx.Tx)
	if !ok {
		return nil
	}
	return tx
}

// TODO
func (p *PgxPool) updateMetrics() {
	// stat := p.Pool().Stat()
	// fmt.Printf("AcquireCount: %d\n", stat.AcquireCount())
	// fmt.Printf("AcquireDiration: %s\n", stat.AcquireDuration())
	// fmt.Printf("AcquiredConns: %d\n", stat.AcquiredConns())
	// fmt.Printf("CanceledAcquiredConns: %d\n", stat.CanceledAcquireCount())
	// fmt.Printf("ConstructingConns: %d\n", stat.ConstructingConns())
	// fmt.Printf("EmptyAcquireCount: %d\n", stat.EmptyAcquireCount())
	// fmt.Printf("IdleConns: %d\n", stat.IdleConns())
	// fmt.Printf("MaxConns: %d\n", stat.MaxConns())
	// fmt.Printf("TotalConns: %d\n", stat.TotalConns())
}

func newPgxPool(conf ClientConfig) (pool *PgxPool, rerr error) {
	defer errs.Recover(&rerr)

	config, err := pgxpool.ParseConfig(conf.DSN)
	if err != nil {
		return nil, err
	}

	if conf.PoolMaxConns == 0 {
		conf.PoolMaxConns = 4
		log.Info("pgxpool set default", log.Int32("PoolMaxConns", conf.PoolMaxConns))
	}

	if conf.PoolMaxConnLifetime == 0 {
		conf.PoolMaxConnLifetime = time.Hour
		log.Info("pgxpool set default", log.Duration("PoolMaxConnLifetime", conf.PoolMaxConnLifetime))
	}

	if conf.PoolMaxConnIdleTime == 0 {
		conf.PoolMaxConnIdleTime = 30 * time.Minute
		log.Info("pgxpool set default", log.Duration("PoolMaxConnIdleTime", conf.PoolMaxConnIdleTime))
	}

	if conf.PoolHealthCheckPeriod == 0 {
		conf.PoolHealthCheckPeriod = time.Minute
		log.Info("pgxpool set default", log.Duration("PoolHealthCheckPeriod", conf.PoolHealthCheckPeriod))
	}

	config.MaxConns = conf.PoolMaxConns
	config.MinConns = conf.PoolMinConns
	config.MaxConnIdleTime = conf.PoolMaxConnIdleTime
	config.MaxConnLifetime = conf.PoolMaxConnLifetime
	config.HealthCheckPeriod = conf.PoolHealthCheckPeriod

	// TODO
	//config.ConnConfig.RuntimeParams = map[string]string{}
	//config.ConnConfig.ConnectTimeout = ???

	// callbacks
	config.ConnConfig.ValidateConnect = validateConnect
	config.ConnConfig.OnNotice = onNotice
	//config.ConnConfig.OnNotification = onNotification
	config.ConnConfig.AfterConnect = afterConnectPgconn
	config.BeforeConnect = beforeConnect
	config.AfterConnect = afterConnect
	config.BeforeAcquire = beforeAcquire
	config.AfterRelease = afterRelease

	// logger
	if conf.LogLevel != "" {
		logLevel, err := pgx.LogLevelFromString(conf.LogLevel)
		if err != nil {
			rerr = errors.Wrap(err, "pgxpool")
			return
		}
		config.ConnConfig.LogLevel = logLevel
	}
	config.ConnConfig.Logger = zapadapter.NewLogger(log.Logger().With(log.String("name", "postgres")))

	p, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		rerr = errors.Wrap(err, "pgxpool")
		return
	}

	log.Info("pgxpool connected")

	pool = &PgxPool{
		pool:   p,
		closer: worker.NewCloser(context.Background(), 1),
	}

	// TODO: interval
	go worker.Worker(pool.closer.Context, "pgxpool", 2*time.Second, pool.updateMetrics, nil, pool.closer.WaitGroup)

	return
}

func beforeConnect(ctx context.Context, connConf *pgx.ConnConfig) error {
	//log.Debug("pgx", zap.String("callback", "BeforeConnect"))
	return nil
}

// TODO
func afterConnect(ctx context.Context, conn *pgx.Conn) error {
	//log.Debug("pgx", zap.String("callback", "AfterConnect"))
	return nil
}

// TODO
func afterConnectPgconn(ctx context.Context, pgconn *pgconn.PgConn) error {
	//log.Debug("pgconn", zap.String("callback", "AfterConnect"))
	return nil
}

// TODO
func beforeAcquire(ctx context.Context, conn *pgx.Conn) bool {
	//log.Debug("pgx", zap.String("callback", "BeforeAcquire"))
	return true
}

// TODO
func afterRelease(conn *pgx.Conn) bool {
	//log.Debug("pgx", zap.String("callback", "AfterRelease"))
	return true
}

// TODO
func validateConnect(ctx context.Context, pgconn *pgconn.PgConn) error {
	//log.Debug("pgconn", zap.String("callback", "ValidateConnect"))
	return nil
}

// TODO
func onNotice(pgconn *pgconn.PgConn, notice *pgconn.Notice) {
	//log.Debug("pgconn", zap.String("callback", "OnNotice"), zap.Any("notice", notice))
}

// NOTIFY/LISTEN
// TODO
// func onNotification(pgconn *pgconn.PgConn, notification *pgconn.Notification) {
// 	//log.Debug("pgconn", zap.String("callback", "OnNotification"), zap.Any("notice", notification))
// }

func readInt(ctx context.Context, conn *pgxpool.Conn, query string, args ...interface{}) (int, error) {
	row := conn.QueryRow(ctx, query, args...)
	value := 0
	err := row.Scan(&value)
	return value, err
}

func txReadInt(ctx context.Context, tx pgx.Tx, query string, args ...interface{}) (int, error) {
	row := tx.QueryRow(ctx, query, args...)
	value := 0
	err := row.Scan(&value)
	return value, err
}

// для тестов
var (
	testPgxpoolOnce  sync.Once
	testPool         *PgxPool
	testPoolErr      error
	TestPoolLogLevel = "trace"
)

func GetTestPgxPool(dsn string) (*PgxPool, error) {
	if dsn == "" {
		dsn = "postgres://root:root@localhost:5431/sbermarket_aliexpress?sslmode=disable"
	}
	testPgxpoolOnce.Do(func() {
		testPool, testPoolErr = NewPgxPool(ClientConfig{
			DSN:      dsn,
			LogLevel: TestPoolLogLevel,
		})
	})
	return testPool, testPoolErr
}

type Options struct {
	withExplain bool

	// results
	Explain []Explain
}

type Explain struct {
	Query string
	Args  []interface{}
	Plan  string
	Err   error
}

type Option func(*Options)

func (o *Options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

func WithExplain() Option {
	return func(o *Options) {
		o.withExplain = true
	}
}

func (o *Options) explain(ctx context.Context, conn *pgxpool.Conn, query string, args ...interface{}) error {
	if o.withExplain {
		e := &Explain{
			Query: query,
			Args:  args,
		}
		query = "EXPLAIN " + query
		row := conn.QueryRow(ctx, query, args...)
		e.Err = row.Scan(&e.Plan)
		o.Explain = append(o.Explain, *e)
		return e.Err
	}
	return nil
}
