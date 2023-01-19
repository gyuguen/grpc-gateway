package rpc

import (
	"context"
	"time"

	pb "github.com/grpc-gateway/pb/echo/v1"
)

type EchoServer struct {
	pb.UnimplementedEcoServiceServer
}

func (s *EchoServer) Echo(ctx context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
	time.Sleep(time.Second * 2)
	return &pb.EchoResponse{
		Message: req.Name,
	}, nil
}
