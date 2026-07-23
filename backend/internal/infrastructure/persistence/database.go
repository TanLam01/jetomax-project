package persistence

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	ORM *gorm.DB
}

func Open(ctx context.Context, databaseURL string) (*Database, error) {
	databaseLogger := logger.New(log.New(os.Stdout, "", log.LstdFlags), logger.Config{
		SlowThreshold:             500 * time.Millisecond,
		LogLevel:                  logger.Warn,
		IgnoreRecordNotFoundError: true,
		ParameterizedQueries:      true,
		Colorful:                  false,
	})
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		TranslateError: true,
		Logger:         databaseLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("open PostgreSQL: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get PostgreSQL connection pool: %w", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping PostgreSQL: %w", err)
	}

	return &Database{ORM: db}, nil
}

func (d *Database) Ping(ctx context.Context) error {
	db, err := d.ORM.DB()
	if err != nil {
		return err
	}
	return db.PingContext(ctx)
}

func (d *Database) Close() error {
	db, err := d.ORM.DB()
	if err != nil {
		return err
	}
	return db.Close()
}
