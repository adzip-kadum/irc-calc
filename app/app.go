package app

import (
	"github.com/adzip-kadum/irc-calc/bot"
	"os"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/atomic"

	"github.com/adzip-kadum/irc-calc/errs"
	"github.com/adzip-kadum/irc-calc/log"
	"github.com/adzip-kadum/irc-calc/postgres"
	"github.com/adzip-kadum/irc-calc/probe"
	"github.com/adzip-kadum/irc-calc/version"
)

type Config struct {
	Postgres       postgres.ClientConfig `yaml:"postgres"`
	Bots           []bot.Config          `yaml:"bots"`
	Logger         log.Config            `yaml:"logger"`
	configFileUsed string                `yaml:"-"`
}

func (c *Config) SetConfigFileUsed(s string) {
	c.configFileUsed = s
}

type Application struct {
	conf        *Config
	pool        *postgres.PgxPool
	servers     []server
	healthcheck *probe.HealthCheckService
	ready       *atomic.Bool
}

type server interface {
	Start() error
	Stop() error
}

func New(conf *Config) (app *Application, rerr error) {
	defer errs.Recover(&rerr)
	defer func() {
		if rerr != nil && app != nil && app.pool != nil {
			app.pool.Close()
		}
	}()

	var cwd string
	cwd, rerr = os.Getwd()
	if rerr != nil {
		return
	}

	log.Info(version.Project,
		log.String("version", version.Semver.String()),
		log.String("build", version.BuildTS),
		log.String("commit", version.GitCommit),
		log.String("branch", version.GitBranch),
		log.String("cwd", cwd),
		log.String("config", conf.configFileUsed),
	)

	app = &Application{
		conf:        conf,
		healthcheck: probe.NewHealthCheckService(time.Second),
		ready:       atomic.NewBool(false),
	}

	app.pool, rerr = postgres.NewPgxPool(conf.Postgres)
	if rerr != nil {
		return
	}

	for _, botConf := range conf.Bots {
		b, err := bot.NewBot(botConf, app.pool)
		if err != nil {
			log.Error(err)
		} else {
			app.servers = append(app.servers, b)
		}
	}

	return
}

func (a *Application) Start() error {
	log.Info("applicaiton starting...")

	probe.RegisterLivenessProbe(version.Project, a.livenessProbe)
	probe.RegisterReadinessProbe(version.Project, a.readinessProbe)

	for _, s := range a.servers {
		if err := s.Start(); err != nil {
			return err
		}
	}

	log.Info("application started ok")

	a.ready.Store(true)

	probe.EnableLivenessProbe()
	probe.EnableReadinessProbe()

	return nil
}

func (a *Application) Stop() error {
	log.Info("application stopping...")

	a.ready.Store(false)
	a.healthcheck.Close()

	probe.UnregisterLivenessProbe(version.Project)
	probe.UnregisterReadinessProbe(version.Project)

	for i := len(a.servers); i >= 0; i-- {
		a.servers[i].Stop()
	}

	a.pool.Close()

	log.Info("application stopped ok")

	probe.DisableLivenessProbe()
	probe.DisableReadinessProbe()

	return nil
}

func (a *Application) livenessProbe() error {
	if !a.ready.Load() {
		return errors.New("application is not ready yet")
	}
	return nil
}

func (a *Application) readinessProbe() error {
	if !a.ready.Load() {
		return errors.New("application is not ready yet")
	}
	return nil
}
