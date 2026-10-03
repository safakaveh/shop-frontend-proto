package main

import (
	"log"
	"os"
	"strconv"
	"sync"

	ordersv1 "example.com/shop/backend/gen/go/orders/v1"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

func main() {
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = "nats://127.0.0.1:4222"
	}

	nc, err := nats.Connect(url, nats.Name("orders-worker"))
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	// Queue group: چند instance افقی scale می‌شوند و هر پیام به یکی می‌رسد.
	_, err = nc.QueueSubscribe("orders.v1.*", "orders-workers", handleMessage)
	if err != nil {
		log.Fatal(err)
	}
	if err := nc.Flush(); err != nil {
		log.Fatal(err)
	}

	log.Println("orders worker ready")
	select {}
}

func handleMessage(msg *nats.Msg) {
	switch msg.Subject {
	case "orders.v1.create":
		handleCreateOrder(msg)
	case "orders.v1.get":
		handleGetOrder(msg)
	default:
		respondProblem(msg, 404, "unknown_subject")
	}
}

func handleCreateOrder(msg *nats.Msg) {
	var req ordersv1.CreateOrderRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		respondProblem(msg, 400, "invalid_protobuf")
		return
	}

	if req.GetUserId() == "" || req.GetProductId() == "" || req.GetQuantity() == 0 {
		respondProblem(msg, 422, "validation_failed")
		return
	}

	key := msg.Header.Get("Idempotency-Key")
	if key == "" {
		respondProblem(msg, 400, "missing_idempotency_key")
		return
	}

	// TODO: جایگزین با تراکنش دیتابیس + قید UNIQUE روی کلید idempotency.
	// تکرار همان کلید باید همان پاسخ قبلی را برگرداند.
	orderID := "order-" + key

	data, err := proto.Marshal(&ordersv1.CreateOrderResponse{
		OrderId: orderID,
		Status:  "created",
	})
	if err != nil {
		respondProblem(msg, 500, "encode_failed")
		return
	}

	respond(msg, 201, "application/protobuf", data)
}

func handleGetOrder(msg *nats.Msg) {
	var req ordersv1.GetOrderRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		respondProblem(msg, 400, "invalid_protobuf")
		return
	}

	// TODO: خواندن واقعی از دیتابیس/سرویس سفارش.
	data, err := proto.Marshal(&ordersv1.GetOrderResponse{
		OrderId:   req.GetOrderId(),
		ProductId: "example-product",
		Quantity:  1,
		Status:    "created",
	})
	if err != nil {
		respondProblem(msg, 500, "encode_failed")
		return
	}

	respond(msg, 200, "application/protobuf", data)
}

func respond(msg *nats.Msg, status int, contentType string, data []byte) {
	if msg.Reply == "" {
		return
	}
	reply := nats.NewMsg(msg.Reply)
	reply.Data = data
	reply.Header.Set("Gateway-HTTP-Status", strconv.Itoa(status))
	reply.Header.Set("Content-Type", contentType)
	_ = msg.RespondMsg(reply)
}

func respondProblem(msg *nats.Msg, status int, code string) {
	respond(msg, status, "application/problem+json", []byte(`{"code":"`+code+`"}`))
}

// جلوگیری از خطای import بلااستفاده اگر sync را حذف کردی.
var _ = sync.Mutex{}
