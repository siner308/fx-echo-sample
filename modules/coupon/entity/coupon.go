package entity

import (
	"time"
	
	"fxserver/modules/item/entity"
)

type CouponStatus string

const (
	CouponStatusActive CouponStatus = "active"
	CouponStatusUsed   CouponStatus = "used"
	CouponStatusExpired CouponStatus = "expired"
)

type RewardType string

const (
	RewardTypeDiscountOnly RewardType = "discount_only" // 할인만
	RewardTypeItemsOnly    RewardType = "items_only"    // 아이템만
	RewardTypeBoth         RewardType = "both"          // 할인+아이템
)

type Coupon struct {
	ID          int           `json:"id"`
	Code        string        `json:"code"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	DiscountType string       `json:"discount_type"` // "percentage" or "fixed"
	DiscountValue float64     `json:"discount_value"`
	MinOrderAmount float64    `json:"min_order_amount"`
	MaxDiscount   *float64    `json:"max_discount,omitempty"`
	RewardType    RewardType  `json:"reward_type"`               // 보상 타입
	RewardItems   []entity.RewardItem `json:"reward_items,omitempty"` // 추가 보상 아이템
	Status        CouponStatus `json:"status"`
	UsedBy        *int         `json:"used_by,omitempty"`
	UsedAt        *time.Time   `json:"used_at,omitempty"`
	ExpiresAt     time.Time    `json:"expires_at"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

type CouponResponse struct {
	ID             int           `json:"id"`
	Code           string        `json:"code"`
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	DiscountType   string        `json:"discount_type"`
	DiscountValue  float64       `json:"discount_value"`
	MinOrderAmount float64       `json:"min_order_amount"`
	MaxDiscount    *float64      `json:"max_discount,omitempty"`
	RewardType     RewardType    `json:"reward_type"`
	RewardItems    []entity.RewardItem `json:"reward_items,omitempty"`
	Status         CouponStatus  `json:"status"`
	ExpiresAt      time.Time     `json:"expires_at"`
	CreatedAt      time.Time     `json:"created_at"`
}

type RedeemCouponResponse struct {
	CouponID       int                   `json:"coupon_id"`
	Code           string                `json:"code"`
	DiscountAmount float64               `json:"discount_amount"`
	RewardItems    []entity.RewardItem   `json:"reward_items,omitempty"`
	UsedAt         time.Time             `json:"used_at"`
	Message        string                `json:"message"`
}

func (c *Coupon) ToResponse() CouponResponse {
	return CouponResponse{
		ID:             c.ID,
		Code:           c.Code,
		Name:           c.Name,
		Description:    c.Description,
		DiscountType:   c.DiscountType,
		DiscountValue:  c.DiscountValue,
		MinOrderAmount: c.MinOrderAmount,
		MaxDiscount:    c.MaxDiscount,
		RewardType:     c.RewardType,
		RewardItems:    c.RewardItems,
		Status:         c.Status,
		ExpiresAt:      c.ExpiresAt,
		CreatedAt:      c.CreatedAt,
	}
}

func (c *Coupon) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

func (c *Coupon) IsUsable() bool {
	return c.Status == CouponStatusActive && !c.IsExpired()
}

func (c *Coupon) CalculateDiscount(orderAmount float64) float64 {
	if orderAmount < c.MinOrderAmount {
		return 0
	}

	var discount float64
	if c.DiscountType == "percentage" {
		discount = orderAmount * (c.DiscountValue / 100)
		if c.MaxDiscount != nil && discount > *c.MaxDiscount {
			discount = *c.MaxDiscount
		}
	} else {
		discount = c.DiscountValue
	}

	if discount > orderAmount {
		return orderAmount
	}

	return discount
}

// HasDiscount returns true if the coupon provides discount
func (c *Coupon) HasDiscount() bool {
	return c.RewardType == RewardTypeDiscountOnly || c.RewardType == RewardTypeBoth
}

// HasRewardItems returns true if the coupon provides reward items
func (c *Coupon) HasRewardItems() bool {
	return (c.RewardType == RewardTypeItemsOnly || c.RewardType == RewardTypeBoth) && len(c.RewardItems) > 0
}

// IsValidRewardType validates if the reward type is valid
func IsValidRewardType(rewardType string) bool {
	switch RewardType(rewardType) {
	case RewardTypeDiscountOnly, RewardTypeItemsOnly, RewardTypeBoth:
		return true
	default:
		return false
	}
}

// GetRewardTypeDescription returns description for reward type
func (r RewardType) GetDescription() string {
	switch r {
	case RewardTypeDiscountOnly:
		return "할인만 제공"
	case RewardTypeItemsOnly:
		return "아이템만 제공"
	case RewardTypeBoth:
		return "할인과 아이템 모두 제공"
	default:
		return "알 수 없는 보상 타입"
	}
}