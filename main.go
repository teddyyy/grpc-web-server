package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"database/sql"
	"time"

	pb "github.com/teddyyy/grpc-web-server/helloworld"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	_ "github.com/go-sql-driver/mysql"
)

const (
	port = ":9090"
	url  = "http://metadata.google.internal/computeMetadata/v1/instance/zone"
)

// server is used to implement helloworld.GreeterServer.
type server struct{
	db *sql.DB
}

func executeDB(db *sql.DB, name string) error {
	_, err := db.Exec(`INSERT INTO demo_history (message, timestamp) VALUES (?, ?) `, name, time.Now())
	if err != nil {
		log.Printf("failed to insert mysql %v", err)
		return err
	}

	return err
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	var region string
	env := os.Getenv("ENV")

	if len(env) == 0 {
		region, _ = getRegionFromMetadata()
	} else if env == "local" {
		region = "localhost"
	} else {
		region, _ = getRegionFromMetadata()
	}

	err := executeDB(s.db, in.Name)
	if err != nil {
		log.Printf("failed to execute DB %v", err)
	}

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

func initializedDb() (*sql.DB, error) {
	db, err := sql.Open("mysql", "root:mysql@tcp(127.0.0.1:3306)/demo")
	if err != nil {
		log.Printf("failed to open mysql %v", err)
		return db, err
	}

	return db, err
}

func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	res, err := handler(ctx, req)
	log.Printf("%s: %v -> %v", info.FullMethod, req, res)
	return res, err
}

func main() {
	log.Printf("Hello gRPC Web Server...")

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor),
	)

	db, err := initializedDb()
	if err != nil {
		log.Printf("failed to initialize db: %v", err)
	}

	srv := &server{}
	srv.db = db

	pb.RegisterGreeterServer(s, srv)
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	db.Close()
}
