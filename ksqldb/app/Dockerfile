FROM golang:1.21-alpine AS builder

RUN apk add --no-cache gcc musl-dev librdkafka-dev pkgconf cyrus-sasl-dev

WORKDIR /app

COPY . .

RUN go mod tidy && CGO_ENABLED=1 go build -tags dynamic -o ksqldb-demo .

FROM alpine:3.19

RUN apk add --no-cache librdkafka

COPY --from=builder /app/ksqldb-demo /usr/local/bin/

ENTRYPOINT ["ksqldb-demo"]
