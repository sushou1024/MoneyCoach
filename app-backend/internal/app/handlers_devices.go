package app

import (
	"net/http"
	"strings"
	"time"
)

type deviceRegisterRequest struct {
	Platform       string `json:"platform"`
	PushProvider   string `json:"push_provider"`
	DeviceToken    string `json:"device_token"`
	ClientDeviceID string `json:"client_device_id"`
	Environment    string `json:"environment"`
	AppVersion     string `json:"app_version"`
	OSVersion      string `json:"os_version"`
	Locale         string `json:"locale"`
	Timezone       string `json:"timezone"`
	PushEnabled    bool   `json:"push_enabled"`
}

type deviceRegisterResponse struct {
	DeviceID   string `json:"device_id"`
	Registered bool   `json:"registered"`
}

func (s *Server) handleDeviceRegister(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}
	var req deviceRegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	if strings.TrimSpace(req.Platform) == "" || strings.TrimSpace(req.PushProvider) == "" || strings.TrimSpace(req.DeviceToken) == "" {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "platform, push_provider, device_token are required", nil)
		return
	}

	now := time.Now().UTC()
	if strings.TrimSpace(req.ClientDeviceID) != "" {
		if err := s.db.DB().WithContext(r.Context()).Model(&DeviceToken{}).
			Where("user_id = ? AND push_provider = ? AND client_device_id = ? AND device_token <> ? AND revoked_at IS NULL", userID, req.PushProvider, req.ClientDeviceID, req.DeviceToken).
			Updates(map[string]any{"revoked_at": &now, "updated_at": now}).Error; err != nil {
			s.logger.Printf("device token cleanup error user=%s err=%v", userID, err)
		}
	}
	var token DeviceToken
	result := s.db.DB().WithContext(r.Context()).Where("user_id = ? AND push_provider = ? AND device_token = ?", userID, req.PushProvider, req.DeviceToken).First(&token)
	if result.Error == nil {
		updates := map[string]any{
			"platform":         req.Platform,
			"client_device_id": nullIfEmpty(req.ClientDeviceID),
			"environment":      req.Environment,
			"app_version":      req.AppVersion,
			"os_version":       req.OSVersion,
			"locale":           req.Locale,
			"timezone":         req.Timezone,
			"push_enabled":     req.PushEnabled,
			"last_seen_at":     now,
			"revoked_at":       nil,
			"updated_at":       now,
		}
		if err := s.db.DB().WithContext(r.Context()).Model(&DeviceToken{}).Where("id = ?", token.ID).Updates(updates).Error; err != nil {
			s.writeError(w, http.StatusInternalServerError, "DEVICE_ERROR", "failed to update device", nil)
			return
		}
		s.writeJSON(w, http.StatusOK, deviceRegisterResponse{DeviceID: token.ID, Registered: true})
		return
	}

	newToken := DeviceToken{
		ID:           newID("dev"),
		UserID:       userID,
		Platform:     req.Platform,
		PushProvider: req.PushProvider,
		DeviceToken:  req.DeviceToken,
		Environment:  req.Environment,
		AppVersion:   req.AppVersion,
		OSVersion:    req.OSVersion,
		Locale:       req.Locale,
		Timezone:     req.Timezone,
		PushEnabled:  req.PushEnabled,
		LastSeenAt:   now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if req.ClientDeviceID != "" {
		newToken.ClientDeviceID = &req.ClientDeviceID
	}

	if err := s.db.DB().WithContext(r.Context()).Create(&newToken).Error; err != nil {
		s.writeError(w, http.StatusInternalServerError, "DEVICE_ERROR", "failed to register device", nil)
		return
	}

	s.writeJSON(w, http.StatusOK, deviceRegisterResponse{DeviceID: newToken.ID, Registered: true})
}

func (s *Server) handleDeviceDelete(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}
	deviceID := chiURLParam(r, "device_id")
	if deviceID == "" {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "device_id required", nil)
		return
	}

	now := time.Now().UTC()
	if err := s.db.DB().WithContext(r.Context()).Model(&DeviceToken{}).
		Where("id = ? AND user_id = ?", deviceID, userID).
		Updates(map[string]any{"revoked_at": &now, "updated_at": now}).Error; err != nil {
		s.writeError(w, http.StatusInternalServerError, "DEVICE_ERROR", "failed to revoke device", nil)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]any{"revoked": true})
}

func nullIfEmpty(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}
