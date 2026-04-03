package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func (s *Server) handleInsightsList(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}
	if !s.hasActiveEntitlement(r.Context(), userID) {
		s.writeError(w, http.StatusForbidden, "ENTITLEMENT_REQUIRED", "active subscription required", nil)
		return
	}

	filter := strings.TrimSpace(r.URL.Query().Get("filter"))
	if filter == "" {
		filter = "all"
	}
	if filter != "all" && filter != insightTypePortfolioWatch && filter != insightTypeMarketAlpha && filter != insightTypeActionAlert {
		filter = "all"
	}

	limit := parseLimit(r, 20, 100)
	cursor := parseCursor(r)
	now := time.Now().UTC()

	query := s.db.DB().WithContext(r.Context()).Where("user_id = ? AND status = ? AND expires_at > ?", userID, "active", now)
	if filter != "all" {
		query = query.Where("type = ?", filter)
	}
	var rows []Insight
	if err := query.Find(&rows).Error; err != nil {
		s.writeError(w, http.StatusInternalServerError, "INSIGHT_ERROR", "failed to load insights", nil)
		return
	}

	var (
		baseCurrency string
		rateFromUSD  float64
		displayReady bool
	)
	resolveDisplay := func() {
		if displayReady {
			return
		}
		displayReady = true
		profile, err := s.ensureUserProfile(r.Context(), userID)
		if err != nil {
			baseCurrency = "USD"
			rateFromUSD = 1
			return
		}
		baseCurrency = normalizeCurrency(profile.BaseCurrency)
		if baseCurrency == "USD" {
			rateFromUSD = 1
			return
		}
		oerResp, err := s.market.openExchangeLatest(r.Context())
		if err != nil {
			baseCurrency = "USD"
			rateFromUSD = 1
			return
		}
		baseCurrency, rateFromUSD = resolveDisplayCurrency(baseCurrency, oerResp.Rates)
	}

	items := make([]insightItem, 0, len(rows))
	for _, row := range rows {
		item := insightItem{
			ID:            row.ID,
			Type:          row.Type,
			Asset:         row.Asset,
			AssetKey:      derefString(row.AssetKey),
			Timeframe:     derefString(row.Timeframe),
			Severity:      row.Severity,
			TriggerReason: row.TriggerReason,
			TriggerKey:    row.TriggerKey,
			StrategyID:    derefString(row.StrategyID),
			PlanID:        derefString(row.PlanID),
			CreatedAt:     row.CreatedAt,
			ExpiresAt:     row.ExpiresAt,
		}
		if row.SuggestedAction != nil {
			item.SuggestedAction = *row.SuggestedAction
		}
		if row.BetaToPortfolio != nil {
			item.BetaToPortfolio = *row.BetaToPortfolio
		}
		if len(row.SuggestedQuantity) > 0 {
			var quantity suggestedQuantity
			if err := json.Unmarshal(row.SuggestedQuantity, &quantity); err == nil {
				if strings.EqualFold(quantity.Mode, "usd") && quantity.AmountUSD != nil && quantity.AmountDisplay == nil {
					resolveDisplay()
					amountDisplay, displayCurrency := displayAmountFromUSD(*quantity.AmountUSD, baseCurrency, rateFromUSD)
					quantity.AmountDisplay = &amountDisplay
					if strings.TrimSpace(quantity.DisplayCurrency) == "" {
						quantity.DisplayCurrency = displayCurrency
					}
				}
				item.SuggestedQuantity = &quantity
			}
		}
		if len(row.CTAPayload) > 0 {
			var payload map[string]any
			if err := json.Unmarshal(row.CTAPayload, &payload); err == nil {
				item.CTAPayload = payload
			}
		}
		items = append(items, item)
	}

	sortInsights(items)
	start := 0
	if cursor != "" {
		for i, item := range items {
			if item.ID == cursor {
				start = i + 1
				break
			}
		}
	}

	end := start + limit
	if end > len(items) {
		end = len(items)
	}
	page := items[start:end]
	nextCursor := ""
	if end < len(items) {
		nextCursor = items[end-1].ID
	}

	responseItems := make([]map[string]any, 0, len(page))
	for _, item := range page {
		responseItems = append(responseItems, map[string]any{
			"id":                 item.ID,
			"type":               item.Type,
			"asset":              item.Asset,
			"asset_key":          item.AssetKey,
			"timeframe":          item.Timeframe,
			"severity":           item.Severity,
			"trigger_reason":     item.TriggerReason,
			"trigger_key":        item.TriggerKey,
			"strategy_id":        item.StrategyID,
			"plan_id":            item.PlanID,
			"suggested_action":   item.SuggestedAction,
			"suggested_quantity": item.SuggestedQuantity,
			"cta_payload":        item.CTAPayload,
			"created_at":         item.CreatedAt.Format(time.RFC3339),
			"expires_at":         item.ExpiresAt.Format(time.RFC3339),
		})
	}

	s.writeJSON(w, http.StatusOK, map[string]any{"items": responseItems, "next_cursor": nextCursor})
}

func (s *Server) handleInsightExecute(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}
	if !s.hasActiveEntitlement(r.Context(), userID) {
		s.writeError(w, http.StatusForbidden, "ENTITLEMENT_REQUIRED", "active subscription required", nil)
		return
	}
	insightID := chiURLParam(r, "insight_id")
	if insightID == "" {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "insight_id required", nil)
		return
	}

	var req insightExecuteRequest
	if err := decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	method := normalizeInsightMethod(req.Method)
	if method == "" {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "method must be suggested, manual, or trade_slip", nil)
		return
	}

	var insight Insight
	if err := s.db.DB().WithContext(r.Context()).First(&insight, "id = ? AND user_id = ?", insightID, userID).Error; err != nil {
		s.writeError(w, http.StatusNotFound, "INSIGHT_NOT_FOUND", "insight not found", nil)
		return
	}
	if insight.Status != "active" {
		s.writeError(w, http.StatusConflict, "INSIGHT_NOT_ACTIVE", "insight already handled", nil)
		return
	}
	if insight.Type == insightTypeMarketAlpha {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "market alpha insights cannot be executed", nil)
		return
	}

	suggested := parseSuggestedQuantity(insight.SuggestedQuantity)

	if method == "trade_slip" {
		if len(req.TransactionIDs) == 0 {
			s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "transaction_ids required for trade_slip", nil)
			return
		}
		txIDs, err := s.validateInsightTransactions(r.Context(), userID, req.TransactionIDs)
		if err != nil {
			s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
			return
		}
		meta := map[string]any{
			"method":          method,
			"transaction_ids": txIDs,
		}
		if err := s.updateInsightStatus(r.Context(), insightID, userID, "executed", "executed", meta); err != nil {
			s.writeError(w, http.StatusInternalServerError, "INSIGHT_ERROR", "failed to execute insight", nil)
			return
		}
		s.writeJSON(w, http.StatusOK, map[string]any{"executed": true, "transaction_ids": txIDs})
		return
	}

	plan, err := s.loadPlanForInsight(r.Context(), userID, insight)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	assetTypes, seriesByAssetKey, err := s.loadInsightPriceSeries(r.Context(), userID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "INSIGHT_ERROR", "failed to load pricing data", nil)
		return
	}

	deltas, err := s.buildInsightDeltas(r.Context(), insight, plan, suggested, req, assetTypes, seriesByAssetKey)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}

	newSnapshotID, txIDs, warnings, err := s.applyDeltaToActive(r.Context(), userID, "", deltas)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "INSIGHT_ERROR", "failed to apply insight update", nil)
		return
	}
	meta := map[string]any{
		"method":                method,
		"quantity":              req.Quantity,
		"quantity_unit":         req.QuantityUnit,
		"transaction_ids":       txIDs,
		"portfolio_snapshot_id": newSnapshotID,
	}
	if suggested != nil {
		meta["suggested_quantity"] = suggested
	}
	if len(warnings) > 0 {
		meta["warnings"] = warnings
	}
	if err := s.updateInsightStatus(r.Context(), insightID, userID, "executed", "executed", meta); err != nil {
		s.writeError(w, http.StatusInternalServerError, "INSIGHT_ERROR", "failed to execute insight", nil)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]any{
		"executed":              true,
		"transaction_ids":       txIDs,
		"portfolio_snapshot_id": newSnapshotID,
		"warnings":              warnings,
	})
}

func (s *Server) handleInsightDismiss(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}
	if !s.hasActiveEntitlement(r.Context(), userID) {
		s.writeError(w, http.StatusForbidden, "ENTITLEMENT_REQUIRED", "active subscription required", nil)
		return
	}
	insightID := chiURLParam(r, "insight_id")
	if insightID == "" {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "insight_id required", nil)
		return
	}
	var req insightDismissRequest
	if err := decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	if !isValidInsightDismissReason(req.Reason) {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "reason must be not_relevant, later, or other", nil)
		return
	}

	var insight Insight
	if err := s.db.DB().WithContext(r.Context()).First(&insight, "id = ? AND user_id = ?", insightID, userID).Error; err != nil {
		s.writeError(w, http.StatusNotFound, "INSIGHT_NOT_FOUND", "insight not found", nil)
		return
	}
	if insight.Status != "active" {
		s.writeError(w, http.StatusConflict, "INSIGHT_NOT_ACTIVE", "insight already handled", nil)
		return
	}

	meta := map[string]any{"reason": req.Reason}
	if err := s.updateInsightStatus(r.Context(), insightID, userID, "dismissed", "dismissed", meta); err != nil {
		s.writeError(w, http.StatusInternalServerError, "INSIGHT_ERROR", "failed to dismiss insight", nil)
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"dismissed": true})
}
