# syntax=docker/dockerfile:1

# ---------- СТАДИЯ СБОРКИ ----------
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o telegram_bot main.go

# ---------- СТАДИЯ ЗАПУСКА ----------
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/telegram_bot .

# передаём адрес gRPC сервера через ENV
ENV GRPC_SERVER_HOST=tts-server
ENV GRPC_SERVER_PORT=50051

CMD ["./telegram_bot"]
