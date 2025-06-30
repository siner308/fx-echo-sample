package payment

import (
	itemEntity "fxserver/modules/item/entity"
	"fxserver/modules/payment/entity"
)

// Payment request DTOs
type CreatePaymentRequest struct {
	UserID      int                   `json:"user_id" validate:"required,gt=0"`
	Amount      float64               `json:"amount" validate:"required,gt=0"`
	Currency    string                `json:"currency" validate:"required,len=3"` // USD, KRW, etc.
	Method      entity.PaymentMethod  `json:"method" validate:"required"`
	ExternalID  string                `json:"external_id" validate:"required"`   // 외부 결제 시스템 ID
	RewardItems []itemEntity.RewardItem   `json:"reward_items" validate:"required,min=1,dive"`
}

type UpdatePaymentStatusRequest struct {
	Status        entity.PaymentStatus `json:"status" validate:"required"`
	FailureReason string               `json:"failure_reason,omitempty"`
}

type RefundPaymentRequest struct {
	Reason string `json:"reason" validate:"required,min=5,max=500"`
}

// Query DTOs
type GetPaymentsQuery struct {
	UserID    int                   `query:"user_id"`
	Status    entity.PaymentStatus  `query:"status"`
	StartDate string                `query:"start_date"` // YYYY-MM-DD format
	EndDate   string                `query:"end_date"`   // YYYY-MM-DD format
	Method    entity.PaymentMethod  `query:"method"`
}

// Response DTOs
type ProcessPaymentResponse struct {
	PaymentID   int                 `json:"payment_id"`
	Status      entity.PaymentStatus `json:"status"`
	Message     string              `json:"message"`
	RewardItems []itemEntity.RewardItem `json:"reward_items,omitempty"`
}

type PaymentMethodInfo struct {
	Method      entity.PaymentMethod `json:"method"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	IsActive    bool                 `json:"is_active"`
}

type PaymentMethodsResponse struct {
	Methods []PaymentMethodInfo `json:"methods"`
}

type PaymentStatusInfo struct {
	Status      entity.PaymentStatus `json:"status"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
}

type PaymentStatusesResponse struct {
	Statuses []PaymentStatusInfo `json:"statuses"`
}

// Helper functions
func GetPaymentMethods() []PaymentMethodInfo {
	return []PaymentMethodInfo{
		{
			Method:      entity.PaymentMethodCard,
			Name:        "신용카드",
			Description: entity.PaymentMethodCard.GetDescription(),
			IsActive:    true,
		},
		{
			Method:      entity.PaymentMethodBank,
			Name:        "계좌이체",
			Description: entity.PaymentMethodBank.GetDescription(),
			IsActive:    true,
		},
		{
			Method:      entity.PaymentMethodPaypal,
			Name:        "PayPal",
			Description: entity.PaymentMethodPaypal.GetDescription(),
			IsActive:    true,
		},
		{
			Method:      entity.PaymentMethodApple,
			Name:        "Apple Pay",
			Description: entity.PaymentMethodApple.GetDescription(),
			IsActive:    true,
		},
		{
			Method:      entity.PaymentMethodGoogle,
			Name:        "Google Pay",
			Description: entity.PaymentMethodGoogle.GetDescription(),
			IsActive:    true,
		},
	}
}

func GetPaymentStatuses() []PaymentStatusInfo {
	return []PaymentStatusInfo{
		{
			Status:      entity.PaymentStatusPending,
			Name:        "대기중",
			Description: entity.PaymentStatusPending.GetDescription(),
		},
		{
			Status:      entity.PaymentStatusProcessing,
			Name:        "처리중",
			Description: entity.PaymentStatusProcessing.GetDescription(),
		},
		{
			Status:      entity.PaymentStatusCompleted,
			Name:        "완료",
			Description: entity.PaymentStatusCompleted.GetDescription(),
		},
		{
			Status:      entity.PaymentStatusFailed,
			Name:        "실패",
			Description: entity.PaymentStatusFailed.GetDescription(),
		},
		{
			Status:      entity.PaymentStatusCancelled,
			Name:        "취소",
			Description: entity.PaymentStatusCancelled.GetDescription(),
		},
		{
			Status:      entity.PaymentStatusRefunded,
			Name:        "환불",
			Description: entity.PaymentStatusRefunded.GetDescription(),
		},
	}
}