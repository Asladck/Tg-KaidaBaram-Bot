FROM golang:1.25.1-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /tgbot ./cmd/tgbot


FROM debian:bookworm-slim

WORKDIR /root/

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /tgbot .
COPY --from=builder /app/.env .

CMD ["./tgbot"]
