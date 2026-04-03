package app

import "math"

const s02OrderMultiplier = 1.5

func buildS02Plan(riskLevel string, metrics portfolioMetrics, ctx assetPlanContext) (lockedPlan, bool) {
	if ctx.Holding.AvgPrice == nil || ctx.Holding.CurrentPrice <= 0 {
		return lockedPlan{}, false
	}
	baseStep := s02BaseStepPct(riskLevel)
	volAdj := 1.0
	if metrics.Volatility30dDaily > 0.05 {
		volAdj = 1.2
	} else if metrics.Volatility30dDaily < 0.02 {
		volAdj = 0.8
	}
	priceStepPct := baseStep * volAdj
	maxAdditionalRatio := s02MaxAdditionalRatio(riskLevel)
	currentInvestment := ctx.Holding.Amount * ctx.Holding.CurrentPrice
	if currentInvestment <= 0 {
		return lockedPlan{}, false
	}
	maxAdditionalFunds := currentInvestment * maxAdditionalRatio
	actualAvailable := math.Min(metrics.IdleCashUSD, maxAdditionalFunds)
	if actualAvailable < 20 {
		return lockedPlan{}, false
	}
	initialOrderUSD := currentInvestment * 0.10
	if initialOrderUSD <= 0 {
		return lockedPlan{}, false
	}

	maxOrders := countOrdersWithinBudget(initialOrderUSD, s02OrderMultiplier, actualAvailable)
	if maxOrders < 3 {
		maxOrders = 3
	}
	if maxOrders > 6 {
		maxOrders = 6
	}

	totalNeeded := seriesSum(initialOrderUSD, s02OrderMultiplier, maxOrders)
	if totalNeeded > actualAvailable && totalNeeded > 0 {
		scale := actualAvailable / totalNeeded
		initialOrderUSD *= scale
	}
	baseOrderUSD := roundTo(initialOrderUSD, 2)

	takeProfitPct := s02TakeProfitPct(riskLevel)
	lossBuffer := s02LossBuffer(riskLevel)
	pnl := 0.0
	if ctx.Holding.PNLPercent != nil {
		pnl = math.Abs(*ctx.Holding.PNLPercent)
	}
	totalStopLossPct := pnl + lossBuffer
	totalStopLossPrice := *ctx.Holding.AvgPrice * (1 - totalStopLossPct)

	orders := make([]map[string]any, 0, maxOrders)
	cumulativeInvestment := ctx.Holding.Amount * *ctx.Holding.AvgPrice
	cumulativeAmount := ctx.Holding.Amount
	for i := 1; i <= maxOrders; i++ {
		triggerPrice := ctx.Holding.CurrentPrice * (1 - priceStepPct*float64(i))
		orderAmountUSD := baseOrderUSD * math.Pow(s02OrderMultiplier, float64(i-1))
		orderAmountAsset := 0.0
		if triggerPrice > 0 {
			orderAmountAsset = orderAmountUSD / triggerPrice
		}
		cumulativeInvestment += orderAmountUSD
		cumulativeAmount += orderAmountAsset
		newAvgCost := 0.0
		if cumulativeAmount > 0 {
			newAvgCost = cumulativeInvestment / cumulativeAmount
		}
		takeProfitPrice := newAvgCost * (1 + takeProfitPct)
		breakeven := 0.0
		if triggerPrice > 0 {
			breakeven = (newAvgCost - triggerPrice) / triggerPrice
		}

		orders = append(orders, map[string]any{
			"order_number":               i,
			"trigger_price":              roundTo(triggerPrice, priceDecimals(ctx.Holding.AssetType)),
			"order_amount_usd":           roundTo(orderAmountUSD, 2),
			"order_amount_asset":         roundTo(orderAmountAsset, amountDecimals(ctx.Holding.AssetType)),
			"new_avg_cost":               roundTo(newAvgCost, priceDecimals(ctx.Holding.AssetType)),
			"cumulative_investment":      roundTo(cumulativeInvestment, 2),
			"cumulative_amount":          roundTo(cumulativeAmount, amountDecimals(ctx.Holding.AssetType)),
			"take_profit_price":          roundTo(takeProfitPrice, priceDecimals(ctx.Holding.AssetType)),
			"breakeven_from_trigger_pct": roundTo(breakeven, 4),
		})
	}

	params := map[string]any{
		"price_step_pct":        roundTo(priceStepPct, 4),
		"max_safety_orders":     maxOrders,
		"safety_order_base_usd": baseOrderUSD,
		"order_multiplier":      roundTo(s02OrderMultiplier, 2),
		"take_profit_pct":       roundTo(takeProfitPct, 4),
		"total_stop_loss_pct":   roundTo(totalStopLossPct, 4),
		"total_stop_loss_price": roundTo(totalStopLossPrice, priceDecimals(ctx.Holding.AssetType)),
		"safety_orders":         orders,
	}

	return lockedPlan{
		StrategyID: "S02",
		AssetType:  ctx.Holding.AssetType,
		Symbol:     ctx.Holding.Symbol,
		AssetKey:   ctx.Holding.AssetKey,
		Parameters: params,
	}, true
}

func s02BaseStepPct(riskLevel string) float64 {
	switch riskLevel {
	case "conservative":
		return 0.03
	case "aggressive":
		return 0.015
	default:
		return 0.02
	}
}

func s02MaxAdditionalRatio(riskLevel string) float64 {
	switch riskLevel {
	case "conservative":
		return 0.30
	case "aggressive":
		return 0.80
	default:
		return 0.50
	}
}

func s02TakeProfitPct(riskLevel string) float64 {
	switch riskLevel {
	case "conservative":
		return 0.02
	case "aggressive":
		return 0.05
	default:
		return 0.03
	}
}

func s02LossBuffer(riskLevel string) float64 {
	switch riskLevel {
	case "conservative":
		return 0.15
	case "aggressive":
		return 0.10
	default:
		return 0.12
	}
}

func countOrdersWithinBudget(base, multiplier, budget float64) int {
	if base <= 0 || multiplier <= 1 {
		return 0
	}
	sum := base
	count := 1
	for sum+base*math.Pow(multiplier, float64(count)) <= budget {
		sum += base * math.Pow(multiplier, float64(count))
		count++
		if count > 10 {
			break
		}
	}
	return count
}

func seriesSum(base, multiplier float64, count int) float64 {
	sum := 0.0
	for i := 0; i < count; i++ {
		sum += base * math.Pow(multiplier, float64(i))
	}
	return sum
}
