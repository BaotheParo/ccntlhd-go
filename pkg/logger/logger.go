package logger

import (
	"go.uber.org/zap"
)

var Log *zap.Logger

func InitLogger(env string) error {
	var err error
	if env == "dev" {
		Log, err = zap.NewDevelopment()
	} else {
		Log, err = zap.NewProduction() // Default to production for safety
	}

	if err != nil {
		return err
	}
	
	// Replace global logger if you want to use zap.S() or zap.L()
	zap.ReplaceGlobals(Log)
	
	return nil
}
