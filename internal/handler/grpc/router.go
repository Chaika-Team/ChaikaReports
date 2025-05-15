package grpc

import (
	"ChaikaReports/internal/handler/grpc/decoder"
	"ChaikaReports/internal/handler/grpc/encoder"
	pb "ChaikaReports/internal/handler/grpc/pb/rprts"
	"ChaikaReports/internal/service"
	"context"
	"github.com/go-kit/log"
	"google.golang.org/grpc"
)

type Router struct {
	svc service.SalesService
	log log.Logger
	pb.UnimplementedSalesServiceServer
}

func NewRouter(svc service.SalesService, logger log.Logger) *Router {
	return &Router{svc: svc, log: logger}
}

func RegisterGRPCServer(s *grpc.Server, router *Router) {
	pb.RegisterSalesServiceServer(s, router)
}

func (r *Router) GetTrip(ctx context.Context, req *pb.GetTripRequest) (*pb.GetTripReply, error) {
	// 1) decode
	tid := decoder.DecodeGetTripRequest(ctx, req)

	// 2) business
	trip, err := r.svc.GetTrip(ctx, tid)
	if err != nil {
		_ = r.log.Log("method", "GetTrip", "err", err)
		return nil, err
	}

	// 3) encode
	return encoder.EncodeGetTripReply(trip), nil
}
