package rpc

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/grpc-gateway/pb/echo/v1"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	grpc_ratelimit "github.com/grpc-ecosystem/go-grpc-middleware/ratelimit"
)

func Serve(rpcPort, restPort int, errCh chan error) error {
	err := serverGRPC(rpcPort, errCh)
	if err != nil {
		return fmt.Errorf("failed to serve gRPC server: %w", err)
	}

	if err := serverREST(restPort, rpcPort, errCh); err != nil {
		return fmt.Errorf("failed to serve REST server: %w", err)
	}

	return err
}

func serverGRPC(port int, errCh chan error) error {

	svr := grpc.NewServer(
		grpc.StreamInterceptor(
			grpc_ratelimit.StreamServerInterceptor(
				&grpcLimiter{limiter: limiter},
			),
		),
		grpc.UnaryInterceptor(
			grpc_ratelimit.UnaryServerInterceptor(
				&grpcLimiter{limiter: limiter},
			),
		),
		grpc.ConnectionTimeout(time.Second),
	)

	pb.RegisterEcoServiceServer(svr, &EchoServer{})

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen port for RPC: %w", err)
	}

	go func() {
		log.Infof("gRPC server listening at %d...", port)
		if err := svr.Serve(lis); err != nil {
			errCh <- err
			return
		}
	}()

	return nil
}

func serverREST(port, rpcPort int, errCh chan error) error {
	ctx := context.Background()

	rpcEndpoint := fmt.Sprintf("localhost:%d", rpcPort)

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime.NewServeMux()

	conn, err := grpc.DialContext(
		ctx,
		rpcEndpoint,
		grpc.WithBlock(),
		grpc.WithInsecure(),
	)

	if err != nil {
		return err
	}

	if err := pb.RegisterEcoServiceHandler(ctx, mux, conn); err != nil {
		return err
	}

	go func() {
		log.Infof("REST  server listening at %d...", port)
		if err := http.ListenAndServe("0.0.0.0:8080", limitREST(mux)); err != nil {
			errCh <- err
			return
		}
	}()
	return nil
}
