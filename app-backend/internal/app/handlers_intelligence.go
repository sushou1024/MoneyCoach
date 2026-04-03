package app

import (
	"net/http"
	"strings"
)

func (s *Server) handleIntelligenceRegime(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}
	if !s.hasActiveEntitlement(r.Context(), userID) {
		s.writeError(w, http.StatusForbidden, "ENTITLEMENT_REQUIRED", "active subscription required", nil)
		return
	}

	response, err := s.buildMarketRegime(r.Context(), userID)
	if err != nil {
		status := http.StatusInternalServerError
		code := "INTELLIGENCE_ERROR"
		message := "failed to load market regime"
		if strings.Contains(err.Error(), "active portfolio not found") {
			status = http.StatusNotFound
			code = "NOT_FOUND"
			message = "active portfolio not found"
		}
		s.writeError(w, status, code, message, nil)
		return
	}
	if response == nil {
		s.writeError(w, http.StatusNotFound, "NOT_FOUND", "market regime unavailable", nil)
		return
	}

	s.writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleIntelligenceAssetBrief(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}
	if !s.hasActiveEntitlement(r.Context(), userID) {
		s.writeError(w, http.StatusForbidden, "ENTITLEMENT_REQUIRED", "active subscription required", nil)
		return
	}
	assetKey := strings.TrimSpace(chiURLParam(r, "asset_key"))
	if assetKey == "" {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "asset_key required", nil)
		return
	}

	response, err := s.buildAssetBrief(r.Context(), userID, assetKey)
	if err != nil {
		status := http.StatusInternalServerError
		code := "INTELLIGENCE_ERROR"
		message := "failed to load asset brief"
		switch {
		case strings.Contains(err.Error(), "asset not found"):
			status = http.StatusNotFound
			code = "NOT_FOUND"
			message = "asset brief not found"
		case strings.Contains(err.Error(), "active portfolio not found"):
			status = http.StatusNotFound
			code = "NOT_FOUND"
			message = "active portfolio not found"
		}
		s.writeError(w, status, code, message, nil)
		return
	}
	if response == nil {
		s.writeError(w, http.StatusNotFound, "NOT_FOUND", "asset brief not found", nil)
		return
	}

	s.writeJSON(w, http.StatusOK, response)
}
