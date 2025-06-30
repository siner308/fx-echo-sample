package payment

import (
	"errors"
	"fmt"
	"time"

	paymentEntity "fxserver/modules/payment/entity"
	"fxserver/modules/payment/repository"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var (
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrInvalidPaymentStatus = errors.New("invalid payment status")
	ErrInvalidPaymentMethod = errors.New("invalid payment method")
	ErrPaymentAlreadyExists = errors.New("payment with external ID already exists")
	ErrCannotRefund         = errors.New("payment cannot be refunded")
	ErrInvalidAmount        = errors.New("invalid payment amount")
)

type Service interface {
	// Payment processing
	ProcessPayment(req CreatePaymentRequest) (*ProcessPaymentResponse, error)
	UpdatePaymentStatus(paymentID int, req UpdatePaymentStatusRequest) (*paymentEntity.Payment, error)
	RefundPayment(paymentID int, req RefundPaymentRequest) (*paymentEntity.Payment, error)

	// Payment queries
	GetPayment(id int) (*paymentEntity.Payment, error)
	GetPaymentByExternalID(externalID string) (*paymentEntity.Payment, error)
	GetUserPayments(userID int) (*paymentEntity.PaymentHistoryResponse, error)
	GetUserPaymentsByStatus(userID int, status paymentEntity.PaymentStatus) (*paymentEntity.PaymentHistoryResponse, error)

	// Admin operations
	GetAllPayments() (*paymentEntity.PaymentHistoryResponse, error)
	GetPaymentsByStatus(status paymentEntity.PaymentStatus) (*paymentEntity.PaymentHistoryResponse, error)
	GetPaymentsByDateRange(startDate, endDate string) (*paymentEntity.PaymentHistoryResponse, error)

	// Statistics
	GetPaymentSummaryByUser(userID int) (*paymentEntity.PaymentSummaryResponse, error)
	GetPaymentSummary() (*paymentEntity.PaymentSummaryResponse, error)

	// Utility
	GetPaymentMethods() []PaymentMethodInfo
	GetPaymentStatuses() []PaymentStatusInfo
}

type service struct {
	repository repository.Repository
	logger     *zap.Logger
}

type ServiceParam struct {
	fx.In
	Repository repository.Repository
	Logger     *zap.Logger
}

func NewService(p ServiceParam) Service {
	return &service{
		repository: p.Repository,
		logger:     p.Logger,
	}
}

func (s *service) ProcessPayment(req CreatePaymentRequest) (*ProcessPaymentResponse, error) {
	// Validate payment method
	if !paymentEntity.IsValidPaymentMethod(string(req.Method)) {
		return nil, ErrInvalidPaymentMethod
	}

	// Validate amount
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	// Check if payment with external ID already exists
	if _, err := s.repository.GetPaymentByExternalID(req.ExternalID); err == nil {
		return nil, ErrPaymentAlreadyExists
	}

	// Validate reward items
	if len(req.RewardItems) == 0 {
		return nil, errors.New("at least one reward item is required")
	}

	for _, item := range req.RewardItems {
		if item.ItemID <= 0 || item.Count <= 0 {
			return nil, fmt.Errorf("invalid reward item: itemID=%d, count=%d", item.ItemID, item.Count)
		}
	}

	// Create payment record
	payment := &paymentEntity.Payment{
		UserID:      req.UserID,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Status:      paymentEntity.PaymentStatusPending,
		Method:      req.Method,
		ExternalID:  req.ExternalID,
		RewardItems: req.RewardItems,
	}

	if err := s.repository.CreatePayment(payment); err != nil {
		s.logger.Error("Failed to create payment", zap.Error(err))
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	s.logger.Info("Payment created successfully", 
		zap.Int("payment_id", payment.ID),
		zap.Int("user_id", req.UserID),
		zap.Float64("amount", req.Amount),
		zap.String("currency", req.Currency),
		zap.String("method", string(req.Method)))

	// In a real implementation, this would integrate with actual payment processors
	// For now, we'll simulate immediate success
	return &ProcessPaymentResponse{
		PaymentID:   payment.ID,
		Status:      payment.Status,
		Message:     "Payment created successfully. Awaiting external payment confirmation.",
		RewardItems: req.RewardItems,
	}, nil
}

func (s *service) UpdatePaymentStatus(paymentID int, req UpdatePaymentStatusRequest) (*paymentEntity.Payment, error) {
	// Validate status
	if !paymentEntity.IsValidPaymentStatus(string(req.Status)) {
		return nil, ErrInvalidPaymentStatus
	}

	// Get existing payment
	payment, err := s.repository.GetPayment(paymentID)
	if err != nil {
		return nil, ErrPaymentNotFound
	}

	// Update status
	if err := s.repository.UpdatePaymentStatus(paymentID, req.Status, req.FailureReason); err != nil {
		s.logger.Error("Failed to update payment status", 
			zap.Error(err), 
			zap.Int("payment_id", paymentID),
			zap.String("status", string(req.Status)))
		return nil, fmt.Errorf("failed to update payment status: %w", err)
	}

	// Get updated payment
	updatedPayment, err := s.repository.GetPayment(paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated payment: %w", err)
	}

	s.logger.Info("Payment status updated", 
		zap.Int("payment_id", paymentID),
		zap.String("old_status", string(payment.Status)),
		zap.String("new_status", string(req.Status)))

	return updatedPayment, nil
}

func (s *service) RefundPayment(paymentID int, req RefundPaymentRequest) (*paymentEntity.Payment, error) {
	// Get existing payment
	payment, err := s.repository.GetPayment(paymentID)
	if err != nil {
		return nil, ErrPaymentNotFound
	}

	// Check if payment can be refunded
	if !payment.CanBeRefunded() {
		return nil, ErrCannotRefund
	}

	// Update payment status to refunded
	if err := s.repository.UpdatePaymentStatus(paymentID, paymentEntity.PaymentStatusRefunded, req.Reason); err != nil {
		s.logger.Error("Failed to refund payment", 
			zap.Error(err), 
			zap.Int("payment_id", paymentID))
		return nil, fmt.Errorf("failed to refund payment: %w", err)
	}

	// Get updated payment
	updatedPayment, err := s.repository.GetPayment(paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated payment: %w", err)
	}

	s.logger.Info("Payment refunded", 
		zap.Int("payment_id", paymentID),
		zap.Int("user_id", payment.UserID),
		zap.Float64("amount", payment.Amount),
		zap.String("reason", req.Reason))

	return updatedPayment, nil
}

func (s *service) GetPayment(id int) (*paymentEntity.Payment, error) {
	payment, err := s.repository.GetPayment(id)
	if err != nil {
		return nil, ErrPaymentNotFound
	}
	return payment, nil
}

func (s *service) GetPaymentByExternalID(externalID string) (*paymentEntity.Payment, error) {
	payment, err := s.repository.GetPaymentByExternalID(externalID)
	if err != nil {
		return nil, ErrPaymentNotFound
	}
	return payment, nil
}

func (s *service) GetUserPayments(userID int) (*paymentEntity.PaymentHistoryResponse, error) {
	payments, err := s.repository.GetUserPayments(userID)
	if err != nil {
		s.logger.Error("Failed to get user payments", zap.Error(err), zap.Int("user_id", userID))
		return nil, fmt.Errorf("failed to get user payments: %w", err)
	}

	paymentResponses := make([]paymentEntity.PaymentResponse, len(payments))
	for i, payment := range payments {
		paymentResponses[i] = payment.ToResponse()
	}

	return &paymentEntity.PaymentHistoryResponse{
		Payments: paymentResponses,
		Total:    len(paymentResponses),
	}, nil
}

func (s *service) GetUserPaymentsByStatus(userID int, status paymentEntity.PaymentStatus) (*paymentEntity.PaymentHistoryResponse, error) {
	if !paymentEntity.IsValidPaymentStatus(string(status)) {
		return nil, ErrInvalidPaymentStatus
	}

	payments, err := s.repository.GetUserPaymentsByStatus(userID, status)
	if err != nil {
		s.logger.Error("Failed to get user payments by status", 
			zap.Error(err), 
			zap.Int("user_id", userID),
			zap.String("status", string(status)))
		return nil, fmt.Errorf("failed to get user payments by status: %w", err)
	}

	paymentResponses := make([]paymentEntity.PaymentResponse, len(payments))
	for i, payment := range payments {
		paymentResponses[i] = payment.ToResponse()
	}

	return &paymentEntity.PaymentHistoryResponse{
		Payments: paymentResponses,
		Total:    len(paymentResponses),
	}, nil
}

func (s *service) GetAllPayments() (*paymentEntity.PaymentHistoryResponse, error) {
	payments, err := s.repository.GetAllPayments()
	if err != nil {
		s.logger.Error("Failed to get all payments", zap.Error(err))
		return nil, fmt.Errorf("failed to get all payments: %w", err)
	}

	paymentResponses := make([]paymentEntity.PaymentResponse, len(payments))
	for i, payment := range payments {
		paymentResponses[i] = payment.ToResponse()
	}

	return &paymentEntity.PaymentHistoryResponse{
		Payments: paymentResponses,
		Total:    len(paymentResponses),
	}, nil
}

func (s *service) GetPaymentsByStatus(status paymentEntity.PaymentStatus) (*paymentEntity.PaymentHistoryResponse, error) {
	if !paymentEntity.IsValidPaymentStatus(string(status)) {
		return nil, ErrInvalidPaymentStatus
	}

	payments, err := s.repository.GetPaymentsByStatus(status)
	if err != nil {
		s.logger.Error("Failed to get payments by status", 
			zap.Error(err), 
			zap.String("status", string(status)))
		return nil, fmt.Errorf("failed to get payments by status: %w", err)
	}

	paymentResponses := make([]paymentEntity.PaymentResponse, len(payments))
	for i, payment := range payments {
		paymentResponses[i] = payment.ToResponse()
	}

	return &paymentEntity.PaymentHistoryResponse{
		Payments: paymentResponses,
		Total:    len(paymentResponses),
	}, nil
}

func (s *service) GetPaymentsByDateRange(startDate, endDate string) (*paymentEntity.PaymentHistoryResponse, error) {
	// Validate date format
	if _, err := time.Parse("2006-01-02", startDate); err != nil {
		return nil, fmt.Errorf("invalid start date format: %w", err)
	}
	if _, err := time.Parse("2006-01-02", endDate); err != nil {
		return nil, fmt.Errorf("invalid end date format: %w", err)
	}

	payments, err := s.repository.GetPaymentsByDateRange(startDate, endDate)
	if err != nil {
		s.logger.Error("Failed to get payments by date range", 
			zap.Error(err), 
			zap.String("start_date", startDate),
			zap.String("end_date", endDate))
		return nil, fmt.Errorf("failed to get payments by date range: %w", err)
	}

	paymentResponses := make([]paymentEntity.PaymentResponse, len(payments))
	for i, payment := range payments {
		paymentResponses[i] = payment.ToResponse()
	}

	return &paymentEntity.PaymentHistoryResponse{
		Payments: paymentResponses,
		Total:    len(paymentResponses),
	}, nil
}

func (s *service) GetPaymentSummaryByUser(userID int) (*paymentEntity.PaymentSummaryResponse, error) {
	summary, err := s.repository.GetPaymentSummaryByUser(userID)
	if err != nil {
		s.logger.Error("Failed to get payment summary by user", zap.Error(err), zap.Int("user_id", userID))
		return nil, fmt.Errorf("failed to get payment summary by user: %w", err)
	}
	return summary, nil
}

func (s *service) GetPaymentSummary() (*paymentEntity.PaymentSummaryResponse, error) {
	summary, err := s.repository.GetPaymentSummary()
	if err != nil {
		s.logger.Error("Failed to get payment summary", zap.Error(err))
		return nil, fmt.Errorf("failed to get payment summary: %w", err)
	}
	return summary, nil
}

func (s *service) GetPaymentMethods() []PaymentMethodInfo {
	return GetPaymentMethods()
}

func (s *service) GetPaymentStatuses() []PaymentStatusInfo {
	return GetPaymentStatuses()
}