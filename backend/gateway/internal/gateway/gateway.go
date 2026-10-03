package gateway

import (
	"context"
	"errors"
	"net/http"

	"example.com/shop/backend/gateway/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"
)

const protobufContentType = "application/protobuf"

type Gateway struct {
	nc       *nats.Conn
	cfg      config.Config
	inflight chan struct{}
}

func New(nc *nats.Conn, cfg config.Config) (*Gateway, error) {
	if nc == nil {
		return nil, errors.New("nats connection is nil")
	}
	return &Gateway{
		nc:       nc,
		cfg:      cfg,
		inflight: make(chan struct{}, cfg.MaxInflight),
	}, nil
}

func (g *Gateway) Handler() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	r.Get("/livez", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Get("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		if !g.nc.IsConnected() {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// فقط server-to-server؛ فرانت مستقیم مرورگر به این مسیرها وصل نمی‌شود.
	r.Group(func(r chi.Router) {
		r.Use(g.requireServiceToken)
		r.Post("/internal/v1/orders", g.createOrder)
		r.Get("/internal/v1/orders/{orderID}", g.getOrder)
	})

	return r
}

func (g *Gateway) acquire(w http.ResponseWriter) bool {
	select {
	case g.inflight <- struct{}{}:
		return true
	default:
		writeProblem(w, http.StatusServiceUnavailable, "gateway_overloaded")
		return false
	}
}

func (g *Gateway) release() {
	<-g.inflight
}

// requestAndWrite پیام را به NATS می‌فرستد و پاسخ Worker را مستقیم برمی‌گرداند.
// گیت‌وی بدنهٔ Protobuf را Decode نمی‌کند (pass-through).
func (g *Gateway) requestAndWrite(
	w http.ResponseWriter,
	r *http.Request,
	msg *nats.Msg,
) {
	ctx, cancel := context.WithTimeout(r.Context(), g.cfg.ReqTimeout)
	defer cancel()

	reply, err := g.nc.RequestMsgWithContext(ctx, msg)
	if err != nil {
		switch {
		case errors.Is(err, nats.ErrNoResponders):
			writeProblem(w, http.StatusServiceUnavailable, "service_unavailable")
		case errors.Is(err, nats.ErrTimeout),
			errors.Is(ctx.Err(), context.DeadlineExceeded):
			writeProblem(w, http.StatusGatewayTimeout, "upstream_timeout")
		default:
			writeProblem(w, http.StatusBadGateway, "upstream_error")
		}
		return
	}

	if len(reply.Data) > g.cfg.MaxBodyBytes {
		writeProblem(w, http.StatusBadGateway, "upstream_response_too_large")
		return
	}

	status := http.StatusOK
	if raw := reply.Header.Get("Gateway-HTTP-Status"); raw != "" {
		parsed, ok := parseStatus(raw)
		if !ok {
			writeProblem(w, http.StatusBadGateway, "invalid_upstream_status")
			return
		}
		status = parsed
	}

	contentType := reply.Header.Get("Content-Type")
	switch contentType {
	case protobufContentType, "application/problem+json":
		// فقط Content-Typeهای شناخته‌شده مجازند.
	default:
		writeProblem(w, http.StatusBadGateway, "invalid_upstream_content_type")
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-Request-ID", middleware.GetReqID(r.Context()))
	w.WriteHeader(status)
	_, _ = w.Write(reply.Data)
}
