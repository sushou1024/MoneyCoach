package app

import (
	"context"

	"gorm.io/gorm"
)

type dbStore struct {
	db *gorm.DB
}

func (s *dbStore) withTx(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return s.db.WithContext(ctx).Transaction(fn)
}

func (s *dbStore) DB() *gorm.DB {
	return s.db
}
