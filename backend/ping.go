package main

import (
	"context"
	pb "invoice-manager/main/proto"
	pbconnect "invoice-manager/main/proto/protoconnect"
	"log"
	"net/http"

	"connectrpc.com/connect"
)

type PingServer struct {
	pbconnect.UnimplementedPingServiceHandler
}

func (ps *PingServer) Ping(
	ctx context.Context,
	req *connect.Request[pb.PingRequest],
) (*connect.Response[pb.PingResponse], error) {
	log.Println("Hit PingServer::Ping()")
	res := connect.NewResponse(&pb.PingResponse{
		Text:   req.Msg.Text,
		Number: req.Msg.Number,
	})
	res.Header().Set("Some-Other-Header", "hello!")
	return res, nil
}

func PingServiceHandler() (string, http.Handler) {
	return pbconnect.NewPingServiceHandler(&PingServer{})
}
