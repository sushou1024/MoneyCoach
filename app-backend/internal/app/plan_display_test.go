package app

import "testing"

func TestConvertLockedPlansToDisplayByPlanUsesPlanCurrency(t *testing.T) {
	holdingByAsset := map[string]*portfolioHolding{
		"stock:mic:XNAS:TSLA": {
			AssetKey:      "stock:mic:XNAS:TSLA",
			AssetType:     "stock",
			QuoteCurrency: "USD",
			FXRateToUSD:   1,
		},
		"stock:mic:XHKG:3681.HK": {
			AssetKey:      "stock:mic:XHKG:3681.HK",
			AssetType:     "stock",
			QuoteCurrency: "HKD",
			FXRateToUSD:   0.128,
		},
	}

	plans := []lockedPlan{
		{
			PlanID:     "plan_us",
			StrategyID: "S09",
			AssetType:  "stock",
			Symbol:     "TSLA",
			AssetKey:   "stock:mic:XNAS:TSLA",
			Parameters: map[string]any{
				"base_addition_usd": 100.0,
				"additions": []any{
					map[string]any{"addition_amount_usd": 125.0},
				},
			},
		},
		{
			PlanID:     "plan_hk",
			StrategyID: "S04",
			AssetType:  "stock",
			Symbol:     "3681.HK",
			AssetKey:   "stock:mic:XHKG:3681.HK",
			Parameters: map[string]any{
				"layers": []any{
					map[string]any{
						"target_price":        10.0,
						"expected_profit_usd": 20.0,
					},
				},
			},
		},
		{
			PlanID:     "plan_pf",
			StrategyID: "S05",
			AssetType:  "portfolio",
			Symbol:     "PORTFOLIO",
			AssetKey:   "portfolio:pf_123",
			Parameters: map[string]any{
				"amount": 100.0,
			},
		},
	}

	got := convertLockedPlansToDisplayByPlan(plans, holdingByAsset, "CNY", 7.2)
	if len(got) != 3 {
		t.Fatalf("expected 3 converted plans, got %d", len(got))
	}

	if got[0].QuoteCurrency != "USD" {
		t.Fatalf("expected TSLA quote currency USD, got %q", got[0].QuoteCurrency)
	}
	if value := got[0].Parameters["base_addition_usd"]; value != 100.0 {
		t.Fatalf("expected TSLA addition to stay in USD, got %#v", value)
	}
	additions, ok := got[0].Parameters["additions"].([]any)
	if !ok || len(additions) != 1 {
		t.Fatalf("expected one TSLA addition, got %#v", got[0].Parameters["additions"])
	}
	firstAddition, ok := additions[0].(map[string]any)
	if !ok || firstAddition["addition_amount_usd"] != 125.0 {
		t.Fatalf("expected TSLA ladder to stay in USD, got %#v", additions[0])
	}

	if got[1].QuoteCurrency != "HKD" {
		t.Fatalf("expected HK stock quote currency HKD, got %q", got[1].QuoteCurrency)
	}
	layers, ok := got[1].Parameters["layers"].([]any)
	if !ok || len(layers) != 1 {
		t.Fatalf("expected one HK layer, got %#v", got[1].Parameters["layers"])
	}
	firstLayer, ok := layers[0].(map[string]any)
	if !ok {
		t.Fatalf("expected HK layer object, got %#v", layers[0])
	}
	if firstLayer["target_price"] != 78.13 {
		t.Fatalf("expected HK target_price in HKD, got %#v", firstLayer["target_price"])
	}
	if firstLayer["expected_profit_usd"] != 156.25 {
		t.Fatalf("expected HK expected_profit in HKD display units, got %#v", firstLayer["expected_profit_usd"])
	}

	if got[2].QuoteCurrency != "CNY" {
		t.Fatalf("expected portfolio quote currency/base currency CNY, got %q", got[2].QuoteCurrency)
	}
	if got[2].Parameters["amount"] != 720.0 {
		t.Fatalf("expected portfolio amount in base currency, got %#v", got[2].Parameters["amount"])
	}
}
