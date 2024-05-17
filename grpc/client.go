package main

import (
    "context"
    "log"
    "time"

    "google.golang.org/grpc"
    pb "example/proto"
)

func main() {
    // Set up a connection to the server.
    conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
    if err != nil {
        log.Fatalf("did not connect: %v", err)
    }
    defer conn.Close()
    client := pb.NewExampleServiceClient(conn)

    // Contact the server and print out its response.
    name := "world"
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    r, err := client.SayHello(ctx, &pb.HelloRequest{Name: name})
    if err != nil {
        log.Fatalf("could not greet: %v", err)
    }
    log.Printf("Greeting: %s", r.GetMessage())
}

