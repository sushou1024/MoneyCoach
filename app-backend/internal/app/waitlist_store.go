package app

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"

	"gorm.io/gorm"
)

var errWaitlistNotFound = errors.New("waitlist entry not found")

func findWaitlistEntry(ctx context.Context, db *gorm.DB, userID, strategyID string) (WaitlistEntry, error) {
	var entry WaitlistEntry
	if db == nil {
		return entry, errWaitlistNotFound
	}
	err := db.WithContext(ctx).
		Where("user_id = ? AND strategy_id = ?", userID, strategyID).
		Order("created_at asc").
		First(&entry).Error
	if err == nil {
		return entry, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return entry, errWaitlistNotFound
	}
	return entry, err
}

func findUserWaitlistEntry(ctx context.Context, db *gorm.DB, userID string) (WaitlistEntry, error) {
	var entry WaitlistEntry
	if db == nil {
		return entry, errWaitlistNotFound
	}
	err := db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at asc").
		First(&entry).Error
	if err == nil {
		return entry, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return entry, errWaitlistNotFound
	}
	return entry, err
}

func resolveWaitlistRank(ctx context.Context, db *gorm.DB, userID string, isPaid bool) (int, error) {
	entry, err := findUserWaitlistEntry(ctx, db, userID)
	if err == nil && entry.Rank > 0 {
		return entry.Rank, nil
	}
	if err != nil && !errors.Is(err, errWaitlistNotFound) {
		return 0, err
	}
	if isPaid {
		return randomRankInRange(30, 99)
	}
	return randomRankInRange(100, 999)
}

func randomRankInRange(min, max int) (int, error) {
	if min > max {
		return 0, errors.New("invalid rank range")
	}
	span := int64(max - min + 1)
	n, err := rand.Int(rand.Reader, big.NewInt(span))
	if err != nil {
		return 0, err
	}
	return min + int(n.Int64()), nil
}
