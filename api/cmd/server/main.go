package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Holtz777/nuveafinds-api/internal/ai"
	"github.com/Holtz777/nuveafinds-api/internal/config"
	"github.com/Holtz777/nuveafinds-api/internal/handlers"
	"github.com/Holtz777/nuveafinds-api/internal/httpx"
	"github.com/Holtz777/nuveafinds-api/internal/pinterest"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	var pClient *pinterest.Client
	var tokenStore *pinterest.TokenStore

	tokenFn := func() string {
		if pClient == nil {
			return cfg.PinterestAccessToken
		}
		if pClient.IsSandbox() && cfg.PinterestSandboxToken != "" {
			return cfg.PinterestSandboxToken
		}
		return cfg.PinterestAccessToken
	}

	if cfg.PinterestClientID != "" && cfg.PinterestClientSecret != "" && cfg.PinterestRefreshToken != "" {
		tokenStore = pinterest.NewTokenStore(
			cfg.PinterestClientID,
			cfg.PinterestClientSecret,
			cfg.PinterestAccessToken,
			cfg.PinterestRefreshToken,
			"tokens.json",
		)
		tokenStore.Load()

		pClient = pinterest.NewClientWithTokenFunc(tokenFn)
		log.Println("token-store: initialized with auto-refresh (production)")
	} else {
		if cfg.PinterestAccessToken == "" {
			log.Println("WARNING: PINTEREST_ACCESS_TOKEN not set — /pin-upload (AI) works, but Pinterest endpoints will fail")
		}
		pClient = pinterest.NewClientWithTokenFunc(tokenFn)
	}

	pClient.SetSandbox(cfg.PinterestSandbox)
	if cfg.PinterestSandbox {
		log.Println("pinterest: using SANDBOX API (api-sandbox.pinterest.com)")
	}

	deps := &handlers.Deps{
		Config:    cfg,
		AI:        ai.NewClient(cfg.OpenRouterAPIKey, cfg.OpenRouterModel, cfg.OpenRouterReferer, cfg.OpenRouterTitle),
		Pinterest: pClient,
		Tokens:    tokenStore,
	}

	mux := http.NewServeMux()
	handlers.Register(mux, deps)

	var handler http.Handler = mux
	handler = httpx.CORS(cfg.CORSOrigin)(handler)
	handler = httpx.Logger(handler)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 15 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	refresherCtx, cancelRefresher := context.WithCancel(context.Background())

	idleClosed := make(chan struct{})
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		log.Println("shutting down...")
		cancelRefresher()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
		close(idleClosed)
	}()

	if tokenStore != nil {
		go tokenStore.RunRefresher(refresherCtx, &http.Client{Timeout: 30 * time.Second})
	}

	log.Printf("nuveafinds-api listening on :%s (model=%s, sandbox=%v)", cfg.Port, cfg.OpenRouterModel, cfg.PinterestSandbox)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
	<-idleClosed
	log.Println("bye")
}
