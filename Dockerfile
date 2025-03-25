FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/bot ./cmd/bot

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/bot /app/bot
COPY migrations /app/migrations

RUN apk --no-cache add postgresql-client

CMD ["/app/bot"]