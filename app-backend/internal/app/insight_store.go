package app

func buildInsightRecord(userID string, item insightItem) Insight {
	assetKey := nullableString(item.AssetKey)
	timeframe := nullableString(item.Timeframe)
	strategyID := nullableString(item.StrategyID)
	planID := nullableString(item.PlanID)
	suggestedAction := nullableString(item.SuggestedAction)
	record := Insight{
		ID:                item.ID,
		UserID:            userID,
		Type:              item.Type,
		Asset:             item.Asset,
		AssetKey:          assetKey,
		Timeframe:         timeframe,
		Severity:          item.Severity,
		TriggerKey:        item.TriggerKey,
		TriggerReason:     item.TriggerReason,
		StrategyID:        strategyID,
		PlanID:            planID,
		SuggestedAction:   suggestedAction,
		SuggestedQuantity: marshalJSON(item.SuggestedQuantity),
		CTAPayload:        marshalJSON(item.CTAPayload),
		Status:            "active",
		CreatedAt:         item.CreatedAt,
		ExpiresAt:         item.ExpiresAt,
	}
	if item.BetaToPortfolio != 0 {
		beta := item.BetaToPortfolio
		record.BetaToPortfolio = &beta
	}
	return record
}
