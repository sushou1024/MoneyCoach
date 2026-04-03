package main

import _ "embed"

//go:embed system-prompts/ocr-portfolio.txt
var OCRPortfolioPrompt string

//go:embed system-prompts/preview-report.txt
var PreviewReportPrompt string

//go:embed system-prompts/paid-report.txt
var PaidReportPrompt string

//go:embed system-prompts/paid-report-direct.txt
var PaidReportDirectPrompt string

//go:embed system-prompts/asset-command.txt
var AssetCommandPrompt string

//go:embed system-prompts/trade-slip-ocr.txt
var TradeSlipOCRPrompt string
