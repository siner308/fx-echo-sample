package entity

import (
	"time"

	itemEntity "fxserver/modules/item/entity"
)

type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"     // 결제 대기
	PaymentStatusProcessing PaymentStatus = "processing"  // 결제 처리 중
	PaymentStatusCompleted  PaymentStatus = "completed"   // 결제+지급 완료
	PaymentStatusFailed     PaymentStatus = "failed"      // 결제 실패
	PaymentStatusCancelled  PaymentStatus = "cancelled"   // 결제 취소
	PaymentStatusRefunded   PaymentStatus = "refunded"    // 환불 완료
)

type PaymentMethod string

const (
	PaymentMethodCard   PaymentMethod = "card"   // 신용카드
	PaymentMethodBank   PaymentMethod = "bank"   // 계좌이체
	PaymentMethodPaypal PaymentMethod = "paypal" // PayPal
	PaymentMethodApple  PaymentMethod = "apple"  // Apple Pay
	PaymentMethodGoogle PaymentMethod = "google" // Google Pay
)

type Payment struct {
	ID             int                   `json:"id"`
	UserID         int                   `json:"user_id"`
	Amount         float64               `json:"amount"`
	Currency       string                `json:"currency"`       // USD, KRW, etc.
	Status         PaymentStatus         `json:"status"`
	Method         PaymentMethod         `json:"method"`
	ExternalID     string                `json:"external_id"`    // 외부 결제 시스템 ID
	RewardItems    []itemEntity.RewardItem   `json:"reward_items"`   // 지급할 아이템들
	ProcessedAt    *time.Time            `json:"processed_at,omitempty"`
	FailureReason  string                `json:"failure_reason,omitempty"`
	RefundedAt     *time.Time            `json:"refunded_at,omitempty"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
}

// Payment Response DTOs
type PaymentResponse struct {
	ID            int                   `json:"id"`
	UserID        int                   `json:"user_id"`
	Amount        float64               `json:"amount"`
	Currency      string                `json:"currency"`
	Status        PaymentStatus         `json:"status"`
	Method        PaymentMethod         `json:"method"`
	ExternalID    string                `json:"external_id"`
	RewardItems   []itemEntity.RewardItem   `json:"reward_items"`
	ProcessedAt   *time.Time            `json:"processed_at,omitempty"`
	FailureReason string                `json:"failure_reason,omitempty"`
	RefundedAt    *time.Time            `json:"refunded_at,omitempty"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
}

type PaymentHistoryResponse struct {
	Payments []PaymentResponse `json:"payments"`
	Total    int               `json:"total"`
}

type PaymentSummaryResponse struct {
	TotalAmount    float64 `json:"total_amount"`
	CompletedCount int     `json:"completed_count"`
	PendingCount   int     `json:"pending_count"`
	FailedCount    int     `json:"failed_count"`
}

// Helper methods
func (p *Payment) ToResponse() PaymentResponse {
	return PaymentResponse{
		ID:            p.ID,
		UserID:        p.UserID,
		Amount:        p.Amount,
		Currency:      p.Currency,
		Status:        p.Status,
		Method:        p.Method,
		ExternalID:    p.ExternalID,
		RewardItems:   p.RewardItems,
		ProcessedAt:   p.ProcessedAt,
		FailureReason: p.FailureReason,
		RefundedAt:    p.RefundedAt,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

func (p *Payment) IsCompleted() bool {
	return p.Status == PaymentStatusCompleted
}

func (p *Payment) IsPending() bool {
	return p.Status == PaymentStatusPending || p.Status == PaymentStatusProcessing
}

func (p *Payment) IsFailed() bool {
	return p.Status == PaymentStatusFailed || p.Status == PaymentStatusCancelled
}

func (p *Payment) CanBeRefunded() bool {
	return p.Status == PaymentStatusCompleted
}

// GetStatusDescription returns description of payment status
func (s PaymentStatus) GetDescription() string {
	switch s {
	case PaymentStatusPending:
		return "결제 대기 중"
	case PaymentStatusProcessing:
		return "결제 처리 중"
	case PaymentStatusCompleted:
		return "결제 완료"
	case PaymentStatusFailed:
		return "결제 실패"
	case PaymentStatusCancelled:
		return "결제 취소"
	case PaymentStatusRefunded:
		return "환불 완료"
	default:
		return "알 수 없는 상태"
	}
}

// GetMethodDescription returns description of payment method
func (m PaymentMethod) GetDescription() string {
	switch m {
	case PaymentMethodCard:
		return "신용카드"
	case PaymentMethodBank:
		return "계좌이체"
	case PaymentMethodPaypal:
		return "PayPal"
	case PaymentMethodApple:
		return "Apple Pay"
	case PaymentMethodGoogle:
		return "Google Pay"
	default:
		return "알 수 없는 결제 방법"
	}
}

// IsValidPaymentStatus validates if the payment status is valid
func IsValidPaymentStatus(status string) bool {
	switch PaymentStatus(status) {
	case PaymentStatusPending, PaymentStatusProcessing, PaymentStatusCompleted,
		 PaymentStatusFailed, PaymentStatusCancelled, PaymentStatusRefunded:
		return true
	default:
		return false
	}
}

// IsValidPaymentMethod validates if the payment method is valid
func IsValidPaymentMethod(method string) bool {
	switch PaymentMethod(method) {
	case PaymentMethodCard, PaymentMethodBank, PaymentMethodPaypal,
		 PaymentMethodApple, PaymentMethodGoogle:
		return true
	default:
		return false
	}
}