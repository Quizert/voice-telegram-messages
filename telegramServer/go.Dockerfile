# syntax=docker/dockerfile:1

# ---------- СТАДИЯ СБОРКИ ----------
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Собираем бинарник из cmd/
RUN go build -o telegram_bot ./cmd

# ---------- СТАДИЯ ЗАПУСКА ----------
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/telegram_bot .

CMD ["./telegram_bot"]
