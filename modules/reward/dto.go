package reward

import "fxserver/modules/item/entity"

// Grant rewards DTOs
type GrantRewardRequest struct {
	UserID      int                   `json:"user_id" validate:"required,gt=0"`
	Items       []entity.RewardItem   `json:"items" validate:"required,min=1,dive"`
	Source      string                `json:"source" validate:"required,min=2,max=50"` // admin, event, compensation, etc.
	Description string                `json:"description" validate:"required,min=5,max=500"`
}

type BulkGrantRewardRequest struct {
	UserIDs     []int                 `json:"user_ids" validate:"required,min=1"`
	Items       []entity.RewardItem   `json:"items" validate:"required,min=1,dive"`
	Source      string                `json:"source" validate:"required,min=2,max=50"`
	Description string                `json:"description" validate:"required,min=5,max=500"`
}

// Response DTOs
type GrantRewardResponse struct {
	UserID      int                   `json:"user_id"`
	Items       []entity.RewardItem   `json:"items"`
	Source      string                `json:"source"`
	Description string                `json:"description"`
	GrantedAt   string                `json:"granted_at"`
	Success     bool                  `json:"success"`
	Message     string                `json:"message,omitempty"`
}

type BulkGrantRewardResponse struct {
	TotalUsers    int                   `json:"total_users"`
	SuccessCount  int                   `json:"success_count"`
	FailureCount  int                   `json:"failure_count"`
	Results       []GrantRewardResponse `json:"results"`
	Items         []entity.RewardItem   `json:"items"`
	Source        string                `json:"source"`
	Description   string                `json:"description"`
}

// Reward source constants
const (
	RewardSourceAdmin        = "admin"        // 관리자 직접 지급
	RewardSourceCoupon       = "coupon"       // 쿠폰 사용
	RewardSourcePayment      = "payment"      // 결제 완료
	RewardSourceEvent        = "event"        // 이벤트 보상
	RewardSourceCompensation = "compensation" // 보상/사과
	RewardSourceDaily        = "daily"        // 일일 보상
	RewardSourceAchievement  = "achievement"  // 업적 달성
)

// IsValidRewardSource validates if the reward source is valid
func IsValidRewardSource(source string) bool {
	switch source {
	case RewardSourceAdmin, RewardSourceCoupon, RewardSourcePayment,
		 RewardSourceEvent, RewardSourceCompensation, RewardSourceDaily,
		 RewardSourceAchievement:
		return true
	default:
		return false
	}
}

// GetRewardSourceDescription returns description for reward source
func GetRewardSourceDescription(source string) string {
	switch source {
	case RewardSourceAdmin:
		return "관리자 직접 지급"
	case RewardSourceCoupon:
		return "쿠폰 사용 보상"
	case RewardSourcePayment:
		return "결제 완료 보상"
	case RewardSourceEvent:
		return "이벤트 보상"
	case RewardSourceCompensation:
		return "보상/사과"
	case RewardSourceDaily:
		return "일일 보상"
	case RewardSourceAchievement:
		return "업적 달성 보상"
	default:
		return "알 수 없는 보상 출처"
	}
}