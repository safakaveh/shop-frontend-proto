# syntax=docker/dockerfile:1

FROM golang:1.27.1-alpine3.24 AS builder


ARG SERVICE=api-getway
ARG BINARY_NAME=app
ARG BUILD_TARGET=./cmd/api-gateway

ARG GOPROXY=https://goproxy.cn,direct
ARG GOSUMDB=sum.golang.org

ENV GOPROXY=${GOPROXY}
ENV GOSUMDB=${GOSUMDB}

WORKDIR /src
RUN apk add --no-cache \
    ca-certificates \
    git

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/gateway \
    ./cmd/gateway

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/userservice \
    ./cmd/userservice


FROM alpine:3.22

RUN apk add --no-cache \
    ca-certificates \
    tzdata

WORKDIR /app

COPY --from=builder /out/gateway /app/gateway
COPY --from=builder /out/userservice /app/userservice

RUN addgroup -S app \
    && adduser -S app -G app

USER app

EXPOSE 8080

CMD ["/app/gateway"]