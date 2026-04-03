package app

import (
	"context"
	"time"
)

const holdingsDailyLimit = 20

func (s *Server) checkHoldingsQuota(ctx context.Context, userID string) (bool, string, error) {
	if s.hasActiveEntitlement(ctx, userID) {
		return true, "", nil
	}
	profile, err := s.ensureUserProfile(ctx, userID)
	if err != nil {
		return false, "", err
	}
	nowUTC := time.Now().UTC()
	loc, err := time.LoadLocation(profile.Timezone)
	timezoneUsed := profile.Timezone
	if err != nil {
		loc = time.UTC
		timezoneUsed = "UTC"
	}

	var usage QuotaUsage
	if err := s.db.DB().WithContext(ctx).
		Where("user_id = ?", userID).
		Order("window_started_at_utc desc").
		First(&usage).Error; err == nil {
		windowStart := usage.WindowStartedUTC
		if windowStart.IsZero() {
			windowStart = usage.UsageDay
		}
		windowEnd := windowStart.Add(24 * time.Hour)
		if nowUTC.Before(windowEnd) {
			if usage.HoldingsCount >= holdingsDailyLimit {
				return false, windowEnd.Format(time.RFC3339), nil
			}
			usage.HoldingsCount++
			if err := s.db.DB().WithContext(ctx).Model(&QuotaUsage{}).Where("id = ?", usage.ID).Update("holdings_batches_count", usage.HoldingsCount).Error; err != nil {
				return false, "", err
			}
			return true, windowEnd.Format(time.RFC3339), nil
		}
	}

	nowLocal := nowUTC.In(loc)
	start := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), 0, 0, 0, 0, loc)
	windowStart := start.UTC()
	windowEnd := windowStart.Add(24 * time.Hour)

	usage = QuotaUsage{
		ID:               newID("quota"),
		UserID:           userID,
		UsageDay:         windowStart,
		TimezoneUsed:     timezoneUsed,
		WindowStartedUTC: windowStart,
		HoldingsCount:    1,
	}
	if err := s.db.DB().WithContext(ctx).Create(&usage).Error; err != nil {
		return false, "", err
	}
	return true, windowEnd.Format(time.RFC3339), nil
}

func (s *Server) hasActiveEntitlement(ctx context.Context, userID string) bool {
	var ent Entitlement
	if err := s.db.DB().WithContext(ctx).First(&ent, "user_id = ?", userID).Error; err != nil {
		return false
	}
	return ent.Status == "active" || ent.Status == "grace"
}
