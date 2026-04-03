package app

import (
	"encoding/json"
)

func convertLockedPlansToDisplay(plans []lockedPlan, rateFromUSD float64) []lockedPlan {
	if len(plans) == 0 {
		return plans
	}
	if rateFromUSD <= 0 {
		rateFromUSD = 1
	}
	out := make([]lockedPlan, 0, len(plans))
	for _, plan := range plans {
		clone := plan
		clone.Parameters = convertPlanParameters(plan, rateFromUSD)
		out = append(out, clone)
	}
	return out
}

func convertLockedPlansToDisplayByPlan(plans []lockedPlan, holdingByAsset map[string]*portfolioHolding, baseCurrency string, rateFromUSD float64) []lockedPlan {
	if len(plans) == 0 {
		return plans
	}
	out := make([]lockedPlan, 0, len(plans))
	for _, plan := range plans {
		clone := plan
		quoteCurrency, planRateFromUSD := planDisplayCurrency(plan, holdingByAsset, baseCurrency, rateFromUSD)
		clone.QuoteCurrency = quoteCurrency
		clone.Parameters = convertPlanParameters(plan, planRateFromUSD)
		out = append(out, clone)
	}
	return out
}

func convertPlanParameters(plan lockedPlan, rateFromUSD float64) map[string]any {
	if len(plan.Parameters) == 0 {
		return plan.Parameters
	}
	params := clonePlanParams(plan.Parameters)
	switch plan.StrategyID {
	case "S01":
		convertPrice(params, "stop_loss_price", plan.AssetType, rateFromUSD)
		convertPrice(params, "support_level", plan.AssetType, rateFromUSD)
	case "S02":
		convertMoney(params, "safety_order_base_usd", rateFromUSD)
		convertPrice(params, "total_stop_loss_price", plan.AssetType, rateFromUSD)
		if orders, ok := params["safety_orders"].([]any); ok {
			for _, entry := range orders {
				order, ok := entry.(map[string]any)
				if !ok {
					continue
				}
				convertPrice(order, "trigger_price", plan.AssetType, rateFromUSD)
				convertMoney(order, "order_amount_usd", rateFromUSD)
				convertPrice(order, "new_avg_cost", plan.AssetType, rateFromUSD)
				convertMoney(order, "cumulative_investment", rateFromUSD)
				convertPrice(order, "take_profit_price", plan.AssetType, rateFromUSD)
			}
		}
	case "S03":
		convertPrice(params, "activation_price", plan.AssetType, rateFromUSD)
		convertPrice(params, "initial_trailing_stop_price", plan.AssetType, rateFromUSD)
	case "S04":
		if layers, ok := params["layers"].([]any); ok {
			for _, entry := range layers {
				layer, ok := entry.(map[string]any)
				if !ok {
					continue
				}
				convertPrice(layer, "target_price", plan.AssetType, rateFromUSD)
				convertMoney(layer, "expected_profit_usd", rateFromUSD)
			}
		}
	case "S05":
		convertMoney(params, "amount", rateFromUSD)
	case "S09":
		convertMoney(params, "base_addition_usd", rateFromUSD)
		if additions, ok := params["additions"].([]any); ok {
			for _, entry := range additions {
				addition, ok := entry.(map[string]any)
				if !ok {
					continue
				}
				convertMoney(addition, "addition_amount_usd", rateFromUSD)
			}
		}
	case "S16":
		convertPrice(params, "spot_price", plan.AssetType, rateFromUSD)
		convertPrice(params, "mark_price", plan.AssetType, rateFromUSD)
		convertMoney(params, "hedge_notional_usd", rateFromUSD)
	case "S18":
		convertPrice(params, "current_price", plan.AssetType, rateFromUSD)
		convertPrice(params, "ma_20", plan.AssetType, rateFromUSD)
		convertPrice(params, "ma_50", plan.AssetType, rateFromUSD)
		convertPrice(params, "ma_200", plan.AssetType, rateFromUSD)
	}
	return params
}

func clonePlanParams(input map[string]any) map[string]any {
	if len(input) == 0 {
		return input
	}
	encoded, _ := json.Marshal(input)
	var out map[string]any
	_ = json.Unmarshal(encoded, &out)
	return out
}

func convertMoney(params map[string]any, key string, rateFromUSD float64) {
	value, ok := params[key].(float64)
	if !ok {
		return
	}
	params[key] = roundTo(value*rateFromUSD, 2)
}

func convertPrice(params map[string]any, key, assetType string, rateFromUSD float64) {
	value, ok := params[key].(float64)
	if !ok {
		return
	}
	params[key] = roundTo(value*rateFromUSD, priceDecimals(assetType))
}
