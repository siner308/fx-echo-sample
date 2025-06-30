package coupon

import (
	"errors"
	"fxserver/modules/coupon/entity"
	"fxserver/modules/coupon/repository"
	"time"

	"go.uber.org/zap"
)

var (
	ErrInvalidCouponData = errors.New("invalid coupon data")
	ErrInvalidOrderAmount = errors.New("invalid order amount")
)

type Service interface {
	CreateCoupon(req CreateCouponRequest) (*entity.Coupon, error)
	GetCoupon(id int) (*entity.Coupon, error)
	GetCouponByCode(code string) (*entity.Coupon, error)
	UpdateCoupon(id int, req UpdateCouponRequest) (*entity.Coupon, error)
	DeleteCoupon(id int) error
	ListCoupons() ([]*entity.Coupon, error)
	ListCouponsByStatus(status entity.CouponStatus) ([]*entity.Coupon, error)
	UseCoupon(req UseCouponRequest) (*entity.UseCouponResponse, error)
}

type service struct {
	repo   repository.CouponRepository
	logger *zap.Logger
}

func NewService(repo repository.CouponRepository, logger *zap.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

func (s *service) CreateCoupon(req CreateCouponRequest) (*entity.Coupon, error) {
	coupon := &entity.Coupon{
		Code:           req.Code,
		Name:           req.Name,
		Description:    req.Description,
		DiscountType:   req.DiscountType,
		DiscountValue:  req.DiscountValue,
		MinOrderAmount: req.MinOrderAmount,
		MaxDiscount:    req.MaxDiscount,
		ExpiresAt:      req.ExpiresAt,
	}

	if err := s.repo.Create(coupon); err != nil {
		if errors.Is(err, repository.ErrCouponExists) {
			s.logger.Warn("Attempt to create coupon with existing code", zap.String("code", req.Code))
			return nil, err
		}
		s.logger.Error("Failed to create coupon", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Coupon created successfully", zap.Int("coupon_id", coupon.ID))
	return coupon, nil
}

func (s *service) GetCoupon(id int) (*entity.Coupon, error) {
	coupon, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrCouponNotFound) {
			s.logger.Warn("Coupon not found", zap.Int("coupon_id", id))
			return nil, err
		}
		s.logger.Error("Failed to get coupon", zap.Int("coupon_id", id), zap.Error(err))
		return nil, err
	}

	return coupon, nil
}

func (s *service) GetCouponByCode(code string) (*entity.Coupon, error) {
	coupon, err := s.repo.GetByCode(code)
	if err != nil {
		if errors.Is(err, repository.ErrCouponNotFound) {
			s.logger.Warn("Coupon not found", zap.String("code", code))
			return nil, err
		}
		s.logger.Error("Failed to get coupon by code", zap.String("code", code), zap.Error(err))
		return nil, err
	}

	return coupon, nil
}

func (s *service) UpdateCoupon(id int, req UpdateCouponRequest) (*entity.Coupon, error) {
	existingCoupon, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrCouponNotFound) {
			s.logger.Warn("Coupon not found for update", zap.Int("coupon_id", id))
			return nil, err
		}
		s.logger.Error("Failed to get coupon for update", zap.Int("coupon_id", id), zap.Error(err))
		return nil, err
	}

	// Update only provided fields
	if req.Code != "" {
		existingCoupon.Code = req.Code
	}
	if req.Name != "" {
		existingCoupon.Name = req.Name
	}
	if req.Description != "" {
		existingCoupon.Description = req.Description
	}
	if req.DiscountType != "" {
		existingCoupon.DiscountType = req.DiscountType
	}
	if req.DiscountValue != 0 {
		existingCoupon.DiscountValue = req.DiscountValue
	}
	if req.MinOrderAmount != 0 {
		existingCoupon.MinOrderAmount = req.MinOrderAmount
	}
	if req.MaxDiscount != nil {
		existingCoupon.MaxDiscount = req.MaxDiscount
	}
	if !req.ExpiresAt.IsZero() {
		existingCoupon.ExpiresAt = req.ExpiresAt
	}

	// Update status based on expiration
	if existingCoupon.IsExpired() && existingCoupon.Status == entity.CouponStatusActive {
		existingCoupon.Status = entity.CouponStatusExpired
	}

	if err := s.repo.Update(existingCoupon); err != nil {
		if errors.Is(err, repository.ErrCouponExists) {
			s.logger.Warn("Attempt to update coupon with existing code", zap.String("code", req.Code))
			return nil, err
		}
		s.logger.Error("Failed to update coupon", zap.Int("coupon_id", id), zap.Error(err))
		return nil, err
	}

	s.logger.Info("Coupon updated successfully", zap.Int("coupon_id", id))
	return existingCoupon, nil
}

func (s *service) DeleteCoupon(id int) error {
	if err := s.repo.Delete(id); err != nil {
		if errors.Is(err, repository.ErrCouponNotFound) {
			s.logger.Warn("Coupon not found for deletion", zap.Int("coupon_id", id))
			return err
		}
		s.logger.Error("Failed to delete coupon", zap.Int("coupon_id", id), zap.Error(err))
		return err
	}

	s.logger.Info("Coupon deleted successfully", zap.Int("coupon_id", id))
	return nil
}

func (s *service) ListCoupons() ([]*entity.Coupon, error) {
	coupons, err := s.repo.List()
	if err != nil {
		s.logger.Error("Failed to list coupons", zap.Error(err))
		return nil, err
	}

	return coupons, nil
}

func (s *service) ListCouponsByStatus(status entity.CouponStatus) ([]*entity.Coupon, error) {
	coupons, err := s.repo.ListByStatus(status)
	if err != nil {
		s.logger.Error("Failed to list coupons by status", zap.String("status", string(status)), zap.Error(err))
		return nil, err
	}

	return coupons, nil
}

func (s *service) UseCoupon(req UseCouponRequest) (*entity.UseCouponResponse, error) {
	if req.OrderAmount <= 0 {
		return nil, ErrInvalidOrderAmount
	}

	coupon, err := s.repo.GetByCode(req.Code)
	if err != nil {
		if errors.Is(err, repository.ErrCouponNotFound) {
			s.logger.Warn("Coupon not found for use", zap.String("code", req.Code))
			return nil, err
		}
		s.logger.Error("Failed to get coupon for use", zap.String("code", req.Code), zap.Error(err))
		return nil, err
	}

	// Check if coupon is usable
	if !coupon.IsUsable() {
		s.logger.Warn("Coupon is not usable", 
			zap.String("code", req.Code), 
			zap.String("status", string(coupon.Status)),
			zap.Bool("expired", coupon.IsExpired()))
		return nil, repository.ErrCouponNotUsable
	}

	// Check if coupon is already used
	if coupon.Status == entity.CouponStatusUsed {
		s.logger.Warn("Coupon already used", zap.String("code", req.Code))
		return nil, repository.ErrCouponAlreadyUsed
	}

	// Calculate discount
	discountAmount := coupon.CalculateDiscount(req.OrderAmount)
	if discountAmount == 0 {
		s.logger.Warn("Order amount does not meet minimum requirement", 
			zap.String("code", req.Code),
			zap.Float64("order_amount", req.OrderAmount),
			zap.Float64("min_order_amount", coupon.MinOrderAmount))
		return nil, errors.New("order amount does not meet minimum requirement")
	}

	// Mark coupon as used
	now := time.Now()
	coupon.Status = entity.CouponStatusUsed
	coupon.UsedBy = &req.UserID
	coupon.UsedAt = &now

	if err := s.repo.Update(coupon); err != nil {
		s.logger.Error("Failed to mark coupon as used", zap.String("code", req.Code), zap.Error(err))
		return nil, err
	}

	response := &entity.UseCouponResponse{
		CouponID:       coupon.ID,
		Code:           coupon.Code,
		DiscountAmount: discountAmount,
		UsedAt:         now,
	}

	s.logger.Info("Coupon used successfully", 
		zap.String("code", req.Code),
		zap.Int("user_id", req.UserID),
		zap.Float64("discount_amount", discountAmount))

	return response, nil
}