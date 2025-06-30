package repository

import (
	"fmt"
	"sync"
	"time"

	"fxserver/modules/payment/entity"
)

type memoryRepository struct {
	payments map[int]*entity.Payment
	counter  int
	mu       sync.RWMutex
}

func NewMemoryRepository() Repository {
	return &memoryRepository{
		payments: make(map[int]*entity.Payment),
		counter:  0,
	}
}

func (r *memoryRepository) CreatePayment(payment *entity.Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.counter++
	payment.ID = r.counter
	payment.CreatedAt = time.Now()
	payment.UpdatedAt = time.Now()
	r.payments[payment.ID] = payment
	return nil
}

func (r *memoryRepository) GetPayment(id int) (*entity.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	payment, exists := r.payments[id]
	if !exists {
		return nil, fmt.Errorf("payment with id %d not found", id)
	}
	return payment, nil
}

func (r *memoryRepository) GetPaymentByExternalID(externalID string) (*entity.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, payment := range r.payments {
		if payment.ExternalID == externalID {
			return payment, nil
		}
	}
	return nil, fmt.Errorf("payment with external id %s not found", externalID)
}

func (r *memoryRepository) UpdatePayment(payment *entity.Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.payments[payment.ID]; !exists {
		return fmt.Errorf("payment with id %d not found", payment.ID)
	}

	payment.UpdatedAt = time.Now()
	r.payments[payment.ID] = payment
	return nil
}

func (r *memoryRepository) UpdatePaymentStatus(id int, status entity.PaymentStatus, reason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	payment, exists := r.payments[id]
	if !exists {
		return fmt.Errorf("payment with id %d not found", id)
	}

	payment.Status = status
	payment.UpdatedAt = time.Now()

	switch status {
	case entity.PaymentStatusCompleted:
		now := time.Now()
		payment.ProcessedAt = &now
		payment.FailureReason = ""
	case entity.PaymentStatusFailed, entity.PaymentStatusCancelled:
		payment.FailureReason = reason
	case entity.PaymentStatusRefunded:
		now := time.Now()
		payment.RefundedAt = &now
	}

	return nil
}

func (r *memoryRepository) GetUserPayments(userID int) ([]*entity.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var userPayments []*entity.Payment
	for _, payment := range r.payments {
		if payment.UserID == userID {
			userPayments = append(userPayments, payment)
		}
	}
	return userPayments, nil
}

func (r *memoryRepository) GetUserPaymentsByStatus(userID int, status entity.PaymentStatus) ([]*entity.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var userPayments []*entity.Payment
	for _, payment := range r.payments {
		if payment.UserID == userID && payment.Status == status {
			userPayments = append(userPayments, payment)
		}
	}
	return userPayments, nil
}

func (r *memoryRepository) GetAllPayments() ([]*entity.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	payments := make([]*entity.Payment, 0, len(r.payments))
	for _, payment := range r.payments {
		payments = append(payments, payment)
	}
	return payments, nil
}

func (r *memoryRepository) GetPaymentsByStatus(status entity.PaymentStatus) ([]*entity.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var payments []*entity.Payment
	for _, payment := range r.payments {
		if payment.Status == status {
			payments = append(payments, payment)
		}
	}
	return payments, nil
}

func (r *memoryRepository) GetPaymentsByDateRange(startDate, endDate string) ([]*entity.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date format: %w", err)
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date format: %w", err)
	}

	// Add 24 hours to end date to include the whole day
	end = end.Add(24 * time.Hour)

	var payments []*entity.Payment
	for _, payment := range r.payments {
		if payment.CreatedAt.After(start) && payment.CreatedAt.Before(end) {
			payments = append(payments, payment)
		}
	}
	return payments, nil
}

func (r *memoryRepository) GetPaymentSummaryByUser(userID int) (*entity.PaymentSummaryResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var totalAmount float64
	var completedCount, pendingCount, failedCount int

	for _, payment := range r.payments {
		if payment.UserID != userID {
			continue
		}

		switch payment.Status {
		case entity.PaymentStatusCompleted:
			totalAmount += payment.Amount
			completedCount++
		case entity.PaymentStatusPending, entity.PaymentStatusProcessing:
			pendingCount++
		case entity.PaymentStatusFailed, entity.PaymentStatusCancelled:
			failedCount++
		}
	}

	return &entity.PaymentSummaryResponse{
		TotalAmount:    totalAmount,
		CompletedCount: completedCount,
		PendingCount:   pendingCount,
		FailedCount:    failedCount,
	}, nil
}

func (r *memoryRepository) GetPaymentSummary() (*entity.PaymentSummaryResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var totalAmount float64
	var completedCount, pendingCount, failedCount int

	for _, payment := range r.payments {
		switch payment.Status {
		case entity.PaymentStatusCompleted:
			totalAmount += payment.Amount
			completedCount++
		case entity.PaymentStatusPending, entity.PaymentStatusProcessing:
			pendingCount++
		case entity.PaymentStatusFailed, entity.PaymentStatusCancelled:
			failedCount++
		}
	}

	return &entity.PaymentSummaryResponse{
		TotalAmount:    totalAmount,
		CompletedCount: completedCount,
		PendingCount:   pendingCount,
		FailedCount:    failedCount,
	}, nil
}