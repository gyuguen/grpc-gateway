package rpc

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/grpc-gateway/pb/echo/v1"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Serve(rpcPort, restPort int) (*grpc.Server, error) {
	svr, err := serverGRPC(rpcPort)
	if err != nil {
		return nil, fmt.Errorf("failed to serve gRPC server: %w", err)
	}

	if err := serverREST(restPort, rpcPort); err != nil {
		return nil, fmt.Errorf("failed to serve REST server: %w", err)
	}

	return svr, err
}

func serverGRPC(port int) (*grpc.Server, error) {
	svr := grpc.NewServer()
	pb.RegisterEcoServiceServer(svr, &EchoServer{})

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen port for RPC: %w", err)
	}

	go func() {
		log.Infof("gRPC server listening at %d...", port)
		if err := svr.Serve(lis); err != nil {
			log.Panicf("gRPC server shutted down: %v", err)
		}
	}()

	return svr, nil
}

func serverREST(port, rpcPort int) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	rpcEndpoint := fmt.Sprintf("localhost:%d", rpcPort)

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := pb.RegisterEcoServiceHandlerFromEndpoint(ctx, mux, rpcEndpoint, opts); err != nil {
		return err
	}
	return http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}
