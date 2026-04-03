package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackcpku/moneycoach/app-backend/internal/app"
)

// flag

func main() {
	cfg, err := app.LoadConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	prompts := app.Prompts{
		OCRPortfolio:     OCRPortfolioPrompt,
		PreviewReport:    PreviewReportPrompt,
		PaidReport:       PaidReportPrompt,
		PaidReportDirect: PaidReportDirectPrompt,
		AssetCommand:     AssetCommandPrompt,
		TradeSlipOCR:     TradeSlipOCRPrompt,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server, err := app.NewServer(ctx, cfg, prompts)
	if err != nil {
		log.Fatalf("server init error: %v", err)
	}

	log.Printf("listening on :%s", cfg.Port)
	if err := server.Start(ctx); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
