package gateway

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"

	ordersv1 "example.com/shop/backend/gen/go/orders/v1"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

func (g *Gateway) createOrder(w http.ResponseWriter, r *http.Request) {
	if !g.acquire(w) {
		return
	}
	defer g.release()

	if !hasMediaType(r.Header.Get("Content-Type"), protobufContentType) {
		writeProblem(w, http.StatusUnsupportedMediaType, "expected_protobuf")
		return
	}

	key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if key == "" || len(key) > 128 {
		writeProblem(w, http.StatusBadRequest, "invalid_idempotency_key")
		return
	}

	body, err := g.readLimitedBody(w, r)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeProblem(w, http.StatusRequestEntityTooLarge, "request_too_large")
			return
		}
		writeProblem(w, http.StatusBadRequest, "invalid_request_body")
		return
	}
	if len(body) == 0 {
		writeProblem(w, http.StatusBadRequest, "empty_request_body")
		return
	}

	// pass-through: بدنه Decode نمی‌شود.
	msg := nats.NewMsg("orders.v1.create")
	msg.Data = body
	msg.Header.Set("Content-Type", protobufContentType)
	msg.Header.Set("Idempotency-Key", key)
	msg.Header.Set("X-Request-ID", middleware.GetReqID(r.Context()))

	g.requestAndWrite(w, r, msg)
}

func (g *Gateway) getOrder(w http.ResponseWriter, r *http.Request) {
	if !g.acquire(w) {
		return
	}
	defer g.release()

	orderID := strings.TrimSpace(chiURLParam(r, "orderID"))
	if orderID == "" || len(orderID) > 128 {
		writeProblem(w, http.StatusBadRequest, "invalid_order_id")
		return
	}

	body, err := proto.Marshal(&ordersv1.GetOrderRequest{OrderId: orderID})
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "encode_failed")
		return
	}

	msg := nats.NewMsg("orders.v1.get")
	msg.Data = body
	msg.Header.Set("Content-Type", protobufContentType)
	msg.Header.Set("X-Request-ID", middleware.GetReqID(r.Context()))

	g.requestAndWrite(w, r, msg)
}

func (g *Gateway) readLimitedBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	r.Body = http.MaxBytesReader(w, r.Body, g.cfg.MaxBodyBytes)
	defer r.Body.Close()
	return io.ReadAll(r.Body)
}

func hasMediaType(value, expected string) bool {
	actual, _, err := mime.ParseMediaType(value)
	return err == nil && actual == expected
}

func parseStatus(raw string) (int, bool) {
	n, err := strconv.Atoi(raw)
	if err != nil || n < 200 || n > 599 {
		return 0, false
	}
	return n, true
}

func writeProblem(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"code":   code,
		"detail": http.StatusText(status),
	})
}
