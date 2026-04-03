package app

import (
	"context"
	"strings"
)

func fetchFuturesSnapshot(ctx context.Context, market *marketClient, holdings []portfolioHolding) map[string]futuresPremiumIndex {
	result := make(map[string]futuresPremiumIndex)
	if market == nil {
		return result
	}
	for _, holding := range holdings {
		if holding.AssetType != "crypto" {
			continue
		}
		if holding.ValuationStatus != "priced" {
			continue
		}
		if holding.BalanceType == "stablecoin" {
			continue
		}
		if holding.AssetKey == "" {
			continue
		}
		if _, ok := result[holding.AssetKey]; ok {
			continue
		}
		symbol := strings.ToUpper(strings.TrimSpace(holding.Symbol))
		if symbol == "" {
			continue
		}
		if futures, ok := fetchFuturesIndexForSymbol(ctx, market, symbol); ok {
			result[holding.AssetKey] = futures
		}
	}
	return result
}

func fetchFuturesIndexForSymbol(ctx context.Context, market *marketClient, symbol string) (futuresPremiumIndex, bool) {
	candidates := []string{symbol + "USDT", "1000" + symbol + "USDT"}
	for _, candidate := range candidates {
		futures, err := market.binanceFuturesPremiumIndex(ctx, candidate)
		if err != nil {
			continue
		}
		if futures.MarkPrice <= 0 {
			continue
		}
		return futures, true
	}
	return futuresPremiumIndex{}, false
}

func (s *Server) loadBatchTimezone(ctx context.Context, batchID string) string {
	if strings.TrimSpace(batchID) == "" {
		return ""
	}
	var batch UploadBatch
	if err := s.db.DB().WithContext(ctx).First(&batch, "id = ?", batchID).Error; err != nil {
		return ""
	}
	if batch.DeviceTimezone == nil {
		return ""
	}
	return strings.TrimSpace(*batch.DeviceTimezone)
}
