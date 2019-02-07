package main

import (
    pb "github.com/teddyyy/grpc-web-server/helloworld"
    "log"
    "net"
    "os"

    "golang.org/x/net/context"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
)

const (
    port = ":9090"
)

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
    if os.Getenv("REGION") == "" {
        os.Setenv("REGION", "???")
    }
    var region = os.Getenv("REGION")

    return &pb.HelloReply{
        Message: "Hello " + in.Name,
        Region: region,
    }, nil
}

func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
    res, err := handler(ctx, req)
	log.Printf("%s: %v -> %v", info.FullMethod, req, res)
	return res, err
}

func main() {
    log.Print("Hello gRPC Web Server...")
    lis, err := net.Listen("tcp", port)
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    s := grpc.NewServer(
        grpc.UnaryInterceptor(loggingInterceptor),
    )

    pb.RegisterGreeterServer(s, &server{})
    // Register reflection service on gRPC server.
    reflection.Register(s)
    if err := s.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}