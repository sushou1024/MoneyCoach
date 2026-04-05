package app

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	defaultHTTPTimeout = 20 * time.Second
	geminiHTTPTimeout  = 90 * time.Second
)

type Server struct {
	cfg            Config
	db             *dbStore
	redis          *redisStore
	storage        storageService
	gemini         *geminiClient
	market         *marketClient
	mailer         *resendClient
	auth           *authService
	queue          *jobQueue
	logos          *logoResolver
	logger         *log.Logger
	httpClient     *http.Client
	prompts        Prompts
	appleIDKeys    appleKeyCache
	appleStoreKeys appleKeyCache
	apnsKeyOnce    sync.Once
	apnsKey        *ecdsa.PrivateKey
	apnsKeyErr     error
}

// NewServer wires dependencies and returns a configured Server.
func NewServer(ctx context.Context, cfg Config, prompts Prompts) (*Server, error) {
	db, err := openDatabase(ctx, cfg)
	if err != nil {
		return nil, err
	}

	redisStore, err := newRedisStore(cfg)
	if err != nil {
		return nil, err
	}

	storage, err := newStorageClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: defaultHTTPTimeout}
	geminiHTTPClient := &http.Client{Timeout: geminiHTTPTimeout}

	mailer := newResendClient(cfg.ResendAPIKey, cfg.ResendFromEmail, client)
	gemini := newGeminiClient(cfg.GeminiAPIKey, geminiHTTPClient)
	marketCache := newMarketCacheStore(db)
	logger := log.Default()
	market := newMarketClient(cfg, client, redisStore, marketCache, logger)
	auth := newAuthService(cfg, db, redisStore)
	queue := newJobQueue(redisStore)
	logos := newLogoResolver(cfg, client, logger)

	return &Server{
		cfg:        cfg,
		db:         &dbStore{db: db},
		redis:      redisStore,
		storage:    storage,
		gemini:     gemini,
		market:     market,
		mailer:     mailer,
		auth:       auth,
		queue:      queue,
		logos:      logos,
		logger:     logger,
		httpClient: client,
		prompts:    prompts,
	}, nil
}

// Start runs HTTP server and background workers.
func (s *Server) Start(ctx context.Context) error {
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	s.queue.startWorkers(workerCtx, s)
	s.startInsightsScheduler(workerCtx)
	s.startBriefingScheduler(workerCtx)

	router := s.routes()
	srv := &http.Server{
		Addr:              ":" + s.cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("listen: %w", err)
	}
	return nil
}
