package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "grpc-server/pb"
)

type helloServer struct {
	pb.UnimplementedHelloServiceServer
}

func (s *helloServer) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	log.Printf("Received SayHello request: name=%s", req.Name)
	return &pb.HelloResponse{
		Message: fmt.Sprintf("Hello, %s! (from gRPC server via Kong)", req.Name),
	}, nil
}

func (s *helloServer) SayHelloServerStream(req *pb.HelloRequest, stream pb.HelloService_SayHelloServerStreamServer) error {
	log.Printf("Received SayHelloServerStream request: name=%s", req.Name)
	for i := 1; i <= 5; i++ {
		msg := &pb.HelloResponse{
			Message: fmt.Sprintf("Hello %s! Message %d of 5", req.Name, i),
		}
		if err := stream.Send(msg); err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil
}

func main() {
	port := ":50051"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterHelloServiceServer(server, &helloServer{})

	// Enable reflection for grpcurl
	reflection.Register(server)

	log.Printf("gRPC server starting on %s", port)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
