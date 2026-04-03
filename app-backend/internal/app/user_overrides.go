package app

import (
	"context"
	"strings"
)

func (s *Server) loadUserAssetOverrides(ctx context.Context, userID string) map[string]UserAssetOverride {
	if strings.TrimSpace(userID) == "" {
		return nil
	}
	var rows []UserAssetOverride
	if err := s.db.DB().WithContext(ctx).Where("user_id = ?", userID).Find(&rows).Error; err != nil {
		return nil
	}
	byKey := make(map[string]UserAssetOverride, len(rows))
	for _, row := range rows {
		if row.AssetKey == "" {
			continue
		}
		byKey[row.AssetKey] = row
	}
	return byKey
}

func applyUserOverride(holding *portfolioHolding, overrides map[string]UserAssetOverride) {
	if holding == nil || holding.AssetKey == "" || holding.AvgPrice != nil {
		return
	}
	if overrides == nil {
		return
	}
	override, ok := overrides[holding.AssetKey]
	if !ok || override.AvgPrice <= 0 {
		return
	}
	avg := override.AvgPrice
	if override.AvgPriceSource != "" {
		holding.AvgPriceSource = override.AvgPriceSource
	} else {
		holding.AvgPriceSource = "user_input"
	}
	holding.AvgPrice = &avg
}
