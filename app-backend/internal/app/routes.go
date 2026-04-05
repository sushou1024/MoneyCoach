package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(s.withRequestID)
	r.Use(s.withRecovery)
	r.Use(s.withLogging)
	r.Use(s.withCORS)
	r.Use(s.withAuth)

	r.Get("/healthz", s.handleHealth)

	r.Route("/v1", func(r chi.Router) {
		r.Use(s.withRateLimit(60, 10))

		if s.cfg.EnableDangerousResetRoutes {
			r.Get("/reset-app-state", s.handleResetAppState)
			r.Post("/debug/reset-user", s.handleResetUserByEmail)
		}

		if s.cfg.ObjectStorageMode == "local" {
			r.Put("/local-uploads/*", s.handleLocalUpload)
		}

		r.Route("/auth", func(r chi.Router) {
			r.Post("/oauth", s.handleAuthOAuth)
			r.Post("/email/register/start", s.handleAuthEmailRegisterStart)
			r.Post("/email/register", s.handleAuthEmailRegister)
			r.Post("/email/login", s.handleAuthEmailLogin)
			r.Post("/refresh", s.handleAuthRefresh)
			r.Post("/logout", s.handleAuthLogout)
		})

		r.Group(func(r chi.Router) {
			r.Use(s.requireAuth)

			r.Route("/users", func(r chi.Router) {
				r.Get("/me", s.handleUsersMe)
				r.Patch("/me", s.handleUsersMeUpdate)
				r.With(s.withIdempotency).Delete("/me", s.handleUsersMeDelete)
			})

			r.Route("/devices", func(r chi.Router) {
				r.With(s.withIdempotency).Post("/register", s.handleDeviceRegister)
				r.Delete("/{device_id}", s.handleDeviceDelete)
			})

			r.Route("/upload-batches", func(r chi.Router) {
				r.With(s.withIdempotency).Post("/", s.handleUploadBatchCreate)
				r.With(s.withIdempotency).Post("/{upload_batch_id}/complete", s.handleUploadBatchComplete)
				r.Get("/{upload_batch_id}", s.handleUploadBatchGet)
				r.With(s.withIdempotency).Post("/{upload_batch_id}/review", s.handleUploadBatchReview)
			})

			r.Route("/portfolio", func(r chi.Router) {
				r.Get("/active", s.handlePortfolioActive)
				r.With(s.withIdempotency).Post("/active/refresh", s.handlePortfolioActiveRefresh)
				r.Get("/snapshots/{portfolio_snapshot_id}", s.handlePortfolioSnapshot)
				r.Get("/snapshots", s.handlePortfolioSnapshots)
			})

			r.Route("/reports", func(r chi.Router) {
				r.Get("/preview/{calculation_id}", s.handleReportPreview)
				r.Post("/active", s.handleReportActive)
				r.Post("/{calculation_id}/paid", s.handleReportPaid)
				r.Get("/{calculation_id}", s.handleReportByID)
				r.Get("/", s.handleReportList)
				r.Get("/{calculation_id}/plans/{plan_id}", s.handleReportPlan)
			})

			r.Route("/billing", func(r chi.Router) {
				r.Get("/plans", s.handleBillingPlans)
				r.Get("/entitlement", s.handleBillingEntitlement)
				r.With(s.withIdempotency).Post("/receipt/ios", s.handleBillingReceiptIOS)
				r.With(s.withIdempotency).Post("/receipt/android", s.handleBillingReceiptAndroid)
				r.With(s.withIdempotency).Post("/stripe/session", s.handleBillingStripeSession)
				r.Post("/dev/entitlement", s.handleBillingDevEntitlement)
				r.Delete("/dev/entitlement", s.handleBillingDevEntitlementDelete)
			})

			r.Route("/briefings", func(r chi.Router) {
				r.Get("/today", s.handleBriefingsToday)
			})

			r.Route("/insights", func(r chi.Router) {
				r.Get("/", s.handleInsightsList)
				r.With(s.withIdempotency).Post("/{insight_id}/execute", s.handleInsightExecute)
				r.With(s.withIdempotency).Post("/{insight_id}/dismiss", s.handleInsightDismiss)
			})

			r.Route("/intelligence", func(r chi.Router) {
				r.Get("/regime", s.handleIntelligenceRegime)
				r.Get("/assets/{asset_key}", s.handleIntelligenceAssetBrief)
			})

			r.Route("/market-data", func(r chi.Router) {
				r.Get("/ohlcv", s.handleMarketDataOHLCV)
			})

			r.Route("/assets", func(r chi.Router) {
				r.Get("/lookup", s.handleAssetsLookup)
				r.With(s.withIdempotency).Post("/commands", s.handleAssetsCommand)
			})

			r.With(s.withIdempotency).Post("/waitlist", s.handleWaitlist)
		})

		r.Route("/webhooks", func(r chi.Router) {
			r.Post("/apple", s.handleWebhookApple)
			r.Post("/google", s.handleWebhookGoogle)
			r.Post("/stripe", s.handleWebhookStripe)
		})
	})

	return r
}
