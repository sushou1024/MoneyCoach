package app

import (
	"errors"
	"net/http"
	"time"
)

type waitlistRequest struct {
	StrategyID    string `json:"strategy_id"`
	CalculationID string `json:"calculation_id"`
}

func (s *Server) handleWaitlist(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}
	var req waitlistRequest
	if err := decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	if req.StrategyID == "" {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "strategy_id required", nil)
		return
	}

	ctx := r.Context()
	existing, err := findWaitlistEntry(ctx, s.db.DB(), userID, req.StrategyID)
	if err == nil {
		s.writeJSON(w, http.StatusOK, map[string]any{"rank": existing.Rank})
		return
	}
	if err != nil && !errors.Is(err, errWaitlistNotFound) {
		s.writeError(w, http.StatusInternalServerError, "WAITLIST_ERROR", "failed to load waitlist entry", nil)
		return
	}

	rank, err := resolveWaitlistRank(ctx, s.db.DB(), userID, s.hasActiveEntitlement(ctx, userID))
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "WAITLIST_ERROR", "failed to generate waitlist rank", nil)
		return
	}

	entry := WaitlistEntry{
		ID:         newID("wait"),
		UserID:     userID,
		StrategyID: req.StrategyID,
		Rank:       rank,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	if err := s.db.DB().WithContext(r.Context()).Create(&entry).Error; err != nil {
		s.writeError(w, http.StatusInternalServerError, "WAITLIST_ERROR", "failed to create waitlist entry", nil)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]any{"rank": rank})
}
