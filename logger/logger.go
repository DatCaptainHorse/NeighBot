package logger

import "go.uber.org/zap"

var Sugar *zap.SugaredLogger

func InitLogger() error {
	logger, err := zap.NewProduction()
	if err != nil {
		return err
	}
	Sugar = logger.Sugar()
	return nil
}

func SyncLogger() {
	if Sugar != nil {
		_ = Sugar.Sync() // flush any buffered logs
	}
}
