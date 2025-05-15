package decoder

import (
	"context"

	pb "ChaikaReports/internal/handler/grpc/pb/rprts"
	"ChaikaReports/internal/models"
)

func DecodeGetTripRequest(_ context.Context, req *pb.GetTripRequest) *models.TripID {
	return &models.TripID{
		RouteID:   req.RouteId,
		Year:      req.Year,
		StartTime: req.StartTime.AsTime(),
	}
}
