FROM golang:1.25 AS builder

WORKDIR /app

COPY backend/services/notifications/go.mod backend/services/notifications/go.sum ./

RUN go mod download

COPY backend/services/notifications .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/notification-service ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/notification-service /app/notification-service

RUN chmod +x /app/notification-service

CMD ["/app/notification-service"]

