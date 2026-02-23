package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New constructs a *zap.Logger tuned for the given environment.
//
// - "production" -> JSON encoder, Info level, no caller/stacktraces on Info
// - everything else -> conosle encoder, Debug level, caller enabled

func New(env string) (*zap.Logger, error) {
	var cfg zap.Config

	switch env {
	case "production":
		cfg = zap.NewProductionConfig()
		cfg.Level.SetLevel(zap.InfoLevel)
		cfg.DisableStacktrace = true

	default: // development
		cfg = zap.NewDevelopmentConfig()
		cfg.Level.SetLevel(zap.DebugLevel)
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.DisableCaller = false
	}

	log, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build zap logger: %w", err)
	}

	return log, nil
}

func Must(env string) *zap.Logger {
	log, err := New(env)
	if err != nil {
		panic(err)
	}
	return log
}
