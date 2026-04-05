package app

import (
	"context"
	"fmt"
)

const (
	jobOCRHoldings     = "ocr_holdings"
	jobOCRTradeSlip    = "ocr_trade_slip"
	jobNormalize       = "normalize_holdings"
	jobPreviewReport   = "preview_report"
	jobPaidReport      = "paid_report"
	jobInsightsRefresh = "insights_refresh"
	jobDailyBriefing   = "daily_briefing"
)

func (s *Server) processJob(ctx context.Context, job jobPayload) error {
	switch job.Type {
	case jobOCRHoldings:
		return s.processOCRHoldings(ctx, job.ID)
	case jobOCRTradeSlip:
		return s.processOCRTradeSlip(ctx, job.ID)
	case jobNormalize:
		return s.processNormalization(ctx, job.ID)
	case jobPreviewReport:
		return s.processPreviewReport(ctx, job.ID)
	case jobPaidReport:
		return s.processPaidReport(ctx, job.ID)
	case jobInsightsRefresh:
		return s.processInsightsRefresh(ctx, job.ID)
	case jobDailyBriefing:
		return s.processDailyBriefing(ctx, job.ID)
	default:
		return fmt.Errorf("unknown job type %q", job.Type)
	}
}
