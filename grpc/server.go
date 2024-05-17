package main

import (
    "context"
    "log"
    "net"
    "net/http"

    "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"

    pb "example/proto/example"
)

type server struct {
    pb.UnimplementedExampleServiceServer
}

func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
    return &pb.HelloResponse{Message: "Hello " + req.Name}, nil
}

func main() {
    // Start the gRPC server
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }

    s := grpc.NewServer()
    pb.RegisterExampleServiceServer(s, &server{})
    reflection.Register(s)

    go func() {
        log.Printf("server listening at %v", lis.Addr())
        if err := s.Serve(lis); err != nil {
            log.Fatalf("failed to serve: %v", err)
        }
    }()


    // Start the gRPC-Gateway server
    ctx := context.Background()
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()
    mux := runtime.NewServeMux()
    opts := []grpc.DialOption{grpc.WithInsecure()}

    err = pb.RegisterExampleServiceHandlerFromEndpoint(ctx, mux, "localhost:50051", opts)
    if err != nil {
        log.Fatalf("failed to start HTTP gateway: %v", err)
    }

    log.Println("HTTP server listening at :8080")
    if err := http.ListenAndServe(":8080", mux); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}

