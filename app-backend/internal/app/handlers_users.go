package app

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type userProfileResponse struct {
	UserID                  string              `json:"user_id"`
	Email                   *string             `json:"email"`
	TotalPaidAmount         float64             `json:"total_paid_amount"`
	Markets                 []string            `json:"markets"`
	Experience              string              `json:"experience"`
	Style                   string              `json:"style"`
	PainPoints              []string            `json:"pain_points"`
	RiskPreference          string              `json:"risk_preference"`
	RiskLevel               string              `json:"risk_level"`
	Language                string              `json:"language"`
	Timezone                string              `json:"timezone"`
	BaseCurrency            string              `json:"base_currency"`
	NotificationPrefs       map[string]bool     `json:"notification_prefs"`
	ActivePortfolioSnapshot *string             `json:"active_portfolio_snapshot_id"`
	Entitlement             entitlementResponse `json:"entitlement"`
}

type entitlementResponse struct {
	Status           string     `json:"status"`
	Provider         string     `json:"provider,omitempty"`
	PlanID           string     `json:"plan_id,omitempty"`
	CurrentPeriodEnd *time.Time `json:"current_period_end,omitempty"`
}

type userProfileUpdateRequest struct {
	Markets           *[]string        `json:"markets"`
	Experience        *string          `json:"experience"`
	Style             *string          `json:"style"`
	PainPoints        *[]string        `json:"pain_points"`
	RiskPreference    *string          `json:"risk_preference"`
	Language          *string          `json:"language"`
	Timezone          *string          `json:"timezone"`
	BaseCurrency      *string          `json:"base_currency"`
	NotificationPrefs *map[string]bool `json:"notification_prefs"`
}

type userDeleteRequest struct {
	ConfirmText string `json:"confirm_text"`
}

func (s *Server) handleUsersMe(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}

	var user User
	if err := s.db.DB().WithContext(r.Context()).First(&user, "id = ?", userID).Error; err != nil {
		s.writeError(w, http.StatusNotFound, "NOT_FOUND", "user not found", nil)
		return
	}
	profile, err := s.ensureUserProfile(r.Context(), userID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "PROFILE_ERROR", "failed to load profile", nil)
		return
	}

	entitlement := s.loadEntitlement(r.Context(), userID)

	resp := userProfileResponse{
		UserID:                  user.ID,
		Email:                   user.Email,
		TotalPaidAmount:         user.TotalPaidAmount,
		Markets:                 []string(profile.Markets),
		Experience:              profile.Experience,
		Style:                   profile.Style,
		PainPoints:              []string(profile.PainPoints),
		RiskPreference:          profile.RiskPreference,
		RiskLevel:               profile.RiskLevel,
		Language:                profile.Language,
		Timezone:                profile.Timezone,
		BaseCurrency:            profile.BaseCurrency,
		NotificationPrefs:       parseNotificationPrefs(profile.NotificationPrefs),
		ActivePortfolioSnapshot: user.ActivePortfolioSnapshot,
		Entitlement:             entitlement,
	}

	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleUsersMeUpdate(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}

	var req userProfileUpdateRequest
	if err := decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}

	profile, err := s.ensureUserProfile(r.Context(), userID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "PROFILE_ERROR", "failed to load profile", nil)
		return
	}

	updates := map[string]any{}
	languageChanged := false
	riskPreference := profile.RiskPreference
	experience := profile.Experience
	if req.Markets != nil {
		updates["markets"] = pq.StringArray(*req.Markets)
	}
	if req.Experience != nil {
		experience = strings.TrimSpace(*req.Experience)
		updates["experience"] = experience
	}
	if req.Style != nil {
		updates["style"] = strings.TrimSpace(*req.Style)
	}
	if req.PainPoints != nil {
		updates["pain_points"] = pq.StringArray(*req.PainPoints)
	}
	if req.RiskPreference != nil {
		riskPreference = strings.TrimSpace(*req.RiskPreference)
		updates["risk_preference"] = riskPreference
	}
	if req.Language != nil {
		language := strings.TrimSpace(*req.Language)
		updates["language"] = language
		if !strings.EqualFold(language, profile.Language) {
			languageChanged = true
		}
	}
	if req.Timezone != nil {
		updates["timezone"] = strings.TrimSpace(*req.Timezone)
	}
	if req.BaseCurrency != nil {
		currency := strings.ToUpper(strings.TrimSpace(*req.BaseCurrency))
		if currency != "" {
			updates["base_currency"] = currency
		}
	}
	if req.NotificationPrefs != nil {
		encoded, _ := json.Marshal(*req.NotificationPrefs)
		updates["notification_prefs"] = datatypes.JSON(encoded)
	}

	if len(updates) > 0 {
		updates["risk_level"] = deriveRiskLevel(riskPreference, experience)
		if err := s.db.DB().WithContext(r.Context()).Model(&UserProfile{}).Where("user_id = ?", userID).Updates(updates).Error; err != nil {
			s.writeError(w, http.StatusInternalServerError, "PROFILE_ERROR", "failed to update profile", nil)
			return
		}
	}

	if languageChanged {
		now := time.Now().UTC()
		if err := s.db.DB().WithContext(r.Context()).
			Model(&Insight{}).
			Where("user_id = ? AND expires_at > ?", userID, now).
			Updates(map[string]any{"expires_at": now, "status": "expired"}).Error; err != nil {
			s.writeError(w, http.StatusInternalServerError, "INSIGHT_ERROR", "failed to refresh insights", nil)
			return
		}
		if err := s.db.DB().WithContext(r.Context()).
			Model(&UserProfile{}).
			Where("user_id = ?", userID).
			Update("insights_refreshed_at", nil).Error; err != nil {
			s.writeError(w, http.StatusInternalServerError, "INSIGHT_ERROR", "failed to refresh insights", nil)
			return
		}
		if s.queue != nil {
			if err := s.queue.enqueue(r.Context(), jobInsightsRefresh, userID); err != nil {
				s.logger.Printf("insights refresh enqueue error user=%s err=%v", userID, err)
			}
		}
	}

	s.writeJSON(w, http.StatusOK, map[string]any{"updated": true})
}

func (s *Server) handleUsersMeDelete(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}

	var req userDeleteRequest
	if err := decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body", nil)
		return
	}
	if strings.TrimSpace(req.ConfirmText) != "DELETE" {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "confirm_text must be DELETE", nil)
		return
	}

	if err := resetUserData(r.Context(), s.db.DB(), userID); err != nil {
		s.logger.Printf("delete account error user=%s err=%v", userID, err)
		s.writeError(w, http.StatusInternalServerError, "DELETE_FAILED", "failed to delete account", nil)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (s *Server) ensureUserProfile(ctx context.Context, userID string) (UserProfile, error) {
	var profile UserProfile
	if err := s.db.DB().WithContext(ctx).First(&profile, "user_id = ?", userID).Error; err == nil {
		return profile, nil
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return UserProfile{}, err
	}

	if err := s.ensureUserProfileTx(ctx, s.db.DB(), userID); err != nil {
		return UserProfile{}, err
	}
	if err := s.db.DB().WithContext(ctx).First(&profile, "user_id = ?", userID).Error; err != nil {
		return UserProfile{}, err
	}
	return profile, nil
}

func (s *Server) ensureUserProfileTx(ctx context.Context, tx *gorm.DB, userID string) error {
	prefs := defaultNotificationPrefs()
	encoded, _ := json.Marshal(prefs)
	profile := UserProfile{
		UserID:            userID,
		Markets:           pq.StringArray{},
		Experience:        "",
		Style:             "",
		PainPoints:        pq.StringArray{},
		RiskPreference:    "",
		RiskLevel:         "",
		Language:          "",
		Timezone:          "",
		BaseCurrency:      "USD",
		NotificationPrefs: datatypes.JSON(encoded),
	}
	return tx.WithContext(ctx).Create(&profile).Error
}

func deriveRiskLevel(riskPreference, experience string) string {
	riskPreference = strings.ToLower(strings.TrimSpace(riskPreference))
	experience = strings.ToLower(strings.TrimSpace(experience))
	if riskPreference == "yield seeker" {
		return "conservative"
	}
	if riskPreference == "speculator" {
		switch experience {
		case "beginner":
			return "moderate"
		case "intermediate", "expert":
			return "aggressive"
		}
	}
	return "moderate"
}

func defaultNotificationPrefs() map[string]bool {
	return map[string]bool{
		"portfolio_alerts": true,
		"market_alpha":     false,
		"action_alerts":    true,
		"daily_briefing":   true,
	}
}

func parseNotificationPrefs(raw datatypes.JSON) map[string]bool {
	if len(raw) == 0 {
		return defaultNotificationPrefs()
	}
	var prefs map[string]bool
	if err := json.Unmarshal(raw, &prefs); err != nil {
		return defaultNotificationPrefs()
	}
	return prefs
}

func (s *Server) loadEntitlement(ctx context.Context, userID string) entitlementResponse {
	var ent Entitlement
	if err := s.db.DB().WithContext(ctx).First(&ent, "user_id = ?", userID).Error; err != nil {
		return entitlementResponse{Status: "expired"}
	}
	status := ent.Status
	if status == "" {
		status = "expired"
	}
	return entitlementResponse{
		Status:           status,
		Provider:         ent.Provider,
		PlanID:           ent.PlanID,
		CurrentPeriodEnd: ent.CurrentPeriodEnd,
	}
}
