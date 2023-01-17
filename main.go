package main

import (
	"flag"

	rpc "github.com/grpc-gateway/rpc"
	log "github.com/sirupsen/logrus"
)

var (
	grpcPort = flag.Int("port", 9090, "The server port")
)

func main() {
	flag.Parse()

	_, err := rpc.Serve(*grpcPort, 8080)
	if err != nil {
		log.Errorf("Err: %w", err)
	}
}
