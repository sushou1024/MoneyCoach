package app

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

func buildInitialPlanStates(userID string, plans []lockedPlan, futuresByAssetKey map[string]futuresPremiumIndex, valuationAsOf time.Time) []PlanState {
	if userID == "" || len(plans) == 0 {
		return nil
	}
	now := time.Now().UTC()
	states := make([]PlanState, 0, len(plans))
	for _, plan := range plans {
		state := map[string]any{}
		switch plan.StrategyID {
		case "S02":
			if orders := planSafetyOrders(plan.Parameters); len(orders) > 0 {
				first := orders[0]
				price, _ := getFloatParam(first, "trigger_price")
				amountUSD, _ := getFloatParam(first, "order_amount_usd")
				state["next_safety_order_index"] = 0
				state["next_safety_order_price"] = price
				state["next_safety_order_amount_usd"] = amountUSD
			}
		case "S03":
			activation, _ := getFloatParam(plan.Parameters, "activation_price")
			trailingStop, _ := getFloatParam(plan.Parameters, "initial_trailing_stop_price")
			state["activated_at"] = valuationAsOf
			state["peak_price_since_activation"] = activation
			state["trailing_stop_price"] = trailingStop
		case "S05":
			if next := getStringParam(plan.Parameters, "next_execution_at"); next != "" {
				state["next_execution_at"] = next
			}
		case "S09":
			if additions := planAdditions(plan.Parameters); len(additions) > 0 {
				first := additions[0]
				trigger, _ := getFloatParam(first, "trigger_profit_pct")
				amountUSD, _ := getFloatParam(first, "addition_amount_usd")
				state["next_addition_index"] = 0
				state["next_trigger_profit_pct"] = trigger
				state["next_addition_amount_usd"] = amountUSD
			}
		case "S16":
			if futures, ok := futuresByAssetKey[plan.AssetKey]; ok {
				state["next_funding_time"] = futures.NextFundingTime
				state["last_funding_rate"] = futures.LastFundingRate
			}
			triggerRate, _ := getFloatParam(plan.Parameters, "trigger_funding_rate")
			triggerBasis, _ := getFloatParam(plan.Parameters, "trigger_basis_pct_max")
			hedgeNotional, _ := getFloatParam(plan.Parameters, "hedge_notional_usd")
			state["trigger_funding_rate"] = triggerRate
			state["trigger_basis_pct_max"] = triggerBasis
			state["hedge_notional_usd"] = hedgeNotional
		case "S18":
			trend := getStringParam(plan.Parameters, "trend_state")
			state["trend_state"] = trend
			state["last_trend_state"] = trend
			state["last_signal_at"] = nil
			state["next_check_at"] = now.Add(24 * time.Hour)
		case "S22":
			threshold, _ := getFloatParam(plan.Parameters, "rebalance_threshold_pct")
			state["last_rebalance_at"] = valuationAsOf
			state["next_rebalance_at"] = now.AddDate(0, 1, 0)
			state["rebalance_threshold_pct"] = threshold
			state["pending_rebalance"] = false
		}
		if len(state) == 0 {
			continue
		}
		states = append(states, PlanState{
			ID:         newID("plan_state"),
			UserID:     userID,
			PlanID:     plan.PlanID,
			StrategyID: plan.StrategyID,
			AssetKey:   plan.AssetKey,
			State:      marshalJSON(state),
			UpdatedAt:  now,
		})
	}
	return states
}

type planStateData struct {
	ID         string
	PlanID     string
	StrategyID string
	AssetKey   string
	State      map[string]any
	Updated    bool
}

func decodePlanStates(rows []PlanState) map[string]*planStateData {
	stateByPlan := make(map[string]*planStateData, len(rows))
	for _, row := range rows {
		state := make(map[string]any)
		if len(row.State) > 0 {
			_ = json.Unmarshal(row.State, &state)
		}
		stateByPlan[row.PlanID] = &planStateData{
			ID:         row.ID,
			PlanID:     row.PlanID,
			StrategyID: row.StrategyID,
			AssetKey:   row.AssetKey,
			State:      state,
		}
	}
	return stateByPlan
}

func (ps *planStateData) getFloat(key string) (float64, bool) {
	if ps == nil || ps.State == nil {
		return 0, false
	}
	value, ok := ps.State[key]
	if !ok {
		return 0, false
	}
	switch v := value.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case json.Number:
		parsed, err := v.Float64()
		return parsed, err == nil
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func (ps *planStateData) getString(key string) string {
	if ps == nil || ps.State == nil {
		return ""
	}
	value, ok := ps.State[key]
	if !ok {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	return ""
}

func (ps *planStateData) getBool(key string) bool {
	if ps == nil || ps.State == nil {
		return false
	}
	value, ok := ps.State[key]
	if !ok {
		return false
	}
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return strings.EqualFold(v, "true")
	default:
		return false
	}
}

func (ps *planStateData) getTime(key string) (time.Time, bool) {
	value := ps.getString(key)
	if value == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}

func (ps *planStateData) set(key string, value any) {
	if ps == nil {
		return
	}
	if ps.State == nil {
		ps.State = make(map[string]any)
	}
	ps.State[key] = value
	ps.Updated = true
}

func (ps *planStateData) remove(key string) {
	if ps == nil || ps.State == nil {
		return
	}
	delete(ps.State, key)
	ps.Updated = true
}
