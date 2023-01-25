package rpc

import (
	"context"
	"time"

	pb "github.com/grpc-gateway/pb/echo/v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type EchoServer struct {
	pb.UnimplementedEcoServiceServer
}

func (s *EchoServer) Echo(ctx context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
	time.Sleep(time.Second * 2)
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(10, "not found")
	}

	auth, ok := md["authorization"]
	if !ok {
		return nil, status.Errorf(10, "not found authorization")
	}
	return &pb.EchoResponse{
		Message: req.Name + " || " + auth[0],
	}, nil
}
