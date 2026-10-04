//npm install @bufbuild/protobuf @connectrpc/connect @connectrpc/connect-web

APP_NAME := api-gateway
GATEWAY_BIN := bin/gateway
USERSERVICE_BIN := bin/userservice

PROTO_DIR := proto
GEN_DIR := gen

.PHONY: all
all: proto build


.PHONY: proto
proto:
	protoc \
		-I $(PROTO_DIR) \
		-I . \
		--go_out=$(GEN_DIR) \
		--go_opt=paths=source_relative \
		$(PROTO_DIR)/common/v1/message.proto \
		$(PROTO_DIR)/user/v1/user.proto


.PHONY: build
build:
	mkdir -p bin

	go build \
		-o $(GATEWAY_BIN) \
		./cmd/gateway

	go build \
		-o $(USERSERVICE_BIN) \
		./cmd/userservice


.PHONY: run
run:
	go run ./cmd/gateway


.PHONY: run-user
run-user:
	go run ./cmd/userservice


.PHONY: nats
nats:
	docker compose up -d nats


.PHONY: up
up:
	docker compose up --build


.PHONY: down
down:
	docker compose down


.PHONY: logs
logs:
	docker compose logs -f


.PHONY: logs-gateway
logs-gateway:
	docker compose logs -f gateway


.PHONY: logs-userservice
logs-userservice:
	docker compose logs -f userservice


.PHONY: test
test:
	go test ./...


.PHONY: tidy
tidy:
	go mod tidy


.PHONY: fmt
fmt:
	go fmt ./...


.PHONY: vet
vet:
	go vet ./...


.PHONY: check
check:
	go fmt ./...
	go vet ./...
	go test ./...


.PHONY: clean
clean:
	rm -rf bin
	rm -rf $(GEN_DIR)/common/v1/*.go
	rm -rf $(GEN_DIR)/user/v1/*.go