package coupon

import (
	"errors"
	"fmt"
	"time"

	"fxserver/modules/coupon/entity"
	"fxserver/modules/coupon/repository"
	"fxserver/modules/reward"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var (
	ErrInvalidCouponData  = errors.New("invalid coupon data")
	ErrInvalidOrderAmount = errors.New("invalid order amount")
	ErrCouponNotFound     = errors.New("coupon not found")
	ErrCouponNotUsable    = errors.New("coupon not usable")
	ErrInvalidRewardType  = errors.New("invalid reward type")
)

type Service interface {
	CreateCoupon(req CreateCouponRequest) (*entity.Coupon, error)
	GetCoupon(id int) (*entity.Coupon, error)
	GetCouponByCode(code string) (*entity.Coupon, error)
	UpdateCoupon(id int, req UpdateCouponRequest) (*entity.Coupon, error)
	DeleteCoupon(id int) error
	ListCoupons() ([]*entity.Coupon, error)
	ListCouponsByStatus(status entity.CouponStatus) ([]*entity.Coupon, error)
	RedeemCoupon(req RedeemCouponRequest) (*entity.RedeemCouponResponse, error)
}

type service struct {
	repo          repository.CouponRepository
	rewardService reward.Service
	logger        *zap.Logger
}

type ServiceParam struct {
	fx.In
	Repository    repository.CouponRepository
	RewardService reward.Service
	Logger        *zap.Logger
}

func NewService(p ServiceParam) Service {
	return &service{
		repo:          p.Repository,
		rewardService: p.RewardService,
		logger:        p.Logger,
	}
}

func (s *service) CreateCoupon(req CreateCouponRequest) (*entity.Coupon, error) {
	// Validate reward type
	if !entity.IsValidRewardType(string(req.RewardType)) {
		return nil, ErrInvalidRewardType
	}

	// Validate reward type and required fields
	if req.RewardType == entity.RewardTypeDiscountOnly || req.RewardType == entity.RewardTypeBoth {
		if req.DiscountType == "" || req.DiscountValue <= 0 {
			return nil, fmt.Errorf("discount type and value are required for discount rewards")
		}
	}

	if req.RewardType == entity.RewardTypeItemsOnly || req.RewardType == entity.RewardTypeBoth {
		if len(req.RewardItems) == 0 {
			return nil, fmt.Errorf("reward items are required for item rewards")
		}
		// Validate reward items
		if err := s.rewardService.ValidateRewardItems(req.RewardItems); err != nil {
			return nil, fmt.Errorf("invalid reward items: %w", err)
		}
	}

	coupon := &entity.Coupon{
		Code:           req.Code,
		Name:           req.Name,
		Description:    req.Description,
		DiscountType:   req.DiscountType,
		DiscountValue:  req.DiscountValue,
		MinOrderAmount: req.MinOrderAmount,
		MaxDiscount:    req.MaxDiscount,
		RewardType:     req.RewardType,
		RewardItems:    req.RewardItems,
		ExpiresAt:      req.ExpiresAt,
		Status:         entity.CouponStatusActive,
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

func (s *service) RedeemCoupon(req RedeemCouponRequest) (*entity.RedeemCouponResponse, error) {
	// Get coupon by code
	coupon, err := s.repo.GetByCode(req.Code)
	if err != nil {
		if errors.Is(err, repository.ErrCouponNotFound) {
			s.logger.Warn("Coupon not found for redemption", zap.String("code", req.Code))
			return nil, ErrCouponNotFound
		}
		s.logger.Error("Failed to get coupon for redemption", zap.String("code", req.Code), zap.Error(err))
		return nil, err
	}

	// Check if coupon is usable
	if !coupon.IsUsable() {
		s.logger.Warn("Coupon is not usable", 
			zap.String("code", req.Code), 
			zap.String("status", string(coupon.Status)),
			zap.Bool("expired", coupon.IsExpired()))
		return nil, ErrCouponNotUsable
	}

	// Check if coupon is already used
	if coupon.Status == entity.CouponStatusUsed {
		s.logger.Warn("Coupon already used", zap.String("code", req.Code))
		return nil, errors.New("coupon already used")
	}

	var discountAmount float64
	var message string

	// Handle discount processing
	if coupon.HasDiscount() {
		if req.OrderAmount <= 0 {
			return nil, fmt.Errorf("order amount is required for discount coupons")
		}
		
		discountAmount = coupon.CalculateDiscount(req.OrderAmount)
		if discountAmount == 0 {
			s.logger.Warn("Order amount does not meet minimum requirement", 
				zap.String("code", req.Code),
				zap.Float64("order_amount", req.OrderAmount),
				zap.Float64("min_order_amount", coupon.MinOrderAmount))
			return nil, fmt.Errorf("order amount does not meet minimum requirement of %.2f", coupon.MinOrderAmount)
		}
	}

	// Handle item rewards
	if coupon.HasRewardItems() {
		// Grant reward items through reward service
		if err := s.rewardService.GrantItemsToUser(
			req.UserID, 
			coupon.RewardItems, 
			reward.RewardSourceCoupon,
			fmt.Sprintf("Coupon redemption: %s", coupon.Name),
		); err != nil {
			s.logger.Error("Failed to grant reward items", 
				zap.String("code", req.Code),
				zap.Int("user_id", req.UserID),
				zap.Error(err))
			return nil, fmt.Errorf("failed to grant reward items: %w", err)
		}
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

	// Prepare response message
	if coupon.HasDiscount() && coupon.HasRewardItems() {
		message = fmt.Sprintf("Coupon redeemed successfully! Received %.2f discount and %d reward items", discountAmount, len(coupon.RewardItems))
	} else if coupon.HasDiscount() {
		message = fmt.Sprintf("Coupon redeemed successfully! Received %.2f discount", discountAmount)
	} else if coupon.HasRewardItems() {
		message = fmt.Sprintf("Coupon redeemed successfully! Received %d reward items", len(coupon.RewardItems))
	} else {
		message = "Coupon redeemed successfully!"
	}

	response := &entity.RedeemCouponResponse{
		CouponID:       coupon.ID,
		Code:           coupon.Code,
		DiscountAmount: discountAmount,
		RewardItems:    coupon.RewardItems,
		UsedAt:         now,
		Message:        message,
	}

	s.logger.Info("Coupon redeemed successfully", 
		zap.String("code", req.Code),
		zap.Int("user_id", req.UserID),
		zap.Float64("discount_amount", discountAmount),
		zap.Int("reward_items_count", len(coupon.RewardItems)))

	return response, nil
}