package coupon

import (
	"fxserver/modules/coupon/entity"
	"time"
)

type CreateCouponRequest struct {
	Code           string    `json:"code" validate:"required,min=3,max=50"`
	Name           string    `json:"name" validate:"required,min=2,max=100"`
	Description    string    `json:"description" validate:"required,min=5,max=500"`
	DiscountType   string    `json:"discount_type" validate:"required,oneof=percentage fixed"`
	DiscountValue  float64   `json:"discount_value" validate:"required,gt=0"`
	MinOrderAmount float64   `json:"min_order_amount" validate:"gte=0"`
	MaxDiscount    *float64  `json:"max_discount,omitempty" validate:"omitempty,gt=0"`
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
	ExpiresAt      time.Time `json:"expires_at,omitempty"`
}

type UseCouponRequest struct {
	Code        string  `json:"code" validate:"required"`
	UserID      int     `json:"user_id" validate:"required,gt=0"`
	OrderAmount float64 `json:"order_amount" validate:"required,gt=0"`
}

type ListCouponsResponse struct {
	Coupons []entity.CouponResponse `json:"coupons"`
	Total   int                     `json:"total"`
}