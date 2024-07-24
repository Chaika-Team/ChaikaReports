package repository

import "ChaikaReports/internal/models"

type SalesRepository interface {
	StoreSalesData(salesData models.SalesData) error
}
