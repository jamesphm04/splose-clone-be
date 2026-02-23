package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	gormlogger "gorm.io/gorm/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/jamesphm04/splose-clone-be/internal/config"
	"github.com/jamesphm04/splose-clone-be/internal/models/entities"
)

// Connect opens a PostgreSQL connection pool using the supplied config and
// returns a configured *gorm.DB instance.
// The provided *zap.Logger is used for GORM's internal SQL logging.
func Connect(cfg config.DBConfig, appEnv string, log *zap.Logger) (*gorm.DB, error) {
	gormLog := newZapGORMLogger(log.Named("gorm"), appEnv)

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger:  gormLog,
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		return nil, fmt.Errorf("gorm.Open: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("db.DB(): %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	log.Info("database connected",
		zap.String("host", cfg.Host),
		zap.String("name", cfg.Name),
		zap.Int("maxOpenConns", cfg.MaxOpenConns),
	)
	return db, nil
}

// Migrate runs GORM auto-migration for every model.
func Migrate(db *gorm.DB, log *zap.Logger) error {
	log.Info("running auto-migration")

	err := db.AutoMigrate(
		&entities.User{},
		&entities.Patient{},
		&entities.Note{},
		&entities.Conversation{},
		&entities.Message{},
		&entities.Attachment{},
		&entities.Prompt{},
	)
	if err != nil {
		return fmt.Errorf("AutoMigrate: %w", err)
	}

	log.Info("migration completed successfully")
	return nil
}

// ---------------------------------------------------------------------------
// zapGORMLogger – adapts *zap.Logger to the gorm/logger.Interface contract.
// ---------------------------------------------------------------------------

type zapGORMLogger struct {
	log                       *zap.Logger
	slowThreshold             time.Duration
	ignoreRecordNotFoundError bool
	level                     gormlogger.LogLevel
}

func newZapGORMLogger(log *zap.Logger, env string) gormlogger.Interface {
	l := &zapGORMLogger{
		log:                       log,
		slowThreshold:             200 * time.Millisecond,
		ignoreRecordNotFoundError: true,
	}
	if env == "production" {
		l.level = gormlogger.Warn // only log slow queries + errors in production
	} else {
		l.level = gormlogger.Info // log all SQL in development
	}
	return l
}

func (z *zapGORMLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	clone := *z
	clone.level = level
	return &clone
}

func (z *zapGORMLogger) Info(_ context.Context, msg string, data ...interface{}) {
	if z.level >= gormlogger.Info {
		z.log.Sugar().Infof(msg, data...)
	}
}

func (z *zapGORMLogger) Warn(_ context.Context, msg string, data ...interface{}) {
	if z.level >= gormlogger.Warn {
		z.log.Sugar().Warnf(msg, data...)
	}
}

func (z *zapGORMLogger) Error(_ context.Context, msg string, data ...interface{}) {
	if z.level >= gormlogger.Error {
		z.log.Sugar().Errorf(msg, data...)
	}
}

func (z *zapGORMLogger) Trace(_ context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if z.level <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := []zap.Field{
		zap.Duration("elapsed", elapsed),
		zap.Int64("rows", rows),
		zap.String("sql", sql),
	}

	switch {
	case err != nil && !(z.ignoreRecordNotFoundError && errors.Is(err, gorm.ErrRecordNotFound)):
		z.log.Error("query error", append(fields, zap.Error(err))...)

	case elapsed > z.slowThreshold && z.slowThreshold > 0:
		z.log.Warn("slow query",
			append(fields, zap.Duration("threshold", z.slowThreshold))...)

	case z.level >= gormlogger.Info:
		z.log.Debug("query", fields...)
	}
}
