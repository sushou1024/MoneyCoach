package app

import (
	"context"
	"time"
)

const briefingTargetHour = 8

type briefingCandidate struct {
	UserID   string `gorm:"column:user_id"`
	Timezone string `gorm:"column:timezone"`
}

func (s *Server) startBriefingScheduler(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.runBriefingScheduler(ctx)
			}
		}
	}()
}

func (s *Server) runBriefingScheduler(ctx context.Context) {
	if s.redis == nil {
		return
	}

	// Use a lock that expires every 10 minutes to avoid running too frequently.
	ok, err := s.redis.setNX(ctx, "lock:daily_briefing", "1", 10*time.Minute)
	if err != nil {
		s.logger.Printf("briefing scheduler lock error: %v", err)
		return
	}
	if !ok {
		return
	}

	var candidates []briefingCandidate
	if err := s.db.DB().WithContext(ctx).
		Table("user_profiles").
		Select("user_profiles.user_id, user_profiles.timezone").
		Joins("JOIN entitlements ON entitlements.user_id = user_profiles.user_id").
		Joins("JOIN users ON users.id = user_profiles.user_id").
		Where("entitlements.status IN ?", []string{"active", "grace"}).
		Where("users.active_portfolio_snapshot_id IS NOT NULL").
		Find(&candidates).Error; err != nil {
		s.logger.Printf("briefing scheduler query error: %v", err)
		return
	}

	now := time.Now().UTC()
	for _, candidate := range candidates {
		if candidate.UserID == "" {
			continue
		}
		if !briefingDue(candidate.Timezone, now) {
			continue
		}

		// Dedupe: only one briefing per user per calendar day (in their local timezone).
		loc := parseTZ(candidate.Timezone)
		localDay := now.In(loc).Format("20060102")
		dedupeKey := "briefing:scheduled:" + candidate.UserID + ":" + localDay
		set, err := s.redis.setNX(ctx, dedupeKey, "1", 24*time.Hour)
		if err != nil {
			s.logger.Printf("briefing dedupe error user=%s err=%v", candidate.UserID, err)
			continue
		}
		if !set {
			continue
		}

		if err := s.queue.enqueue(ctx, jobDailyBriefing, candidate.UserID); err != nil {
			s.logger.Printf("briefing enqueue error user=%s err=%v", candidate.UserID, err)
		}
	}
}

// briefingDue returns true when the current UTC time falls within the 8:00
// hour (inclusive) of the user's local timezone.
func briefingDue(tz string, nowUTC time.Time) bool {
	loc := parseTZ(tz)
	local := nowUTC.In(loc)
	return local.Hour() == briefingTargetHour
}

// parseTZ loads a timezone; falls back to Asia/Shanghai (UTC+8) for empty or
// invalid values.
func parseTZ(tz string) *time.Location {
	if tz == "" {
		tz = "Asia/Shanghai"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc, _ = time.LoadLocation("Asia/Shanghai")
	}
	return loc
}
