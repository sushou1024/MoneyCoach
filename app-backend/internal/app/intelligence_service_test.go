package app

import "testing"

func TestSelectDistinctTopMovers(t *testing.T) {
	candidates := []intelligenceAssetLeader{
		{AssetKey: "stock:mic:XNYS:CRCL", Symbol: "CRCL", Change30d: 0.45},
		{AssetKey: "stock:mic:XNAS:TSLA", Symbol: "TSLA", Change30d: -0.06},
		{AssetKey: "crypto:cg:bitcoin", Symbol: "BTC", Change30d: -0.076},
		{AssetKey: "stock:mic:XNAS:NVDA", Symbol: "NVDA", Change30d: 0.12},
	}

	leaders, laggards := selectDistinctTopMovers(candidates, 2)

	if len(leaders) != 2 {
		t.Fatalf("leaders len = %d, want 2", len(leaders))
	}
	if leaders[0].Symbol != "CRCL" || leaders[1].Symbol != "NVDA" {
		t.Fatalf("leaders = %#v, want CRCL then NVDA", leaders)
	}

	if len(laggards) != 2 {
		t.Fatalf("laggards len = %d, want 2", len(laggards))
	}
	if laggards[0].Symbol != "BTC" || laggards[1].Symbol != "TSLA" {
		t.Fatalf("laggards = %#v, want BTC then TSLA", laggards)
	}
}

func TestSelectDistinctTopMoversSkipsNonPositiveLeaderAndNonNegativeLaggard(t *testing.T) {
	candidates := []intelligenceAssetLeader{
		{AssetKey: "stock:mic:XNAS:TSLA", Symbol: "TSLA", Change30d: -0.06},
		{AssetKey: "crypto:cg:bitcoin", Symbol: "BTC", Change30d: -0.076},
	}

	leaders, laggards := selectDistinctTopMovers(candidates, 2)

	if len(leaders) != 0 {
		t.Fatalf("leaders len = %d, want 0", len(leaders))
	}
	if len(laggards) != 2 {
		t.Fatalf("laggards len = %d, want 2", len(laggards))
	}
	if laggards[0].Symbol != "BTC" || laggards[1].Symbol != "TSLA" {
		t.Fatalf("laggards = %#v, want BTC then TSLA", laggards)
	}
}
