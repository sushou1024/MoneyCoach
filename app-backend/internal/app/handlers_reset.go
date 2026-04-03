package app

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"

	"gorm.io/gorm"
)

func (s *Server) resetAppSecret() string {
	return strings.TrimSpace(s.cfg.ResetAppSecret)
}

func (s *Server) verifyResetSecret(w http.ResponseWriter, r *http.Request) bool {
	secret := s.resetAppSecret()
	if secret == "" {
		s.writeError(w, http.StatusNotFound, "NOT_FOUND", "endpoint disabled", nil)
		return false
	}
	provided := strings.TrimSpace(r.Header.Get("X-Reset-Secret"))
	if !secureCompare(provided, secret) {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid reset secret", nil)
		return false
	}
	return true
}

func (s *Server) handleResetAppState(w http.ResponseWriter, r *http.Request) {
	if !s.verifyResetSecret(w, r) {
		return
	}

	if err := resetDatabase(r.Context(), s.db.DB()); err != nil {
		s.logger.Printf("reset app state db error: %v", err)
		s.writeError(w, http.StatusInternalServerError, "RESET_FAILED", "failed to reset database", nil)
		return
	}
	if s.redis != nil {
		if err := s.redis.flushAll(r.Context()); err != nil {
			s.logger.Printf("reset app state redis error: %v", err)
			s.writeError(w, http.StatusInternalServerError, "RESET_FAILED", "failed to reset cache", nil)
			return
		}
	}

	s.writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type resetUserRequest struct {
	Email string `json:"email"`
}

func (s *Server) handleResetUserByEmail(w http.ResponseWriter, r *http.Request) {
	if !s.verifyResetSecret(w, r) {
		return
	}

	var req resetUserRequest
	if err := decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body", nil)
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "email is required", nil)
		return
	}

	userIDs, err := findUserIDsByEmail(r.Context(), s.db.DB(), email)
	if err != nil {
		s.logger.Printf("reset user lookup error email=%s err=%v", email, err)
		s.writeError(w, http.StatusInternalServerError, "RESET_FAILED", "failed to lookup user", nil)
		return
	}
	if len(userIDs) == 0 {
		s.writeError(w, http.StatusNotFound, "NOT_FOUND", "user not found", nil)
		return
	}

	for _, userID := range userIDs {
		if err := resetUserData(r.Context(), s.db.DB(), userID); err != nil {
			s.logger.Printf("reset user data error user=%s err=%v", userID, err)
			s.writeError(w, http.StatusInternalServerError, "RESET_FAILED", "failed to reset user data", nil)
			return
		}
	}

	s.writeJSON(w, http.StatusOK, map[string]any{
		"ok":          true,
		"user_ids":    userIDs,
		"email":       email,
		"reset_count": len(userIDs),
	})
}

func secureCompare(left, right string) bool {
	if left == "" || right == "" {
		return false
	}
	if len(left) != len(right) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}

func resetDatabase(ctx context.Context, db *gorm.DB) error {
	queries := []string{
		"DROP SCHEMA IF EXISTS public CASCADE",
		"CREATE SCHEMA public",
		"GRANT ALL ON SCHEMA public TO public",
	}
	for _, query := range queries {
		if err := db.WithContext(ctx).Exec(query).Error; err != nil {
			return fmt.Errorf("reset schema: %w", err)
		}
	}
	if err := migrateDatabase(ctx, db); err != nil {
		return err
	}
	return nil
}

func findUserIDsByEmail(ctx context.Context, db *gorm.DB, email string) ([]string, error) {
	ids := make(map[string]struct{})
	var userIDs []string

	if err := db.WithContext(ctx).
		Model(&User{}).
		Where("lower(email) = ?", email).
		Pluck("id", &userIDs).Error; err != nil {
		return nil, err
	}
	for _, id := range userIDs {
		ids[id] = struct{}{}
	}

	userIDs = nil
	if err := db.WithContext(ctx).
		Model(&AuthIdentity{}).
		Where("lower(email) = ?", email).
		Pluck("user_id", &userIDs).Error; err != nil {
		return nil, err
	}
	for _, id := range userIDs {
		ids[id] = struct{}{}
	}

	userIDs = nil
	if err := db.WithContext(ctx).
		Model(&AuthIdentity{}).
		Where("provider = ?", "email").
		Where("lower(provider_user_id) = ?", email).
		Pluck("user_id", &userIDs).Error; err != nil {
		return nil, err
	}
	for _, id := range userIDs {
		ids[id] = struct{}{}
	}

	unique := make([]string, 0, len(ids))
	for id := range ids {
		unique = append(unique, id)
	}
	return unique, nil
}

func resetUserData(ctx context.Context, db *gorm.DB, userID string) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var batchIDs []string
		if err := tx.Model(&UploadBatch{}).Where("user_id = ?", userID).Pluck("id", &batchIDs).Error; err != nil {
			return err
		}

		var imageIDs []string
		if len(batchIDs) > 0 {
			if err := tx.Model(&UploadImage{}).
				Where("upload_batch_id IN ?", batchIDs).
				Pluck("id", &imageIDs).Error; err != nil {
				return err
			}
		}

		if len(imageIDs) > 0 {
			if err := tx.Where("upload_image_id IN ?", imageIDs).Delete(&OCRAsset{}).Error; err != nil {
				return err
			}
		}

		if len(batchIDs) > 0 {
			if err := tx.Where("upload_batch_id IN ?", batchIDs).Delete(&OCRAmbiguity{}).Error; err != nil {
				return err
			}
			if err := tx.Where("upload_batch_id IN ?", batchIDs).Delete(&UploadImage{}).Error; err != nil {
				return err
			}
			if err := tx.Where("user_id = ?", userID).Delete(&UploadBatch{}).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("user_id = ?", userID).Delete(&AmbiguityResolution{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&UserAssetOverride{}).Error; err != nil {
			return err
		}

		var snapshotIDs []string
		if err := tx.Model(&PortfolioSnapshot{}).Where("user_id = ?", userID).Pluck("id", &snapshotIDs).Error; err != nil {
			return err
		}

		var calculationIDs []string
		if len(snapshotIDs) > 0 {
			if err := tx.Model(&Calculation{}).
				Where("portfolio_snapshot_id IN ?", snapshotIDs).
				Pluck("calculation_id", &calculationIDs).Error; err != nil {
				return err
			}
		}

		if len(calculationIDs) > 0 {
			if err := tx.Where("calculation_id IN ?", calculationIDs).Delete(&ReportRisk{}).Error; err != nil {
				return err
			}
			if err := tx.Where("calculation_id IN ?", calculationIDs).Delete(&ReportStrategy{}).Error; err != nil {
				return err
			}
			if err := tx.Where("calculation_id IN ?", calculationIDs).Delete(&Calculation{}).Error; err != nil {
				return err
			}
		}

		if len(snapshotIDs) > 0 {
			if err := tx.Where("portfolio_snapshot_id IN ?", snapshotIDs).Delete(&PortfolioHolding{}).Error; err != nil {
				return err
			}
			if err := tx.Where("user_id = ?", userID).Delete(&PortfolioSnapshot{}).Error; err != nil {
				return err
			}
		}

		var insightIDs []string
		if err := tx.Model(&Insight{}).Where("user_id = ?", userID).Pluck("id", &insightIDs).Error; err != nil {
			return err
		}
		if len(insightIDs) > 0 {
			if err := tx.Where("insight_id IN ?", insightIDs).Delete(&InsightEvent{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("user_id = ?", userID).Delete(&Insight{}).Error; err != nil {
			return err
		}

		if err := tx.Where("user_id = ?", userID).Delete(&PortfolioTransaction{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&PlanState{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&Entitlement{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&ExternalSubscription{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&Payment{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&WaitlistEntry{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&QuotaUsage{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&DeviceToken{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&AuthSession{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&AuthIdentity{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&UserProfile{}).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", userID).Delete(&User{}).Error; err != nil {
			return err
		}

		return nil
	})
}
