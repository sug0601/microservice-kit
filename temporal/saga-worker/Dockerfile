FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY saga-worker/go.mod saga-worker/go.sum ./
RUN go mod download

COPY saga-worker/*.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o saga-worker .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/saga-worker .

CMD ["./saga-worker"]
