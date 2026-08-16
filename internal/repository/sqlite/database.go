package sqlite

import (
	"context"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct{ DB *gorm.DB }

func Open(ctx context.Context, dsn string) (*Database, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn), PrepareStmt: true})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, err
	}
	return &Database{DB: db}, nil
}

func (d *Database) Migrate(ctx context.Context) error {
	return d.DB.WithContext(ctx).AutoMigrate(&UserModel{}, &VaultModel{}, &VaultLeaseModel{}, &EntryModel{}, &AuditModel{}, &BackupModel{}, &CommentModel{}, &ActivityModel{}, &RefreshTokenModel{}, &IdempotencyModel{})
}

func (d *Database) Ready(ctx context.Context) error {
	db, err := d.DB.DB()
	if err != nil {
		return err
	}
	return db.PingContext(ctx)
}
