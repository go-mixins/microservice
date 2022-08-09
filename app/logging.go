package app

import (
	"fmt"

	"github.com/evalphobia/logrus_sentry"
	"github.com/getsentry/raven-go"
	logrusLogger "github.com/go-mixins/log/logrus"
	"github.com/sirupsen/logrus"
	graylog "gopkg.in/gemnasium/logrus-graylog-hook.v2"
)

func (app *App) connectLogs() error {
	cfg := app.Config
	loggerGetter, ok := app.Logger.(interface{ GetLogger() *logrus.Logger })
	if !ok {
		return fmt.Errorf("provided logger is not compatible with Logrus")
	}
	logger := loggerGetter.GetLogger()
	if cfg.Debug {
		logger.SetLevel(logrus.DebugLevel)
	}
	if cfg.GraylogURI != "" {
		hook := graylog.NewAsyncGraylogHook(cfg.GraylogURI, nil)
		logger.Hooks.Add(hook)
		app.flushers = append(app.flushers, hook)
	}
	if cfg.SentryDSN == "" {
		return nil
	}
	client, err := raven.NewWithTags(cfg.SentryDSN, map[string]string{
		"service_name": cfg.ServiceName,
	})
	if err != nil {
		return fmt.Errorf("create raven client: %v", err)
	}
	client.SetEnvironment(cfg.Environment)
	hook, err := logrus_sentry.NewAsyncWithClientSentryHook(client, []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
	})
	if err != nil {
		return fmt.Errorf("create sentry hook: %v", err)
	}
	hook.AddIgnore("server_name")
	logger.AddHook(hook)
	app.flushers = append(app.flushers, hook)
	return nil
}
