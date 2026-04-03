package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type insightExecuteRequest struct {
	Method         string   `json:"method"`
	Quantity       float64  `json:"quantity"`
	QuantityUnit   string   `json:"quantity_unit"`
	TransactionIDs []string `json:"transaction_ids"`
}

type insightDismissRequest struct {
	Reason string `json:"reason"`
}

func normalizeInsightMethod(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "suggested"
	}
	switch value {
	case "suggested", "manual", "trade_slip":
		return value
	default:
		return ""
	}
}

func isValidInsightDismissReason(value string) bool {
	switch strings.TrimSpace(value) {
	case "not_relevant", "later", "other":
		return true
	default:
		return false
	}
}

func parseSuggestedQuantity(raw datatypes.JSON) *suggestedQuantity {
	if len(raw) == 0 {
		return nil
	}
	var qty suggestedQuantity
	if err := json.Unmarshal(raw, &qty); err != nil {
		return nil
	}
	if strings.TrimSpace(qty.Mode) == "" {
		return nil
	}
	return &qty
}

func (s *Server) loadPlanForInsight(ctx context.Context, userID string, insight Insight) (lockedPlan, error) {
	planID := strings.TrimSpace(derefString(insight.PlanID))
	if planID == "" {
		return lockedPlan{}, fmt.Errorf("plan_id required for insight execution")
	}
	plans, err := s.loadLatestPaidPlansForUser(ctx, userID)
	if err != nil {
		return lockedPlan{}, err
	}
	for _, plan := range plans {
		if plan.PlanID != planID {
			continue
		}
		strategyID := strings.TrimSpace(derefString(insight.StrategyID))
		if strategyID != "" && strategyID != plan.StrategyID {
			return lockedPlan{}, fmt.Errorf("insight strategy mismatch")
		}
		return plan, nil
	}
	return lockedPlan{}, fmt.Errorf("plan not found for insight")
}

func (s *Server) loadInsightPriceSeries(ctx context.Context, userID string) (map[string]string, map[string][]ohlcPoint, error) {
	var user User
	if err := s.db.DB().WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return nil, nil, err
	}
	if user.ActivePortfolioSnapshot == nil {
		return nil, nil, errNotFound
	}
	_, holdings, err := s.loadSnapshotWithHoldings(ctx, *user.ActivePortfolioSnapshot)
	if err != nil {
		return nil, nil, err
	}

	assetTypes := make(map[string]string, len(holdings))
	for _, holding := range holdings {
		if holding.AssetKey == "" {
			continue
		}
		assetTypes[holding.AssetKey] = holding.AssetType
	}

	seriesByAssetKey := fetchPriceSeries(ctx, s.market, holdings)
	return assetTypes, seriesByAssetKey, nil
}

func (s *Server) buildInsightDeltas(ctx context.Context, insight Insight, plan lockedPlan, suggested *suggestedQuantity, req insightExecuteRequest, assetTypes map[string]string, seriesByAssetKey map[string][]ohlcPoint) ([]transactionDelta, error) {
	method := normalizeInsightMethod(req.Method)
	switch method {
	case "suggested":
		return s.buildInsightDeltasFromSuggested(ctx, insight, plan, suggested, assetTypes, seriesByAssetKey)
	case "manual":
		return s.buildInsightDeltasFromManual(ctx, insight, plan, suggested, req, assetTypes, seriesByAssetKey)
	default:
		return nil, fmt.Errorf("unsupported insight execution method")
	}
}

func (s *Server) buildInsightDeltasFromSuggested(ctx context.Context, insight Insight, plan lockedPlan, suggested *suggestedQuantity, assetTypes map[string]string, seriesByAssetKey map[string][]ohlcPoint) ([]transactionDelta, error) {
	if suggested == nil {
		return nil, fmt.Errorf("suggested quantity is not available")
	}
	mode := strings.ToLower(strings.TrimSpace(suggested.Mode))
	switch mode {
	case "usd":
		if suggested.AmountUSD == nil || *suggested.AmountUSD <= 0 {
			return nil, fmt.Errorf("suggested usd amount missing")
		}
		sign, err := insightTradeSign(plan, insight.Type)
		if err != nil {
			return nil, err
		}
		delta, err := s.buildInsightDelta(ctx, insight, plan, *suggested.AmountUSD, "usd", sign, assetTypes, seriesByAssetKey)
		if err != nil {
			return nil, err
		}
		return []transactionDelta{delta}, nil
	case "asset":
		if suggested.AmountAsset == nil || *suggested.AmountAsset <= 0 {
			return nil, fmt.Errorf("suggested asset amount missing")
		}
		sign, err := insightTradeSign(plan, insight.Type)
		if err != nil {
			return nil, err
		}
		delta, err := s.buildInsightDelta(ctx, insight, plan, *suggested.AmountAsset, "asset", sign, assetTypes, seriesByAssetKey)
		if err != nil {
			return nil, err
		}
		return []transactionDelta{delta}, nil
	case "rebalance":
		if plan.StrategyID != "S22" {
			return nil, fmt.Errorf("rebalance suggested quantity is only valid for S22")
		}
		return s.buildInsightRebalanceDeltas(ctx, suggested.Trades, assetTypes, seriesByAssetKey)
	default:
		return nil, fmt.Errorf("unsupported suggested quantity mode")
	}
}

func (s *Server) buildInsightDeltasFromManual(ctx context.Context, insight Insight, plan lockedPlan, suggested *suggestedQuantity, req insightExecuteRequest, assetTypes map[string]string, seriesByAssetKey map[string][]ohlcPoint) ([]transactionDelta, error) {
	if plan.StrategyID == "S22" {
		return nil, fmt.Errorf("manual execution is not supported for rebalance insights")
	}
	quantity := req.Quantity
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be greater than zero")
	}
	unit := strings.ToLower(strings.TrimSpace(req.QuantityUnit))
	if unit == "" {
		unit = "asset"
	}
	if unit != "asset" && unit != "usd" && unit != "base" && len(unit) != 3 {
		return nil, fmt.Errorf("quantity_unit must be asset, usd, base, or a 3-letter currency")
	}
	if unit != "asset" && unit != "usd" {
		converted, err := s.convertInsightAmountToUSD(ctx, insight.UserID, quantity, unit)
		if err != nil {
			return nil, err
		}
		quantity = converted
		unit = "usd"
	}
	if suggested != nil {
		switch strings.ToLower(strings.TrimSpace(suggested.Mode)) {
		case "rebalance":
			return nil, fmt.Errorf("manual execution is not supported for rebalance insights")
		case "usd":
			if unit != "usd" {
				return nil, fmt.Errorf("quantity_unit must match suggested usd")
			}
		case "asset":
			if unit != "asset" {
				return nil, fmt.Errorf("quantity_unit must match suggested asset")
			}
		}
	}
	sign, err := insightTradeSign(plan, insight.Type)
	if err != nil {
		return nil, err
	}
	delta, err := s.buildInsightDelta(ctx, insight, plan, quantity, unit, sign, assetTypes, seriesByAssetKey)
	if err != nil {
		return nil, err
	}
	return []transactionDelta{delta}, nil
}

func (s *Server) convertInsightAmountToUSD(ctx context.Context, userID string, amount float64, unit string) (float64, error) {
	if amount <= 0 {
		return 0, fmt.Errorf("amount must be greater than zero")
	}
	unit = strings.ToUpper(strings.TrimSpace(unit))
	if unit == "" {
		unit = "USD"
	}
	if unit == "BASE" {
		profile, err := s.ensureUserProfile(ctx, userID)
		if err != nil {
			return 0, fmt.Errorf("failed to load profile")
		}
		unit = normalizeCurrency(profile.BaseCurrency)
	}
	if unit == "USD" || stablecoinSet()[unit] {
		return amount, nil
	}
	oerResp, err := s.market.openExchangeLatest(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to load FX rates")
	}
	if rateToUSD, ok := currencyRateToUSD(unit, oerResp.Rates); ok {
		return amount * rateToUSD, nil
	}
	return 0, fmt.Errorf("unsupported currency %s", unit)
}

func (s *Server) buildInsightDelta(ctx context.Context, insight Insight, plan lockedPlan, amount float64, unit string, sign float64, assetTypes map[string]string, seriesByAssetKey map[string][]ohlcPoint) (transactionDelta, error) {
	if amount <= 0 {
		return transactionDelta{}, fmt.Errorf("amount must be greater than zero")
	}
	assetKey := strings.TrimSpace(plan.AssetKey)
	if assetKey == "" {
		assetKey = strings.TrimSpace(derefString(insight.AssetKey))
	}
	symbol := strings.TrimSpace(plan.Symbol)
	if symbol == "" {
		symbol = strings.TrimSpace(insight.Asset)
	}
	if assetKey == "" || symbol == "" {
		return transactionDelta{}, fmt.Errorf("asset key and symbol required")
	}
	assetType := strings.TrimSpace(plan.AssetType)
	if assetType == "" {
		assetType = assetTypes[assetKey]
	}
	if assetType == "" {
		assetType = assetTypeFromAssetKey(assetKey)
	}
	if assetType == "" {
		return transactionDelta{}, fmt.Errorf("asset type not found for insight")
	}
	price, err := s.resolveInsightPrice(ctx, assetKey, symbol, assetType, seriesByAssetKey)
	if err != nil {
		return transactionDelta{}, err
	}
	amountAsset := amount
	if unit == "usd" {
		if price <= 0 {
			return transactionDelta{}, fmt.Errorf("price not available for usd conversion")
		}
		amountAsset = amount / price
	}
	if amountAsset <= 0 {
		return transactionDelta{}, fmt.Errorf("asset amount must be greater than zero")
	}
	return transactionDelta{
		Symbol:      symbol,
		AssetType:   assetType,
		AssetKey:    assetKey,
		Amount:      sign * amountAsset,
		PriceUSD:    price,
		PriceNative: price,
		Currency:    "USD",
	}, nil
}

func (s *Server) buildInsightRebalanceDeltas(ctx context.Context, trades []rebalanceTrade, assetTypes map[string]string, seriesByAssetKey map[string][]ohlcPoint) ([]transactionDelta, error) {
	if len(trades) == 0 {
		return nil, fmt.Errorf("rebalance trades missing")
	}
	deltas := make([]transactionDelta, 0, len(trades))
	for _, trade := range trades {
		assetKey := strings.TrimSpace(trade.AssetKey)
		symbol := strings.TrimSpace(trade.Symbol)
		if assetKey == "" || symbol == "" {
			return nil, fmt.Errorf("rebalance trade missing asset")
		}
		assetType := assetTypes[assetKey]
		if assetType == "" {
			assetType = assetTypeFromAssetKey(assetKey)
		}
		if assetType == "" {
			return nil, fmt.Errorf("asset type missing for rebalance trade")
		}
		price, err := s.resolveInsightPrice(ctx, assetKey, symbol, assetType, seriesByAssetKey)
		if err != nil {
			return nil, err
		}
		amount := 0.0
		if trade.AmountAsset != nil && *trade.AmountAsset > 0 {
			amount = *trade.AmountAsset
		} else if trade.AmountUSD > 0 && price > 0 {
			amount = trade.AmountUSD / price
		}
		if amount <= 0 {
			return nil, fmt.Errorf("rebalance trade amount missing")
		}
		sign := 1.0
		switch strings.ToLower(strings.TrimSpace(trade.Side)) {
		case "buy":
			sign = 1
		case "sell":
			sign = -1
		default:
			return nil, fmt.Errorf("rebalance trade side must be buy or sell")
		}
		deltas = append(deltas, transactionDelta{
			Symbol:      symbol,
			AssetType:   assetType,
			AssetKey:    assetKey,
			Amount:      sign * amount,
			PriceUSD:    price,
			PriceNative: price,
			Currency:    "USD",
		})
	}
	return deltas, nil
}

func (s *Server) resolveInsightPrice(ctx context.Context, assetKey, symbol, assetType string, seriesByAssetKey map[string][]ohlcPoint) (float64, error) {
	series, ok := indicatorSeriesForHolding(ctx, s.market, portfolioHolding{
		AssetType: assetType,
		Symbol:    symbol,
		AssetKey:  assetKey,
	}, seriesByAssetKey)
	if !ok || len(series.Points) == 0 {
		return 0, fmt.Errorf("price unavailable for %s", symbol)
	}
	return series.Points[len(series.Points)-1].Close, nil
}

func insightTradeSign(plan lockedPlan, insightType string) (float64, error) {
	switch insightType {
	case insightTypePortfolioWatch:
		switch plan.StrategyID {
		case "S01", "S04":
			return -1, nil
		}
	case insightTypeActionAlert:
		switch plan.StrategyID {
		case "S02", "S05", "S09", "S16":
			return 1, nil
		case "S03":
			return -1, nil
		case "S18":
			action := strings.ToLower(strings.TrimSpace(getStringParam(plan.Parameters, "trend_action")))
			switch action {
			case "reduce_exposure":
				return -1, nil
			case "hold_or_add":
				return 1, nil
			case "hold", "wait":
				return 0, fmt.Errorf("trend action does not require execution")
			default:
				return 0, fmt.Errorf("unsupported trend action")
			}
		}
	}
	return 0, fmt.Errorf("unsupported insight strategy")
}

func assetTypeFromAssetKey(assetKey string) string {
	switch {
	case strings.HasPrefix(assetKey, "crypto:"):
		return "crypto"
	case strings.HasPrefix(assetKey, "stock:"):
		return "stock"
	case strings.HasPrefix(assetKey, "forex:"):
		return "forex"
	default:
		return ""
	}
}

func (s *Server) validateInsightTransactions(ctx context.Context, userID string, ids []string) ([]string, error) {
	unique := make(map[string]struct{}, len(ids))
	ordered := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := unique[id]; ok {
			continue
		}
		unique[id] = struct{}{}
		ordered = append(ordered, id)
	}
	if len(ordered) == 0 {
		return nil, fmt.Errorf("transaction_ids required")
	}

	var rows []PortfolioTransaction
	if err := s.db.DB().WithContext(ctx).
		Where("user_id = ? AND id IN (?)", userID, ordered).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) != len(ordered) {
		return nil, fmt.Errorf("one or more transaction_ids not found")
	}
	return ordered, nil
}

func (s *Server) updateInsightStatus(ctx context.Context, insightID, userID, status, eventType string, metadata map[string]any) error {
	if metadata == nil {
		metadata = map[string]any{}
	}
	now := time.Now().UTC()
	return s.db.withTx(ctx, func(tx *gorm.DB) error {
		if err := tx.Model(&Insight{}).
			Where("id = ? AND user_id = ?", insightID, userID).
			Update("status", status).Error; err != nil {
			return err
		}
		event := InsightEvent{
			ID:        newID("iev"),
			InsightID: insightID,
			EventType: eventType,
			Metadata:  marshalJSON(metadata),
			CreatedAt: now,
		}
		return tx.Create(&event).Error
	})
}
