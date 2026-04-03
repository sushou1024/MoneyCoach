package app

import (
	"fmt"
	"sort"
	"strings"
)

type assetAllocationItem struct {
	Label           string  `json:"label"`
	ValueUSD        float64 `json:"value_usd"`
	ValueDisplay    float64 `json:"value_display"`
	DisplayCurrency string  `json:"display_currency"`
	WeightPct       float64 `json:"weight_pct"`
}

func buildAssetAllocation(holdings []portfolioHolding, netWorth float64, baseCurrency string, rateFromUSD float64) []assetAllocationItem {
	if netWorth <= 0 {
		return nil
	}
	if rateFromUSD <= 0 {
		rateFromUSD = 1
		baseCurrency = "USD"
	}
	buckets := map[string]float64{
		"crypto": 0,
		"stock":  0,
		"cash":   0,
		"manual": 0,
	}
	for _, holding := range holdings {
		if holding.ValuationStatus != "priced" && holding.ValuationStatus != "user_provided" {
			continue
		}
		label := ""
		if strings.HasPrefix(holding.AssetKey, "manual:") && holding.ValuationStatus == "user_provided" {
			label = "manual"
		} else if holding.AssetType == "crypto" && holding.BalanceType != "stablecoin" {
			label = "crypto"
		} else if holding.AssetType == "stock" {
			label = "stock"
		} else if isCashLike(holding) || holding.BalanceType == "stablecoin" {
			label = "cash"
		}
		if label == "" {
			continue
		}
		buckets[label] += holding.ValueUSD
	}

	order := []string{"crypto", "stock", "cash", "manual"}
	out := make([]assetAllocationItem, 0, len(order))
	for _, label := range order {
		value := buckets[label]
		if value <= 0 {
			continue
		}
		out = append(out, assetAllocationItem{
			Label:           label,
			ValueUSD:        roundTo(value, 2),
			ValueDisplay:    roundTo(value*rateFromUSD, 2),
			DisplayCurrency: baseCurrency,
			WeightPct:       roundTo(value/netWorth, 4),
		})
	}
	return out
}

func buildIdentifiedRisks(metrics portfolioMetrics) []previewRisk {
	liquidity := previewRisk{
		RiskID:   "risk_01",
		Type:     "Liquidity Risk",
		Severity: liquiditySeverity(metrics.CashPct),
	}
	concentration := previewRisk{
		RiskID:   "risk_02",
		Type:     "Concentration Risk",
		Severity: concentrationSeverity(metrics.TopAssetPct),
	}
	remaining := []riskCandidate{
		{Type: "Drawdown Risk", Severity: drawdownSeverity(metrics.MaxDrawdown90d), Priority: 0},
		{Type: "Volatility Risk", Severity: volatilitySeverity(metrics.Volatility30dAnnualized), Priority: 1},
		{Type: "Correlation Risk", Severity: correlationSeverity(metrics.AvgPairwiseCorr), Priority: 2},
		{Type: "Inefficient Capital Risk", Severity: inefficientCapitalSeverity(metrics.CashPct), Priority: 3},
	}
	sort.SliceStable(remaining, func(i, j int) bool {
		rankI := riskSeverityRank(remaining[i].Severity)
		rankJ := riskSeverityRank(remaining[j].Severity)
		if rankI != rankJ {
			return rankI < rankJ
		}
		return remaining[i].Priority < remaining[j].Priority
	})

	risk03 := previewRisk{
		RiskID:   "risk_03",
		Type:     remaining[0].Type,
		Severity: remaining[0].Severity,
	}
	return []previewRisk{liquidity, concentration, risk03}
}

func validatePreviewRisks(risks []previewRisk, metrics portfolioMetrics) []string {
	var violations []string
	if len(risks) != 3 {
		violations = append(violations, fmt.Sprintf("identified_risks must have 3 items (got %d)", len(risks)))
	}

	allowedTypes := map[string]struct{}{
		"Liquidity Risk":           {},
		"Concentration Risk":       {},
		"Volatility Risk":          {},
		"Correlation Risk":         {},
		"Drawdown Risk":            {},
		"Inefficient Capital Risk": {},
	}
	allowedSeverity := map[string]struct{}{
		"Low":      {},
		"Medium":   {},
		"High":     {},
		"Critical": {},
	}

	ids := make(map[string]struct{}, len(risks))
	types := make(map[string]struct{}, len(risks))
	teasers := make(map[string]struct{}, len(risks))
	for _, risk := range risks {
		id := strings.TrimSpace(risk.RiskID)
		if id == "" {
			violations = append(violations, "risk_id is required")
		} else {
			if id != "risk_01" && id != "risk_02" && id != "risk_03" {
				violations = append(violations, fmt.Sprintf("unexpected risk_id %s", id))
			}
			if _, ok := ids[id]; ok {
				violations = append(violations, fmt.Sprintf("duplicate risk_id %s", id))
			}
			ids[id] = struct{}{}
		}

		typ := strings.TrimSpace(risk.Type)
		if _, ok := allowedTypes[typ]; !ok {
			violations = append(violations, fmt.Sprintf("unexpected risk type %s", typ))
		} else {
			if _, ok := types[typ]; ok {
				violations = append(violations, fmt.Sprintf("duplicate risk type %s", typ))
			}
			types[typ] = struct{}{}
		}

		severity := strings.TrimSpace(risk.Severity)
		if _, ok := allowedSeverity[severity]; !ok {
			violations = append(violations, fmt.Sprintf("unexpected severity %s", severity))
		}

		teaser := strings.TrimSpace(risk.TeaserText)
		if teaser == "" {
			violations = append(violations, fmt.Sprintf("missing teaser_text for %s", id))
		} else {
			key := strings.ToLower(teaser)
			if _, ok := teasers[key]; ok {
				violations = append(violations, "duplicate teaser_text")
			}
			teasers[key] = struct{}{}
		}
	}

	if metrics.MetricsIncomplete {
		for _, risk := range risks {
			if risk.RiskID == "risk_03" && strings.TrimSpace(risk.Severity) == "Low" {
				violations = append(violations, "risk_03 severity must be at least Medium when metrics_incomplete=true")
				break
			}
		}
	}

	return violations
}

func mergePaidRisks(paid *paidReportPayload, previewRisks []previewRisk) {
	if paid == nil {
		return
	}
	messageByType := make(map[string]string)
	messageByID := make(map[string]string)
	for _, risk := range paid.RiskInsights {
		if risk.Message != "" {
			messageByType[risk.Type] = risk.Message
			messageByID[risk.RiskID] = risk.Message
		}
	}
	merged := make([]paidRisk, 0, len(previewRisks))
	for _, risk := range previewRisks {
		message := messageByID[risk.RiskID]
		if message == "" {
			message = messageByType[risk.Type]
		}
		if message == "" {
			message = "Risk detected based on your current portfolio mix."
		}
		merged = append(merged, paidRisk{
			RiskID:   risk.RiskID,
			Type:     risk.Type,
			Severity: risk.Severity,
			Message:  message,
		})
	}
	paid.RiskInsights = merged
	paid.ExposureAnalysis = append([]paidRisk(nil), merged...)
	if paid.TheVerdict.ConstructiveComment != "" {
		paid.RiskSummary = paid.TheVerdict.ConstructiveComment
	}
}

type riskCandidate struct {
	Type     string
	Severity string
	Priority int
}

func liquiditySeverity(cashPct float64) string {
	switch {
	case cashPct < 0.03:
		return "Critical"
	case cashPct < 0.08:
		return "High"
	case cashPct < 0.15:
		return "Medium"
	default:
		return "Low"
	}
}

func concentrationSeverity(topAssetPct float64) string {
	switch {
	case topAssetPct >= 0.70:
		return "Critical"
	case topAssetPct >= 0.50:
		return "High"
	case topAssetPct >= 0.35:
		return "Medium"
	default:
		return "Low"
	}
}

func volatilitySeverity(volAnnual float64) string {
	switch {
	case volAnnual >= 0.80:
		return "Critical"
	case volAnnual >= 0.60:
		return "High"
	case volAnnual >= 0.40:
		return "Medium"
	default:
		return "Low"
	}
}

func drawdownSeverity(drawdown float64) string {
	switch {
	case drawdown >= 0.50:
		return "Critical"
	case drawdown >= 0.35:
		return "High"
	case drawdown >= 0.20:
		return "Medium"
	default:
		return "Low"
	}
}

func correlationSeverity(corr float64) string {
	switch {
	case corr >= 0.80:
		return "Critical"
	case corr >= 0.65:
		return "High"
	case corr >= 0.50:
		return "Medium"
	default:
		return "Low"
	}
}

func inefficientCapitalSeverity(cashPct float64) string {
	switch {
	case cashPct >= 0.60:
		return "Critical"
	case cashPct >= 0.45:
		return "High"
	case cashPct >= 0.30:
		return "Medium"
	default:
		return "Low"
	}
}

func riskSeverityRank(severity string) int {
	switch strings.ToLower(strings.TrimSpace(severity)) {
	case "critical":
		return 0
	case "high":
		return 1
	case "medium":
		return 2
	default:
		return 3
	}
}
