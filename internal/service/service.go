package service

type SalesService interface {
	InsertData()
	GetEmployeeCartsInTrip()
	GetEmployeeIDsByTrip()
	UpdateItemQuantity()
	DeleteItemFromCart()
}

type salesService struct {
}

func NewSalesService() SalesService {

}

func (s *salesService) InsertData() error {
}

func (s *salesService) GetEmployeeCartsInTrip() {
}

func (s *salesService) GetEmployeeIDsByTrip() {
}

func (s *salesService) UpdateItemQuantity() {
}

func (s *salesService) DeleteItemFromCart() {

}
