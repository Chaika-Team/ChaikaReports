package decoder

import (
	"context"

	"ChaikaReports/internal/models"
	pb "github.com/Chaika-Team/chaika-proto/gen/rprts"
)

func DecodeGetTripRequest(_ context.Context, req *pb.GetTripRequest) *models.TripID {
	return &models.TripID{
		RouteID:   req.RouteId,
		Year:      req.Year,
		StartTime: req.StartTime.AsTime(),
	}
}
