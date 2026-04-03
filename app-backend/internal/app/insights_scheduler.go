package app

import (
	"context"
	"hash/fnv"
	"time"
)

const insightsRefreshInterval = 15 * time.Minute

type insightsRefreshCandidate struct {
	UserID            string     `gorm:"column:user_id"`
	InsightsRefreshed *time.Time `gorm:"column:insights_refreshed_at"`
}

func (s *Server) startInsightsScheduler(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.runInsightsScheduler(ctx)
			}
		}
	}()
}

func (s *Server) runInsightsScheduler(ctx context.Context) {
	if s.redis == nil {
		return
	}
	ok, err := s.redis.setNX(ctx, "lock:insights_refresh", "1", time.Minute)
	if err != nil {
		s.logger.Printf("insights refresh lock error: %v", err)
		return
	}
	if !ok {
		return
	}

	var candidates []insightsRefreshCandidate
	if err := s.db.DB().WithContext(ctx).
		Table("user_profiles").
		Select("user_profiles.user_id, user_profiles.insights_refreshed_at").
		Joins("JOIN entitlements ON entitlements.user_id = user_profiles.user_id").
		Joins("JOIN users ON users.id = user_profiles.user_id").
		Where("entitlements.status IN ?", []string{"active", "grace"}).
		Where("users.active_portfolio_snapshot_id IS NOT NULL").
		Find(&candidates).Error; err != nil {
		s.logger.Printf("insights refresh query error: %v", err)
		return
	}

	now := time.Now().UTC()
	for _, candidate := range candidates {
		if candidate.UserID == "" {
			continue
		}
		if !insightsRefreshDue(candidate.UserID, candidate.InsightsRefreshed, now) {
			continue
		}
		if err := s.queue.enqueue(ctx, jobInsightsRefresh, candidate.UserID); err != nil {
			s.logger.Printf("insights refresh enqueue error user=%s err=%v", candidate.UserID, err)
		}
	}
}

func insightsRefreshDue(userID string, lastRefreshed *time.Time, now time.Time) bool {
	if lastRefreshed != nil && now.Sub(*lastRefreshed) < insightsRefreshInterval {
		return false
	}
	offset := insightsRefreshOffset(userID)
	windowStart := now.Truncate(insightsRefreshInterval)
	target := windowStart.Add(offset)
	if now.Before(target) || now.After(target.Add(time.Minute)) {
		return false
	}
	return true
}

func insightsRefreshOffset(userID string) time.Duration {
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(userID))
	return time.Duration(hasher.Sum32()%900) * time.Second
}
