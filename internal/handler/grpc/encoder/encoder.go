package encoder

import (
	pb "ChaikaReports/internal/handler/grpc/pb/rprts"
	"ChaikaReports/internal/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func EncodeGetTripReply(trip models.Trip) *pb.GetTripReply {
	out := &pb.GetTripReply{Trip: &pb.Trip{}}
	for _, car := range trip.Carriage {
		pbc := &pb.Carriage{
			TripId: &pb.TripID{
				RouteId:   car.TripID.RouteID,
				Year:      car.TripID.Year,
				StartTime: timestamppb.New(car.TripID.StartTime),
			},
			EndTime:    timestamppb.New(car.EndTime),
			CarriageId: int32(car.CarriageID),
		}
		for _, c := range car.Carts {
			pbcart := &pb.Cart{
				CartId: &pb.CartID{
					EmployeeId:    c.CartID.EmployeeID,
					OperationTime: timestamppb.New(c.CartID.OperationTime),
				},
				OperationType: int32(c.OperationType),
			}
			for _, it := range c.Items {
				pbcart.Items = append(pbcart.Items, &pb.Item{
					ProductId: int32(it.ProductID),
					Quantity:  int32(it.Quantity),
					Price:     it.Price,
				})
			}
			pbc.Carts = append(pbc.Carts, pbcart)
		}
		out.Trip.Carriage = append(out.Trip.Carriage, pbc)
	}
	return out
}
