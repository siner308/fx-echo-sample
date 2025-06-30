package repository

import (
	"errors"
	"fxserver/modules/coupon/entity"
)

var (
	ErrCouponNotFound    = errors.New("coupon not found")
	ErrCouponExists      = errors.New("coupon already exists")
	ErrCouponNotUsable   = errors.New("coupon is not usable")
	ErrCouponAlreadyUsed = errors.New("coupon already used")
)

type CouponRepository interface {
	Create(coupon *entity.Coupon) error
	GetByID(id int) (*entity.Coupon, error)
	GetByCode(code string) (*entity.Coupon, error)
	Update(coupon *entity.Coupon) error
	Delete(id int) error
	List() ([]*entity.Coupon, error)
	ListByStatus(status entity.CouponStatus) ([]*entity.Coupon, error)
}