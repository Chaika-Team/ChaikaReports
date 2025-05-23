package grpc

import (
	"ChaikaReports/internal/handler/grpc/decoder"
	"ChaikaReports/internal/handler/grpc/encoder"
	pb "ChaikaReports/internal/handler/grpc/pb/rprts"
	"ChaikaReports/internal/service"
	"context"
	"github.com/go-kit/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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
	reflection.Register(s)
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

func (r *Router) DeleteSyncedTrip(
	ctx context.Context, req *pb.DeleteSyncedTripRequest) (*pb.AckReply, error) {

	if err := r.svc.DeleteSyncedTrip(ctx,
		req.RouteId, req.StartTime.AsTime()); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &pb.AckReply{Message: "deleted"}, nil
}

func (r *Router) GetUnsyncedTrips(
	ctx context.Context, _ *emptypb.Empty) (*pb.GetUnsyncedTripsReply, error) {

	trips, err := r.svc.GetUnsyncedTrips(ctx)
	if err != nil {
		_ = r.log.Log("method", "GetUnsyncedTrips", "err", err)
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return encoder.EncodeGetUnsyncedTripsReply(trips), nil
}
