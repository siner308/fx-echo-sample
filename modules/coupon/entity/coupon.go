package entity

import "time"

type CouponStatus string

const (
	CouponStatusActive CouponStatus = "active"
	CouponStatusUsed   CouponStatus = "used"
	CouponStatusExpired CouponStatus = "expired"
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
	Status         CouponStatus  `json:"status"`
	ExpiresAt      time.Time     `json:"expires_at"`
	CreatedAt      time.Time     `json:"created_at"`
}

type UseCouponResponse struct {
	CouponID       int       `json:"coupon_id"`
	Code           string    `json:"code"`
	DiscountAmount float64   `json:"discount_amount"`
	UsedAt         time.Time `json:"used_at"`
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