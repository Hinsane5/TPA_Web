FROM golang:1.25 AS builder

WORKDIR /app

COPY backend/api-gateway/go.mod backend/api-gateway/go.sum ./

COPY backend/proto /proto

RUN go mod download

COPY backend/api-gateway .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/api-gateway ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/api-gateway /app/api-gateway

RUN chmod +x /app/api-gateway

EXPOSE 8080

CMD ["/app/api-gateway"]
