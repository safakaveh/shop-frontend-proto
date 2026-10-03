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

	"example.com/shop/backend/gateway/internal/config"
	"example.com/shop/backend/gateway/internal/gateway"

	"github.com/nats-io/nats.go"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	nc, err := nats.Connect(
		cfg.NATSURL,
		nats.Name("orders-api-gateway"),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(500*time.Millisecond),
		nats.ReconnectJitter(100*time.Millisecond, 2*time.Second),
	)
	if err != nil {
		log.Fatalf("connect to NATS: %v", err)
	}
	defer nc.Close()

	gw, err := gateway.New(nc, cfg)
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           gw.Handler(),
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	log.Printf("gateway listening on %s", cfg.HTTPAddr)

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
