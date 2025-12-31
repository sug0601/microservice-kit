FROM golang:1.23-alpine AS builder

RUN apk add --no-cache protobuf protobuf-dev

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

WORKDIR /app

# Copy proto files first
COPY proto/ ./proto/

# Generate gRPC code
RUN protoc --go_out=. --go_opt=module=grpc-server \
           --go-grpc_out=. --go-grpc_opt=module=grpc-server \
           proto/hello.proto

# Copy go.mod and source
COPY server/go.mod ./
COPY server/main.go ./

# Update dependencies and build
RUN go mod tidy && go build -o grpc-server .

FROM alpine:3.19

WORKDIR /app
COPY --from=builder /app/grpc-server .
COPY --from=builder /app/proto ./proto

EXPOSE 50051

CMD ["./grpc-server"]
