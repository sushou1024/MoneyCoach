package app

type ocrPromptResponse struct {
	Images []ocrImage `json:"images"`
}

type ocrImage struct {
	ImageID       string     `json:"image_id"`
	Status        string     `json:"status"`
	ErrorReason   *string    `json:"error_reason"`
	PlatformGuess string     `json:"platform_guess"`
	Assets        []ocrAsset `json:"assets"`
}

type ocrAsset struct {
	SymbolRaw           string   `json:"symbol_raw"`
	Symbol              *string  `json:"symbol"`
	AssetType           string   `json:"asset_type"`
	Amount              float64  `json:"amount"`
	ValueFromScreenshot *float64 `json:"value_from_screenshot"`
	DisplayCurrency     *string  `json:"display_currency"`
	PNLPercent          *float64 `json:"pnl_percent"`
	AvgPrice            *float64 `json:"avg_price"`
}

type tradeSlipOCRResponse struct {
	ImageID string           `json:"image_id"`
	Trades  []tradeSlipTrade `json:"trades"`
}

type tradeSlipTrade struct {
	Side       string         `json:"side"`
	Symbol     string         `json:"symbol"`
	Amount     *float64       `json:"amount"`
	Price      *float64       `json:"price"`
	Currency   *string        `json:"currency"`
	ExecutedAt *string        `json:"executed_at"`
	Fees       *tradeSlipFees `json:"fees"`
}

type tradeSlipFees struct {
	Amount   *float64 `json:"amount"`
	Currency *string  `json:"currency"`
}

type assetCommandResponse struct {
	Intent   string                `json:"intent"`
	Payloads []assetCommandPayload `json:"payloads"`
}

type assetCommandPayload struct {
	TargetAsset   *assetCommandTarget  `json:"target_asset"`
	FundingSource *assetCommandFunding `json:"funding_source"`
	PricePerUnit  *float64             `json:"price_per_unit"`
}

type assetCommandTarget struct {
	Ticker string   `json:"ticker"`
	Amount *float64 `json:"amount"`
	Action string   `json:"action"`
}

type assetCommandFunding struct {
	Ticker     *string  `json:"ticker"`
	Amount     *float64 `json:"amount"`
	IsExplicit bool     `json:"is_explicit"`
}
