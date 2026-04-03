package app

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

type planStateUpdate struct {
	ID    string
	State map[string]any
}

func (s *Server) buildPlanStateUpdates(ctx context.Context, userID, sourceBatchID string, deltas []transactionDelta, now time.Time) []planStateUpdate {
	if userID == "" || len(deltas) == 0 {
		return nil
	}

	plans, err := s.loadLatestPaidPlansForUser(ctx, userID)
	if err != nil {
		s.logger.Printf("plan state updates: load plans failed: %v", err)
		return nil
	}
	if len(plans) == 0 {
		return nil
	}

	var planRows []PlanState
	if err := s.db.DB().WithContext(ctx).Where("user_id = ?", userID).Find(&planRows).Error; err != nil {
		s.logger.Printf("plan state updates: load plan states failed: %v", err)
		return nil
	}
	planStates := decodePlanStates(planRows)
	if len(planStates) == 0 {
		return nil
	}

	profile, err := s.ensureUserProfile(ctx, userID)
	if err != nil {
		s.logger.Printf("plan state updates: load user profile failed: %v", err)
		return nil
	}
	timezone := resolvePlanTimezone(profile, s.loadBatchTimezone(ctx, sourceBatchID))

	advancePlanStatesForDeltas(plans, planStates, deltas, now, timezone)

	updates := make([]planStateUpdate, 0)
	for _, state := range planStates {
		if !state.Updated {
			continue
		}
		updates = append(updates, planStateUpdate{
			ID:    state.ID,
			State: state.State,
		})
	}
	return updates
}

func resolvePlanTimezone(profile UserProfile, deviceTimezone string) string {
	timezone := strings.TrimSpace(profile.Timezone)
	if timezone == "" {
		timezone = strings.TrimSpace(deviceTimezone)
	}
	if timezone == "" {
		timezone = "UTC"
	}
	return timezone
}

func advancePlanStatesForDeltas(plans []lockedPlan, planStates map[string]*planStateData, deltas []transactionDelta, now time.Time, timezone string) {
	if len(plans) == 0 || len(planStates) == 0 || len(deltas) == 0 {
		return
	}
	planByAsset := make(map[string]lockedPlan, len(plans))
	planByID := make(map[string]lockedPlan, len(plans))
	for _, plan := range plans {
		planByID[plan.PlanID] = plan
		if plan.AssetKey != "" {
			planByAsset[plan.AssetKey] = plan
		}
	}

	executed := make(map[string]bool)
	for _, delta := range deltas {
		if delta.AssetKey == "" || delta.Amount == 0 {
			continue
		}
		plan, ok := planByAsset[delta.AssetKey]
		if !ok || executed[plan.PlanID] {
			continue
		}
		state := planStates[plan.PlanID]
		if state == nil {
			continue
		}
		if !deltaMatchesPlan(plan, delta) {
			continue
		}
		advancePlanState(plan, state, now, timezone)
		executed[plan.PlanID] = true
	}

	for _, plan := range plans {
		if plan.StrategyID != "S22" {
			continue
		}
		state := planStates[plan.PlanID]
		if state == nil || !state.getBool("pending_rebalance") {
			continue
		}
		trades := planStateRebalanceTrades(state)
		if len(trades) == 0 {
			continue
		}
		if rebalanceTradesExecuted(trades, deltas) {
			advancePlanState(plan, state, now, timezone)
		}
	}
}

func deltaMatchesPlan(plan lockedPlan, delta transactionDelta) bool {
	if delta.Amount == 0 {
		return false
	}
	switch plan.StrategyID {
	case "S02", "S05", "S09":
		return delta.Amount > 0
	case "S03":
		return delta.Amount < 0
	case "S16":
		return true
	case "S18":
		action := getStringParam(plan.Parameters, "trend_action")
		if action == "reduce_exposure" {
			return delta.Amount < 0
		}
		if action == "hold_or_add" {
			return delta.Amount > 0
		}
		return false
	default:
		return false
	}
}

func advancePlanState(plan lockedPlan, state *planStateData, now time.Time, timezone string) {
	if state == nil {
		return
	}
	switch plan.StrategyID {
	case "S02":
		advanceSafetyOrderState(plan, state)
	case "S03":
		state.set("last_signal_at", now.Format(time.RFC3339))
	case "S05":
		frequency := getStringParam(plan.Parameters, "frequency")
		next := nextExecutionAt(timezone, frequency, now)
		if next != "" {
			state.set("next_execution_at", next)
		}
	case "S09":
		advanceAdditionState(plan, state)
	case "S16":
		nextFunding := now.Add(8 * time.Hour)
		state.set("cooldown_until", nextFunding.Format(time.RFC3339))
		state.set("next_funding_time", nextFunding.Format(time.RFC3339))
	case "S18":
		state.set("last_signal_at", now.Format(time.RFC3339))
		state.set("next_check_at", now.Add(24*time.Hour).Format(time.RFC3339))
	case "S22":
		state.set("pending_rebalance", false)
		state.remove("rebalance_trades")
		state.set("last_rebalance_at", now.Format(time.RFC3339))
		state.set("next_rebalance_at", now.AddDate(0, 1, 0).Format(time.RFC3339))
	}
}

func advanceSafetyOrderState(plan lockedPlan, state *planStateData) {
	orders := planSafetyOrders(plan.Parameters)
	if len(orders) == 0 {
		return
	}
	index := planStateIndex(state, "next_safety_order_index")
	nextIndex := index + 1
	if nextIndex >= len(orders) {
		state.remove("next_safety_order_index")
		state.remove("next_safety_order_price")
		state.remove("next_safety_order_amount_usd")
		return
	}
	next := orders[nextIndex]
	price, _ := getFloatParam(next, "trigger_price")
	amountUSD, _ := getFloatParam(next, "order_amount_usd")
	state.set("next_safety_order_index", nextIndex)
	state.set("next_safety_order_price", price)
	state.set("next_safety_order_amount_usd", amountUSD)
}

func advanceAdditionState(plan lockedPlan, state *planStateData) {
	additions := planAdditions(plan.Parameters)
	if len(additions) == 0 {
		return
	}
	index := planStateIndex(state, "next_addition_index")
	nextIndex := index + 1
	if nextIndex >= len(additions) {
		state.remove("next_addition_index")
		state.remove("next_trigger_profit_pct")
		state.remove("next_addition_amount_usd")
		return
	}
	next := additions[nextIndex]
	trigger, _ := getFloatParam(next, "trigger_profit_pct")
	amountUSD, _ := getFloatParam(next, "addition_amount_usd")
	state.set("next_addition_index", nextIndex)
	state.set("next_trigger_profit_pct", trigger)
	state.set("next_addition_amount_usd", amountUSD)
}

func planStateRebalanceTrades(state *planStateData) []rebalanceTrade {
	if state == nil || state.State == nil {
		return nil
	}
	raw, ok := state.State["rebalance_trades"]
	if !ok || raw == nil {
		return nil
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	var trades []rebalanceTrade
	if err := json.Unmarshal(encoded, &trades); err != nil {
		return nil
	}
	return trades
}

func rebalanceTradesExecuted(trades []rebalanceTrade, deltas []transactionDelta) bool {
	if len(trades) == 0 || len(deltas) == 0 {
		return false
	}
	matched := make(map[int]struct{}, len(trades))
	for i, trade := range trades {
		for _, delta := range deltas {
			if delta.AssetKey != trade.AssetKey || delta.Amount == 0 {
				continue
			}
			if strings.EqualFold(trade.Side, "buy") && delta.Amount > 0 {
				matched[i] = struct{}{}
			}
			if strings.EqualFold(trade.Side, "sell") && delta.Amount < 0 {
				matched[i] = struct{}{}
			}
		}
	}
	return len(matched) == len(trades)
}
