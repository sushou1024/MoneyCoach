package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

func (s *Server) claimExternalSubscription(ctx context.Context, provider, externalID, userID, planID string) error {
	provider = strings.TrimSpace(provider)
	externalID = strings.TrimSpace(externalID)
	userID = strings.TrimSpace(userID)
	if provider == "" || externalID == "" || userID == "" {
		return fmt.Errorf("missing external subscription identifiers")
	}
	planID = strings.TrimSpace(planID)

	var existing ExternalSubscription
	err := s.db.DB().WithContext(ctx).
		First(&existing, "provider = ? AND external_id = ?", provider, externalID).Error
	if err == nil {
		if existing.UserID != userID {
			return fmt.Errorf("%w", errSubscriptionClaimed)
		}
		if planID != "" && existing.PlanID != planID {
			if err := s.db.DB().WithContext(ctx).
				Model(&ExternalSubscription{}).
				Where("id = ?", existing.ID).
				Updates(map[string]any{"plan_id": planID, "updated_at": time.Now().UTC()}).Error; err != nil {
				return err
			}
		}
		return nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	record := ExternalSubscription{
		ID:         newID("sub"),
		Provider:   provider,
		ExternalID: externalID,
		UserID:     userID,
		PlanID:     planID,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	if err := s.db.DB().WithContext(ctx).Create(&record).Error; err != nil {
		var claimed ExternalSubscription
		if err := s.db.DB().WithContext(ctx).
			First(&claimed, "provider = ? AND external_id = ?", provider, externalID).Error; err == nil {
			if claimed.UserID != userID {
				return fmt.Errorf("%w", errSubscriptionClaimed)
			}
			return nil
		}
		return err
	}

	return nil
}
