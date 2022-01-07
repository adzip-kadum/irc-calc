package log

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TheZeroSlave/zapsentry"
	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/adzip-kadum/irc-calc/version"
)

type Config struct {
	Level   string        `yaml:"level"`
	Time    TimeConfig    `yaml:"time"`
	Message MessageConfig `yaml:"message"`
	Sentry  SentryConfig  `yaml:"sentry"`
}

type TimeConfig struct {
	Key    string `yaml:"key"`
	Layout string `yaml:"layout"`
}

type MessageConfig struct {
	Key string `yaml:"key"`
}

type SentryConfig struct {
	DSN  string            `yaml:"dsn"`
	Tags map[string]string `yaml:"tags"`
}

var (
	DefaultConfig = Config{
		Level: "debug",
	}

	logger *zap.Logger
	hooks  []Hooker
)

func init() {
	err := Init(DefaultConfig)
	if err != nil {
		panic(err)
	}
}

func Init(conf Config) error {
	var level zap.AtomicLevel
	err := level.UnmarshalText([]byte(conf.Level))
	if err != nil {
		return err
	}

	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})

	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return level.Enabled(lvl) && lvl < zapcore.ErrorLevel
	})

	zapconf := zap.NewProductionEncoderConfig()
	if conf.Time.Key != "" {
		zapconf.TimeKey = conf.Time.Key
	}
	if conf.Time.Layout != "" {
		zapconf.EncodeTime = zapcore.TimeEncoderOfLayout(conf.Time.Layout)
	}
	if conf.Message.Key != "" {
		zapconf.MessageKey = conf.Message.Key
	}

	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(zapconf), zapcore.Lock(os.Stderr), highPriority),
		zapcore.NewCore(zapcore.NewJSONEncoder(zapconf), zapcore.Lock(os.Stdout), lowPriority),
	)
	// do not event think to use zap logger sampler!
	// https://github.com/uber-go/zap/blob/master/FAQ.md#why-sample-application-logs
	logger = zap.New(core).WithOptions(zap.AddCallerSkip(1))

	if conf.Sentry.DSN != "" {
		cfg := zapsentry.Configuration{
			Level: zapcore.ErrorLevel, //when to send message to sentry
			Tags:  conf.Sentry.Tags,
		}
		client, err := sentry.NewClient(sentry.ClientOptions{
			Dsn:     conf.Sentry.DSN,
			Release: version.Semver.String(),
		})
		if err != nil {
			Error(err)
		}
		core, err := zapsentry.NewCore(cfg, zapsentry.NewSentryClientFromClient(client))
		//in case of err it will return noop core. so we can safely attach it
		if err != nil {
			Error(err)
		}
		logger = zapsentry.AttachCoreToLogger(core, logger)
	}

	return nil
}

func Logger() *zap.Logger { return logger }

var (
	Int      = zap.Int
	Int32    = zap.Int32
	Float64  = zap.Float64
	String   = zap.String
	Strings  = zap.Strings
	Stringer = zap.Stringer
	Bool     = zap.Bool
	Duration = zap.Duration
	Any      = zap.Any
	Err      = zap.Error
	Object   = zap.Object
)

type Field = zapcore.Field

func Debug(msg string, fields ...Field) {
	logger.Debug(msg, fields...)
	for _, h := range hooks {
		h.Debug(msg, fields...)
	}
}

func Info(msg string, fields ...Field) {
	logger.Info(msg, fields...)
	for _, h := range hooks {
		h.Info(msg, fields...)
	}
}

func Error(err error, fields ...Field) {
	fields = append(fields, Err(err))
	logger.Error(err.Error(), fields...)
	for _, h := range hooks {
		h.Error(err.Error(), fields...)
	}
}

func DebugEnabled() bool {
	return logger.Core().Enabled(zapcore.DebugLevel)
}

func Sync() {
	if err := logger.Sync(); err != nil {
		msg := fmt.Sprintf("[ERROR] log: %s", err)
		if !strings.Contains(msg, "inappropriate ioctl for device") &&
			!strings.Contains(msg, "operation not supported by device") {
			fmt.Fprintf(os.Stderr, "%s\n", msg)
		}
	}
}

func Filename(name string) string {
	return fmt.Sprintf("%s/%s", filepath.Base(filepath.Dir(name)), filepath.Base(name))
}

func AddHook(h Hooker) {
	hooks = append(hooks, h)
}

func DeleteHook(h Hooker) {
	for i, hook := range hooks {
		if h == hook {
			hooks = append(hooks[:i], hooks[i+1:]...)
			DeleteHook(h)
		}
	}
}

func Hooks() []Hooker {
	return hooks
}
