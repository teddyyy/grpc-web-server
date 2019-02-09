package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	pb "github.com/teddyyy/grpc-web-server/helloworld"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":9090"
	url  = "http://metadata.google.internal/computeMetadata/v1/instance/zone"
)

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	region, _ := getRegionFromMetadata()

	return &pb.HelloReply{
		Message: "Hello " + in.Name,
		Region:  region,
	}, nil
}

func parseResponseBody(body string) string {
	bodyArray := strings.Split(body, "/")
	zone := bodyArray[3]

	regionArray := strings.Split(zone, "-")
	region := regionArray[0] + "-" + regionArray[1]

	return region
}

func getRegionFromMetadata() (string, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Metadata-Flavor", "Google")

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect metadata", err)
		return "", nil
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read response body", err)
		return "", nil
	}

	var region = parseResponseBody(string(body))

	return region, nil
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
