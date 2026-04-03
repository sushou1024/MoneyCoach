package app

import (
	"context"
	"net/http"
	"strings"
	"time"
)

type portfolioSnapshotResponse struct {
	PortfolioSnapshotID  string                 `json:"portfolio_snapshot_id"`
	MarketDataSnapshotID string                 `json:"market_data_snapshot_id"`
	ValuationAsOf        string                 `json:"valuation_as_of"`
	SnapshotType         string                 `json:"snapshot_type"`
	NetWorthUSD          float64                `json:"net_worth_usd"`
	Holdings             []portfolioHoldingView `json:"holdings"`
	UnpricedHoldings     []portfolioHoldingView `json:"unpriced_holdings"`
	DashboardMetrics     *dashboardMetrics      `json:"dashboard_metrics,omitempty"`
}

type portfolioHoldingView struct {
	AssetType           string   `json:"asset_type"`
	Symbol              string   `json:"symbol"`
	AssetKey            string   `json:"asset_key"`
	CoinGeckoID         *string  `json:"coingecko_id,omitempty"`
	ExchangeMIC         *string  `json:"exchange_mic,omitempty"`
	Name                *string  `json:"name,omitempty"`
	LogoURL             *string  `json:"logo_url,omitempty"`
	Amount              float64  `json:"amount"`
	ValueFromScreenshot *float64 `json:"value_from_screenshot,omitempty"`
	ValueUSD            float64  `json:"value_usd_priced"`
	CurrentPrice        float64  `json:"current_price"`
	ValueQuote          float64  `json:"value_quote"`
	QuoteCurrency       string   `json:"quote_currency"`
	PricingSource       string   `json:"pricing_source"`
	ValuationStatus     string   `json:"valuation_status"`
	CurrencyConverted   bool     `json:"currency_converted"`
	CostBasisStatus     string   `json:"cost_basis_status"`
	BalanceType         string   `json:"balance_type"`
	AvgPrice            *float64 `json:"avg_price,omitempty"`
	AvgPriceQuote       *float64 `json:"avg_price_quote,omitempty"`
	AvgPriceSource      *string  `json:"avg_price_source,omitempty"`
	PNLPercent          *float64 `json:"pnl_percent,omitempty"`
	Sources             []string `json:"sources"`
	ActionBias          *string  `json:"action_bias,omitempty"`
}

type dashboardMetrics struct {
	NetWorthUSD       float64 `json:"net_worth_usd"`
	NetWorthDisplay   float64 `json:"net_worth_display"`
	BaseCurrency      string  `json:"base_currency"`
	BaseFXRateToUSD   float64 `json:"base_fx_rate_to_usd"`
	HealthScore       int     `json:"health_score"`
	HealthStatus      string  `json:"health_status"`
	VolatilityScore   int     `json:"volatility_score"`
	ValuationAsOf     string  `json:"valuation_as_of"`
	MetricsIncomplete bool    `json:"metrics_incomplete"`
	ScoreMode         string  `json:"score_mode"`
}

func (s *Server) handlePortfolioActive(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}

	var user User
	if err := s.db.DB().WithContext(r.Context()).First(&user, "id = ?", userID).Error; err != nil || user.ActivePortfolioSnapshot == nil {
		s.writeError(w, http.StatusNotFound, "NOT_FOUND", "active portfolio not found", nil)
		return
	}

	resp, err := s.buildPortfolioSnapshotResponse(r.Context(), userID, *user.ActivePortfolioSnapshot)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "PORTFOLIO_ERROR", "failed to load portfolio", nil)
		return
	}

	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handlePortfolioActiveRefresh(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}

	snapshotID, err := s.refreshActivePortfolio(r.Context(), userID)
	if err != nil {
		if err == errNotFound {
			s.writeError(w, http.StatusNotFound, "NOT_FOUND", "active portfolio not found", nil)
			return
		}
		s.writeError(w, http.StatusInternalServerError, "PORTFOLIO_ERROR", "failed to refresh portfolio", nil)
		return
	}

	resp, err := s.buildPortfolioSnapshotResponse(r.Context(), userID, snapshotID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "PORTFOLIO_ERROR", "failed to load refreshed portfolio", nil)
		return
	}

	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) buildPortfolioSnapshotResponse(ctx context.Context, userID string, snapshotID string) (portfolioSnapshotResponse, error) {
	snapshot, holdings, err := s.loadSnapshotWithHoldings(ctx, snapshotID)
	if err != nil {
		return portfolioSnapshotResponse{}, err
	}

	oerResp, _ := s.market.openExchangeLatest(ctx)
	metrics := computePortfolioMetrics(holdings, fetchPriceSeriesAsOf(ctx, s.market, holdings, snapshot.ValuationAsOf))
	profile, _ := s.ensureUserProfile(ctx, userID)
	baseCurrency, rateFromUSD := resolveDisplayCurrency(profile.BaseCurrency, oerResp.Rates)
	converted := metrics.NetWorthUSD * rateFromUSD
	baseRateToUSD := rateToUSDFromRateFromUSD(rateFromUSD)

	actionBiases := s.buildPortfolioHoldingActionBiases(ctx, snapshot, holdings)

	resp := portfolioSnapshotResponse{
		PortfolioSnapshotID:  snapshot.ID,
		MarketDataSnapshotID: snapshot.MarketDataSnapshotID,
		ValuationAsOf:        snapshot.ValuationAsOf.Format(time.RFC3339),
		SnapshotType:         snapshot.SnapshotType,
		NetWorthUSD:          snapshot.NetWorthUSD,
		Holdings:             s.mapHoldingsView(ctx, holdings, false, actionBiases, profile.Language),
		UnpricedHoldings:     s.mapHoldingsView(ctx, filterUnpricedHoldings(holdings), true, nil, profile.Language),
		DashboardMetrics: &dashboardMetrics{
			NetWorthUSD:       metrics.NetWorthUSD,
			NetWorthDisplay:   converted,
			BaseCurrency:      baseCurrency,
			BaseFXRateToUSD:   baseRateToUSD,
			HealthScore:       metrics.HealthScoreBaseline,
			HealthStatus:      healthStatusFromScore(metrics.HealthScoreBaseline),
			VolatilityScore:   metrics.VolatilityScoreBaseline,
			ValuationAsOf:     snapshot.ValuationAsOf.Format(time.RFC3339),
			MetricsIncomplete: metrics.MetricsIncomplete,
			ScoreMode:         "lightweight",
		},
	}

	return resp, nil
}

func (s *Server) handlePortfolioSnapshot(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}
	id := chiURLParam(r, "portfolio_snapshot_id")
	snapshot, holdings, err := s.loadSnapshotWithHoldings(r.Context(), id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "NOT_FOUND", "snapshot not found", nil)
		return
	}
	if snapshot.UserID != userID {
		s.writeError(w, http.StatusForbidden, "FORBIDDEN", "snapshot not accessible", nil)
		return
	}
	profile, _ := s.ensureUserProfile(r.Context(), userID)
	resp := portfolioSnapshotResponse{
		PortfolioSnapshotID:  snapshot.ID,
		MarketDataSnapshotID: snapshot.MarketDataSnapshotID,
		ValuationAsOf:        snapshot.ValuationAsOf.Format(time.RFC3339),
		SnapshotType:         snapshot.SnapshotType,
		NetWorthUSD:          snapshot.NetWorthUSD,
		Holdings:             s.mapHoldingsView(r.Context(), holdings, false, nil, profile.Language),
		UnpricedHoldings:     s.mapHoldingsView(r.Context(), filterUnpricedHoldings(holdings), true, nil, profile.Language),
	}
	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handlePortfolioSnapshots(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	if userID == "" {
		s.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token", nil)
		return
	}
	limit := parseLimit(r, 20, 100)
	cursor := parseCursor(r)

	query := s.db.DB().WithContext(r.Context()).Where("user_id = ?", userID).Order("created_at desc").Limit(limit + 1)
	if cursor != "" {
		query = query.Where("id < ?", cursor)
	}

	var snapshots []PortfolioSnapshot
	if err := query.Find(&snapshots).Error; err != nil {
		s.writeError(w, http.StatusInternalServerError, "PORTFOLIO_ERROR", "failed to list snapshots", nil)
		return
	}

	items := make([]map[string]any, 0, len(snapshots))
	nextCursor := ""
	for i, snapshot := range snapshots {
		if i == limit {
			nextCursor = snapshot.ID
			break
		}
		calcID := s.findCalculationForSnapshot(r.Context(), snapshot.ID)
		items = append(items, map[string]any{
			"portfolio_snapshot_id": snapshot.ID,
			"created_at":            snapshot.CreatedAt.Format(time.RFC3339),
			"net_worth_usd":         snapshot.NetWorthUSD,
			"snapshot_type":         snapshot.SnapshotType,
			"status":                snapshot.Status,
			"calculation_id":        calcID,
		})
	}

	s.writeJSON(w, http.StatusOK, map[string]any{"items": items, "next_cursor": nextCursor})
}

func (s *Server) loadSnapshotWithHoldings(ctx context.Context, snapshotID string) (PortfolioSnapshot, []portfolioHolding, error) {
	var snapshot PortfolioSnapshot
	if err := s.db.DB().WithContext(ctx).First(&snapshot, "id = ?", snapshotID).Error; err != nil {
		return PortfolioSnapshot{}, nil, err
	}
	var rows []PortfolioHolding
	if err := s.db.DB().WithContext(ctx).Where("portfolio_snapshot_id = ?", snapshot.ID).Find(&rows).Error; err != nil {
		return snapshot, nil, err
	}
	return snapshot, s.enrichHoldingsWithQuoteMetadata(ctx, snapshot, mapHoldingsFromRows(rows)), nil
}

func mapHoldingsFromRows(rows []PortfolioHolding) []portfolioHolding {
	out := make([]portfolioHolding, 0, len(rows))
	stablecoins := stablecoinSet()
	for _, row := range rows {
		currentPrice := 0.0
		if row.Amount > 0 {
			currentPrice = row.ValueUSD / row.Amount
		}
		balanceType := row.BalanceType
		if balanceType == "" || balanceType == "unknown" {
			if stablecoins[normalizeSymbol(row.Symbol)] {
				balanceType = "stablecoin"
			}
		}
		holding := portfolioHolding{
			Symbol:              row.Symbol,
			SymbolRaw:           row.Symbol,
			AssetType:           row.AssetType,
			AssetKey:            row.AssetKey,
			CoinGeckoID:         derefString(row.CoinGeckoID),
			ExchangeMIC:         derefString(row.ExchangeMIC),
			Amount:              row.Amount,
			ValueUSD:            row.ValueUSD,
			ValueFromScreenshot: row.ValueFromScreenshot,
			PricingSource:       row.PricingSource,
			ValuationStatus:     row.ValuationStatus,
			BalanceType:         balanceType,
			CurrentPrice:        currentPrice,
			AvgPrice:            row.AvgPrice,
			AvgPriceSource:      derefString(row.AvgPriceSource),
			PNLPercent:          row.PNLPercent,
			CostBasisStatus:     row.CostBasisStatus,
			CurrencyConverted:   row.CurrencyConverted,
		}
		out = append(out, holding)
	}
	return out
}

func (s *Server) mapHoldingsView(
	ctx context.Context,
	holdings []portfolioHolding,
	onlyUnpriced bool,
	actionBiases map[string]string,
	language string,
) []portfolioHoldingView {
	out := make([]portfolioHoldingView, 0, len(holdings))
	logosByAssetKey := s.resolveHoldingLogos(ctx, holdings)
	nameBySymbol := map[string]*string{}
	if s != nil && s.market != nil {
		for _, holding := range holdings {
			displaySymbol := hkDisplaySymbol(holding.Symbol)
			if !isHongKongStockSymbol(displaySymbol, holding.ExchangeMIC) {
				continue
			}
			if _, ok := nameBySymbol[displaySymbol]; ok {
				continue
			}
			if localized, ok := hkNameForLanguage(displaySymbol, language); ok {
				nameBySymbol[displaySymbol] = &localized
				continue
			}
			ticker, err := s.market.marketstackTicker(ctx, displaySymbol)
			if err != nil {
				nameBySymbol[displaySymbol] = nil
				continue
			}
			name := strings.TrimSpace(ticker.Name)
			if name == "" {
				nameBySymbol[displaySymbol] = nil
				continue
			}
			nameBySymbol[displaySymbol] = &name
		}
	}
	for _, holding := range holdings {
		if onlyUnpriced && holding.ValuationStatus != "unpriced" {
			continue
		}
		displaySymbol := hkDisplaySymbol(holding.Symbol)
		currentPrice := roundTo(quotePriceForHolding(holding), priceDecimals(holding.AssetType))
		valueQuote := roundTo(quoteValueForHolding(holding), 2)
		avgPriceQuote := quoteAvgPriceForHolding(holding)
		if avgPriceQuote != nil {
			value := roundTo(*avgPriceQuote, priceDecimals(holding.AssetType))
			avgPriceQuote = &value
		}
		view := portfolioHoldingView{
			AssetType:           holding.AssetType,
			Symbol:              displaySymbol,
			AssetKey:            holding.AssetKey,
			CoinGeckoID:         nullableString(holding.CoinGeckoID),
			ExchangeMIC:         nullableString(holding.ExchangeMIC),
			Name:                nameBySymbol[displaySymbol],
			LogoURL:             nullableString(logosByAssetKey[holding.AssetKey]),
			Amount:              holding.Amount,
			ValueFromScreenshot: holding.ValueFromScreenshot,
			ValueUSD:            holding.ValueUSD,
			CurrentPrice:        currentPrice,
			ValueQuote:          valueQuote,
			QuoteCurrency:       quoteCurrencyForHolding(holding),
			PricingSource:       holding.PricingSource,
			ValuationStatus:     holding.ValuationStatus,
			CurrencyConverted:   holding.CurrencyConverted,
			CostBasisStatus:     holding.CostBasisStatus,
			BalanceType:         holding.BalanceType,
			AvgPrice:            holding.AvgPrice,
			AvgPriceQuote:       avgPriceQuote,
			AvgPriceSource:      nullableString(holding.AvgPriceSource),
			PNLPercent:          holding.PNLPercent,
			Sources:             []string{},
		}
		if actionBiases != nil {
			if actionBias, ok := actionBiases[holding.AssetKey]; ok {
				view.ActionBias = &actionBias
			}
		}
		out = append(out, view)
	}
	return out
}

func (s *Server) buildPortfolioHoldingActionBiases(
	ctx context.Context,
	snapshot PortfolioSnapshot,
	holdings []portfolioHolding,
) map[string]string {
	if len(holdings) == 0 {
		return nil
	}
	end := snapshot.ValuationAsOf.UTC()
	if end.IsZero() {
		end = time.Now().UTC()
	}
	start := end.AddDate(0, 0, -intelligenceLookbackDays)
	seriesByAssetKey := fetchPriceSeriesRange(ctx, s.market, holdings, start, end)
	portfolioReturns := returnsMap(portfolioReturnsFromIntersection(buildEligibleReturnSeries(holdings, seriesByAssetKey, 20)))
	actionBiases := make(map[string]string)
	for _, holding := range holdings {
		if holding.AssetKey == "" || holding.ValuationStatus != "priced" || isCashLike(holding) {
			continue
		}
		data, ok := buildIntelligenceCoreData(snapshot, holding, true, seriesByAssetKey[holding.AssetKey], portfolioReturns)
		if !ok || data.ActionBias == "" {
			continue
		}
		actionBiases[holding.AssetKey] = data.ActionBias
	}
	if len(actionBiases) == 0 {
		return nil
	}
	return actionBiases
}

func filterUnpricedHoldings(holdings []portfolioHolding) []portfolioHolding {
	result := make([]portfolioHolding, 0)
	for _, holding := range holdings {
		if holding.ValuationStatus == "unpriced" {
			result = append(result, holding)
		}
	}
	return result
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (s *Server) convertUSDToBase(ctx context.Context, usd float64, base string) float64 {
	if base == "USD" || usd == 0 {
		return usd
	}
	resp, err := s.market.openExchangeLatest(ctx)
	if err != nil {
		return usd
	}
	rate, ok := resp.Rates[base]
	if !ok || rate == 0 {
		return usd
	}
	return usd * rate
}

func resolveDisplayCurrency(base string, rates map[string]float64) (string, float64) {
	base = normalizeCurrency(base)
	rateFromUSD, ok := currencyRateFromUSD(base, rates)
	if !ok || rateFromUSD <= 0 {
		return "USD", 1
	}
	return base, rateFromUSD
}

func (s *Server) findCalculationForSnapshot(ctx context.Context, snapshotID string) string {
	var calc Calculation
	if err := s.db.DB().WithContext(ctx).Where("portfolio_snapshot_id = ?", snapshotID).First(&calc).Error; err != nil {
		return ""
	}
	return calc.ID
}
