package repository

type SalesRepository interface {
	InsertData() error
	GetConductorCarts()
	UpdateItemCount() error
	DeleteItemFromCart() error
	GetEmployeesByTrip()
}
