package coupon

import (
	"fxserver/modules/coupon/entity"
	itemEntity "fxserver/modules/item/entity"
	"time"
)

type CreateCouponRequest struct {
	Code           string    `json:"code" validate:"required,min=3,max=50"`
	Name           string    `json:"name" validate:"required,min=2,max=100"`
	Description    string    `json:"description" validate:"required,min=5,max=500"`
	DiscountType   string    `json:"discount_type" validate:"omitempty,oneof=percentage fixed"`
	DiscountValue  float64   `json:"discount_value" validate:"omitempty,gt=0"`
	MinOrderAmount float64   `json:"min_order_amount" validate:"omitempty,gte=0"`
	MaxDiscount    *float64  `json:"max_discount,omitempty" validate:"omitempty,gt=0"`
	RewardType     entity.RewardType `json:"reward_type" validate:"required"`
	RewardItems    []itemEntity.RewardItem `json:"reward_items,omitempty" validate:"omitempty,dive"`
	ExpiresAt      time.Time `json:"expires_at" validate:"required"`
}

type UpdateCouponRequest struct {
	Code           string    `json:"code,omitempty" validate:"omitempty,min=3,max=50"`
	Name           string    `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Description    string    `json:"description,omitempty" validate:"omitempty,min=5,max=500"`
	DiscountType   string    `json:"discount_type,omitempty" validate:"omitempty,oneof=percentage fixed"`
	DiscountValue  float64   `json:"discount_value,omitempty" validate:"omitempty,gt=0"`
	MinOrderAmount float64   `json:"min_order_amount,omitempty" validate:"omitempty,gte=0"`
	MaxDiscount    *float64  `json:"max_discount,omitempty" validate:"omitempty,gt=0"`
	RewardType     entity.RewardType `json:"reward_type,omitempty"`
	RewardItems    []itemEntity.RewardItem `json:"reward_items,omitempty" validate:"omitempty,dive"`
	ExpiresAt      time.Time `json:"expires_at,omitempty"`
}

type RedeemCouponRequest struct {
	Code        string  `json:"code" validate:"required"`
	UserID      int     `json:"user_id" validate:"required,gt=0"`
	OrderAmount float64 `json:"order_amount" validate:"omitempty,gt=0"` // 할인 타입인 경우에만 필수
}

type ListCouponsResponse struct {
	Coupons []entity.CouponResponse `json:"coupons"`
	Total   int                     `json:"total"`
}