package repository

import "fxserver/modules/payment/entity"

type Repository interface {
	// Payment operations
	CreatePayment(payment *entity.Payment) error
	GetPayment(id int) (*entity.Payment, error)
	GetPaymentByExternalID(externalID string) (*entity.Payment, error)
	UpdatePayment(payment *entity.Payment) error
	UpdatePaymentStatus(id int, status entity.PaymentStatus, reason string) error
	
	// User payment history
	GetUserPayments(userID int) ([]*entity.Payment, error)
	GetUserPaymentsByStatus(userID int, status entity.PaymentStatus) ([]*entity.Payment, error)
	
	// Admin operations
	GetAllPayments() ([]*entity.Payment, error)
	GetPaymentsByStatus(status entity.PaymentStatus) ([]*entity.Payment, error)
	GetPaymentsByDateRange(startDate, endDate string) ([]*entity.Payment, error)
	
	// Statistics
	GetPaymentSummaryByUser(userID int) (*entity.PaymentSummaryResponse, error)
	GetPaymentSummary() (*entity.PaymentSummaryResponse, error)
}