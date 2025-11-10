FROM golang:1.25 AS builder

WORKDIR /app

COPY backend/services/users/go.mod backend/services/users/go.sum ./

COPY backend/proto /proto

RUN go mod download

COPY backend/services/users .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/users-service ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/users-service /app/users-service

RUN chmod +x /app/users-service

EXPOSE 50051

CMD ["/app/users-service"]