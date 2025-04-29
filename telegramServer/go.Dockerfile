
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o telegram_bot ./cmd

FROM alpine:latest
RUN apk add --no-cache ffmpeg

WORKDIR /app
COPY --from=builder /app/telegram_bot .

CMD ["./telegram_bot"]
