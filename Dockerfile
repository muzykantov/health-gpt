FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bot ./cmd/bot

# ------------------------------------------------------

FROM alpine:latest

WORKDIR /app

COPY --from=builder /bot .

COPY config.json .

CMD ["./bot", "-config", "config.json"]