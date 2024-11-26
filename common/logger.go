package common

import (
	"context"
	"ifttt/handler/application/config"

	"github.com/sirupsen/logrus"
	"github.com/yukitsune/lokirus"
)

func CreateLogrus() *logrus.Logger {
	opts := lokirus.NewLokiHookOptions().
		WithLevelMap(lokirus.LevelMap{logrus.PanicLevel: "critical"}).
		WithFormatter(&logrus.JSONFormatter{}).
		WithStaticLabels(lokirus.Labels{
			"app": "ifttt/handler",
		})

	hook := lokirus.NewLokiHookWithOpts(
		config.GetConfigProp("loki.host"),
		opts,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel)

	logger := logrus.New()
	// logger.SetFormatter(&logrus.TextFormatter{
	// 	ForceColors:   true,
	// 	FullTimestamp: true,
	// })
	logger.AddHook(hook)

	return logger
}

func LogWithTracer(user string, msg string, data any, err bool, ctx context.Context) {
	logCtx, ok := GetCtxState(ctx).Load(ContextLogger)
	if !ok {
		return
	}
	logger, ok := logCtx.(*logrus.Logger)
	if !ok {
		return
	}
	tracer, ok := GetCtxState(ctx).Load(ContextTracer)
	if !ok {
		return
	}

	logFields := logrus.Fields{
		"tracer": tracer,
		"user":   user,
		"data":   data,
	}
	if err {
		logger.WithFields(logFields).Error(msg)
	} else {
		logger.WithFields(logFields).Info(msg)
	}
}
