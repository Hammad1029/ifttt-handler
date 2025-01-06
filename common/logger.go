package common

import (
	"context"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

func CreateLogrus() *logrus.Logger {
	logger := logrus.New()

	logger.SetFormatter(&logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime: "timestamp",
			logrus.FieldKeyMsg:  "message",
		},
	})
	logger.SetOutput(os.Stdout)

	return logger
}

func LogWithTracer(user string, msg string, data any, err bool, ctx context.Context) {
	// ctxState := GetCtxState(ctx)
	// logCtx, ok := ctxState.Load(ContextLogger)
	// if !ok {
	// 	return
	// }
	// logger, ok := logCtx.(*logrus.Logger)
	// if !ok {
	// 	return
	// }
	// tracer, ok := ctxState.Load(ContextTracer)
	// if !ok {
	// 	return
	// }
	// logStage, ok := ctxState.Load(ContextLogStage)
	// if !ok {
	// 	return
	// }

	// logFields := logrus.Fields{
	// 	"stage":  logStage,
	// 	"tracer": tracer,
	// 	"user":   user,
	// 	"data":   data,
	// }
	// if err {
	// 	logger.WithFields(logFields).Error(msg)
	// } else {
	// 	logger.WithFields(logFields).Info(msg)
	// }
}

type LogEnd struct {
	ApiPath          string         `json:"api_path" mapstructure:"api_path"`
	ApiName          string         `json:"api_name" mapstructure:"api_name"`
	Start            time.Time      `json:"start" mapstructure:"start"`
	End              time.Time      `json:"end" mapstructure:"end"`
	ExecutionTime    uint64         `json:"execution_time" mapstructure:"execution_time"`
	InternalExecTime uint64         `json:"internal_exec_time" mapstructure:"internal_exec_time"`
	ExternalExecTime uint64         `json:"external_exec_time" mapstructure:"external_exec_time"`
	RequestData      map[string]any `json:"request_data" mapstructure:"request_data"`
	Error            string         `json:"error" mapstructure:"error"`
}
