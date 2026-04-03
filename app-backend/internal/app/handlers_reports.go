package app

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
)

func (s *Server) handleReportPreview(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}
	calcID := chiURLParam(r, "calculation_id")
	var calc Calculation
	if err := s.db.DB().WithContext(r.Context()).First(&calc, "calculation_id = ?", calcID).Error; err != nil {
		s.writeError(w, http.StatusNotFound, "NOT_FOUND", "report not found", nil)
		return
	}
	if !s.calculationOwnedBy(r.Context(), calc, userID) {
		s.writeError(w, http.StatusForbidden, "FORBIDDEN", "report not accessible", nil)
		return
	}
	if calc.StatusPreview != "ready" {
		s.writeError(w, http.StatusAccepted, "NOT_READY", "preview not ready", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(calc.PreviewPayload)
}

func (s *Server) handleReportPaid(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}
	calcID := chiURLParam(r, "calculation_id")
	var calc Calculation
	if err := s.db.DB().WithContext(r.Context()).First(&calc, "calculation_id = ?", calcID).Error; err != nil {
		s.writeError(w, http.StatusNotFound, "NOT_FOUND", "report not found", nil)
		return
	}
	if !s.calculationOwnedBy(r.Context(), calc, userID) {
		s.writeError(w, http.StatusForbidden, "FORBIDDEN", "report not accessible", nil)
		return
	}
	if !s.hasActiveEntitlement(r.Context(), userID) {
		s.writeError(w, http.StatusForbidden, "ENTITLEMENT_REQUIRED", "active subscription required", nil)
		return
	}
	switch strings.ToLower(strings.TrimSpace(calc.StatusPaid)) {
	case "ready":
		s.writeJSON(w, http.StatusOK, map[string]any{"calculation_id": calcID, "status": "ready"})
		return
	case "processing":
		s.writeJSON(w, http.StatusOK, map[string]any{"calculation_id": calcID, "status": "processing"})
		return
	}
	if err := s.db.DB().WithContext(r.Context()).
		Model(&Calculation{}).
		Where("calculation_id = ?", calcID).
		Update("status_paid", "processing").Error; err != nil {
		s.writeError(w, http.StatusInternalServerError, "REPORT_ERROR", "failed to queue paid report", nil)
		return
	}
	_ = s.queue.enqueue(r.Context(), jobPaidReport, calcID)
	s.writeJSON(w, http.StatusOK, map[string]any{"calculation_id": calcID, "status": "processing"})
}

type reportActiveRequest struct {
	Tier string `json:"tier"`
}

type reportActiveResponse struct {
	CalculationID string `json:"calculation_id"`
	Status        string `json:"status"`
}

func (s *Server) handleReportActive(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}

	var req reportActiveRequest
	if err := decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	tier := strings.ToLower(strings.TrimSpace(req.Tier))
	if tier != "preview" && tier != "paid" {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "tier must be preview or paid", nil)
		return
	}
	if tier == "paid" && !s.hasActiveEntitlement(r.Context(), userID) {
		s.writeError(w, http.StatusForbidden, "ENTITLEMENT_REQUIRED", "active subscription required", nil)
		return
	}

	var user User
	if err := s.db.DB().WithContext(r.Context()).First(&user, "id = ?", userID).Error; err != nil || user.ActivePortfolioSnapshot == nil {
		s.writeError(w, http.StatusNotFound, "NOT_FOUND", "active portfolio not found", nil)
		return
	}
	snapshotID := *user.ActivePortfolioSnapshot

	var calc Calculation
	err := s.db.DB().WithContext(r.Context()).
		Where("portfolio_snapshot_id = ?", snapshotID).
		Order("created_at desc").
		First(&calc).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			s.writeError(w, http.StatusInternalServerError, "REPORT_ERROR", "failed to load report", nil)
			return
		}
		now := time.Now().UTC()
		calc = Calculation{
			ID:                  newID("calc"),
			PortfolioSnapshotID: snapshotID,
			StatusPreview:       "not_started",
			StatusPaid:          "not_started",
			ModelVersionPreview: geminiModel,
			PromptHashPreview:   hashPrompt(s.prompts.PreviewReport),
			CreatedAt:           now,
		}
		if tier == "preview" {
			calc.StatusPreview = "processing"
		} else {
			calc.StatusPaid = "processing"
		}
		if err := s.db.DB().WithContext(r.Context()).Create(&calc).Error; err != nil {
			s.writeError(w, http.StatusInternalServerError, "REPORT_ERROR", "failed to create report", nil)
			return
		}
	} else {
		if tier == "preview" {
			status := strings.ToLower(strings.TrimSpace(calc.StatusPreview))
			if status == "processing" || status == "ready" {
				s.writeJSON(w, http.StatusOK, reportActiveResponse{CalculationID: calc.ID, Status: status})
				return
			}
			if err := s.db.DB().WithContext(r.Context()).
				Model(&Calculation{}).
				Where("calculation_id = ?", calc.ID).
				Updates(map[string]any{
					"status_preview":        "processing",
					"preview_payload":       nil,
					"model_version_preview": geminiModel,
					"prompt_hash_preview":   hashPrompt(s.prompts.PreviewReport),
				}).Error; err != nil {
				s.writeError(w, http.StatusInternalServerError, "REPORT_ERROR", "failed to queue preview report", nil)
				return
			}
			calc.StatusPreview = "processing"
		} else {
			status := strings.ToLower(strings.TrimSpace(calc.StatusPaid))
			if status == "processing" || status == "ready" {
				s.writeJSON(w, http.StatusOK, reportActiveResponse{CalculationID: calc.ID, Status: status})
				return
			}
			if err := s.db.DB().WithContext(r.Context()).
				Model(&Calculation{}).
				Where("calculation_id = ?", calc.ID).
				Updates(map[string]any{
					"status_paid":  "processing",
					"paid_payload": nil,
					"paid_at":      nil,
				}).Error; err != nil {
				s.writeError(w, http.StatusInternalServerError, "REPORT_ERROR", "failed to queue paid report", nil)
				return
			}
			calc.StatusPaid = "processing"
		}
	}

	if tier == "preview" {
		_ = s.queue.enqueue(r.Context(), jobPreviewReport, calc.ID)
		s.writeJSON(w, http.StatusOK, reportActiveResponse{CalculationID: calc.ID, Status: "processing"})
		return
	}
	_ = s.queue.enqueue(r.Context(), jobPaidReport, calc.ID)
	s.writeJSON(w, http.StatusOK, reportActiveResponse{CalculationID: calc.ID, Status: "processing"})
}

func (s *Server) handleReportByID(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}
	calcID := chiURLParam(r, "calculation_id")
	var calc Calculation
	if err := s.db.DB().WithContext(r.Context()).First(&calc, "calculation_id = ?", calcID).Error; err != nil {
		s.writeError(w, http.StatusNotFound, "NOT_FOUND", "report not found", nil)
		return
	}
	if !s.calculationOwnedBy(r.Context(), calc, userID) {
		s.writeError(w, http.StatusForbidden, "FORBIDDEN", "report not accessible", nil)
		return
	}
	if calc.StatusPaid == "ready" {
		w.Header().Set("Content-Type", "application/json")
		w.Write(calc.PaidPayload)
		return
	}
	if calc.StatusPreview == "ready" {
		w.Header().Set("Content-Type", "application/json")
		w.Write(calc.PreviewPayload)
		return
	}
	s.writeError(w, http.StatusAccepted, "NOT_READY", "report not ready", nil)
}

func (s *Server) handleReportList(w http.ResponseWriter, r *http.Request) {
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
	limit := parseLimit(r, 20, 100)
	cursor := parseCursor(r)

	query := s.db.DB().WithContext(r.Context()).Joins("JOIN portfolio_snapshots ON portfolio_snapshots.id = calculations.portfolio_snapshot_id").Where("portfolio_snapshots.user_id = ?", userID).Order("calculations.created_at desc").Limit(limit + 1)
	if cursor != "" {
		query = query.Where("calculations.calculation_id < ?", cursor)
	}

	var calcs []Calculation
	if err := query.Find(&calcs).Error; err != nil {
		s.writeError(w, http.StatusInternalServerError, "REPORT_ERROR", "failed to list reports", nil)
		return
	}

	items := make([]map[string]any, 0, len(calcs))
	nextCursor := ""
	for i, calc := range calcs {
		if i == limit {
			nextCursor = calc.ID
			break
		}
		reportTier := "preview"
		status := calc.StatusPreview
		paidStatus := strings.ToLower(strings.TrimSpace(calc.StatusPaid))
		if paidStatus != "" && paidStatus != "not_started" {
			reportTier = "paid"
			status = calc.StatusPaid
		}
		items = append(items, map[string]any{
			"calculation_id":        calc.ID,
			"report_tier":           reportTier,
			"created_at":            calc.CreatedAt.Format(time.RFC3339),
			"health_score":          calc.HealthScore,
			"status":                status,
			"portfolio_snapshot_id": calc.PortfolioSnapshotID,
			"is_active":             user.ActivePortfolioSnapshot != nil && calc.PortfolioSnapshotID == *user.ActivePortfolioSnapshot,
		})
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"items": items, "next_cursor": nextCursor})
}

func (s *Server) handleReportPlan(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}
	calcID := chiURLParam(r, "calculation_id")
	planID := chiURLParam(r, "plan_id")
	var calc Calculation
	if err := s.db.DB().WithContext(r.Context()).First(&calc, "calculation_id = ?", calcID).Error; err != nil {
		s.writeError(w, http.StatusNotFound, "NOT_FOUND", "report not found", nil)
		return
	}
	if !s.calculationOwnedBy(r.Context(), calc, userID) {
		s.writeError(w, http.StatusForbidden, "FORBIDDEN", "report not accessible", nil)
		return
	}
	if calc.StatusPaid != "ready" {
		s.writeError(w, http.StatusNotFound, "NOT_FOUND", "paid report not available", nil)
		return
	}
	var plan ReportStrategy
	if err := s.db.DB().WithContext(r.Context()).Where("calculation_id = ? AND plan_id = ?", calcID, planID).First(&plan).Error; err != nil {
		s.writeError(w, http.StatusNotFound, "NOT_FOUND", "plan not found", nil)
		return
	}
	riskSeverity := ""
	if plan.LinkedRiskID != "" {
		var risk ReportRisk
		if err := s.db.DB().WithContext(r.Context()).
			Where("calculation_id = ? AND risk_id = ?", calcID, plan.LinkedRiskID).
			First(&risk).Error; err == nil {
			riskSeverity = risk.Severity
		}
	}
	var snapshot PortfolioSnapshot
	if err := s.db.DB().WithContext(r.Context()).First(&snapshot, "id = ?", calc.PortfolioSnapshotID).Error; err != nil {
		s.writeError(w, http.StatusNotFound, "NOT_FOUND", "portfolio snapshot not found", nil)
		return
	}
	var parameters map[string]any
	if len(plan.Parameters) > 0 {
		_ = json.Unmarshal(plan.Parameters, &parameters)
	}

	quoteCurrency := "USD"
	rateFromUSD := 1.0
	if plan.AssetType == "portfolio" || strings.HasPrefix(plan.AssetKey, "portfolio:") {
		quoteCurrency = normalizeCurrency(snapshot.BaseCurrency)
		if quoteCurrency == "" {
			quoteCurrency = "USD"
		}
		if snapshot.BaseFXRateToUSD != nil && *snapshot.BaseFXRateToUSD > 0 {
			rateFromUSD = 1 / *snapshot.BaseFXRateToUSD
		}
	} else if _, holdings, err := s.loadSnapshotWithHoldings(r.Context(), snapshot.ID); err == nil {
		for _, holding := range holdings {
			if holding.AssetKey != plan.AssetKey {
				continue
			}
			quoteCurrency = quoteCurrencyForHolding(holding)
			if fxRateToUSD := quoteFXRateToUSD(holding); fxRateToUSD > 0 {
				rateFromUSD = 1 / fxRateToUSD
			}
			break
		}
	}
	if len(parameters) > 0 && rateFromUSD != 1 {
		displayPlans := convertLockedPlansToDisplay([]lockedPlan{{
			PlanID:       plan.PlanID,
			StrategyID:   plan.StrategyID,
			AssetType:    plan.AssetType,
			Symbol:       plan.Symbol,
			AssetKey:     plan.AssetKey,
			LinkedRiskID: plan.LinkedRiskID,
			Parameters:   parameters,
		}}, rateFromUSD)
		if len(displayPlans) == 1 {
			parameters = displayPlans[0].Parameters
		}
	}
	chartSeries := s.buildPlanChartSeries(r.Context(), snapshot, plan)

	s.writeJSON(w, http.StatusOK, map[string]any{
		"plan_id":          plan.PlanID,
		"strategy_id":      plan.StrategyID,
		"asset_type":       plan.AssetType,
		"symbol":           plan.Symbol,
		"asset_key":        plan.AssetKey,
		"quote_currency":   quoteCurrency,
		"linked_risk_id":   plan.LinkedRiskID,
		"priority":         planPriorityFromSeverity(riskSeverity),
		"parameters":       parameters,
		"rationale":        plan.Rationale,
		"expected_outcome": plan.ExpectedOutcome,
		"chart_series":     chartSeries,
	})
}

func (s *Server) calculationOwnedBy(ctx context.Context, calc Calculation, userID string) bool {
	var snapshot PortfolioSnapshot
	if err := s.db.DB().WithContext(ctx).First(&snapshot, "id = ?", calc.PortfolioSnapshotID).Error; err != nil {
		return false
	}
	return snapshot.UserID == userID
}
