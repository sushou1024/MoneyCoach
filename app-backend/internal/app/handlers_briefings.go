package app

import (
	"net/http"
	"time"
)

type briefingView struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Priority  int    `json:"priority"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	PushText  string `json:"push_text"`
	CreatedAt string `json:"created_at"`
}

func (s *Server) handleBriefingsToday(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}

	profile, err := s.ensureUserProfile(r.Context(), userID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "PROFILE_ERROR", "failed to load profile", nil)
		return
	}

	loc := parseTZ(profile.Timezone)
	today := time.Now().In(loc).Format("2006-01-02")

	var rows []Briefing
	if err := s.db.DB().WithContext(r.Context()).
		Where("user_id = ? AND briefing_date = ?", userID, today).
		Order("priority ASC").
		Find(&rows).Error; err != nil {
		s.writeError(w, http.StatusInternalServerError, "BRIEFING_ERROR", "failed to load briefings", nil)
		return
	}

	views := make([]briefingView, 0, len(rows))
	for _, row := range rows {
		views = append(views, briefingView{
			ID:        row.ID,
			Type:      row.Type,
			Priority:  row.Priority,
			Title:     row.Title,
			Body:      row.Body,
			PushText:  row.PushText,
			CreatedAt: row.CreatedAt.Format(time.RFC3339),
		})
	}

	s.writeJSON(w, http.StatusOK, map[string]any{
		"briefings": views,
		"date":      today,
	})
}
