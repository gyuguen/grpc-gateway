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

	errCh := make(chan error)

	err := rpc.Serve(*grpcPort, 8080, errCh)
	if err != nil {
		log.Errorf("Err: %w", err)
	}

	if err := <-errCh; err != nil {
		log.Errorf("Err: %w", err)
	}
}
