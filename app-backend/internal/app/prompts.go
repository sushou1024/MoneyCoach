package app

// Prompts stores system prompts for LLM calls.
type Prompts struct {
	OCRPortfolio     string
	PreviewReport    string
	PaidReport       string
	PaidReportDirect string
	AssetCommand     string
	TradeSlipOCR     string
	DailyBriefing    string
}
