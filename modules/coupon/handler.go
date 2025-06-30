package coupon

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"fxserver/modules/coupon/entity"
	"fxserver/modules/coupon/repository"
	"fxserver/pkg/dto"
	"fxserver/pkg/validator"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Handler struct {
	service   Service
	validator validator.Validator
	logger    *zap.Logger
}

type HandlerParam struct {
	fx.In
	Service   Service
	Validator validator.Validator
	Logger    *zap.Logger
}

func NewHandler(p HandlerParam) *Handler {
	return &Handler{
		service:   p.Service,
		validator: p.Validator,
		logger:    p.Logger,
	}
}

// CreateCoupon creates a new coupon
func (h *Handler) CreateCoupon(c echo.Context) error {
	var req CreateCouponRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("Invalid request format", "invalid_request_error"))
	}

	if err := h.validator.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewValidationErrors(err))
	}

	coupon, err := h.service.CreateCoupon(req)
	if err != nil {
		if errors.Is(err, repository.ErrCouponExists) {
			return c.JSON(http.StatusConflict, dto.NewError("Coupon with this code already exists", "invalid_request_error"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("Failed to create coupon"))
	}

	return c.JSON(http.StatusCreated, coupon.ToResponse())
}

// GetCoupon retrieves coupon details by ID
func (h *Handler) GetCoupon(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("Invalid coupon ID", "invalid_request_error"))
	}

	coupon, err := h.service.GetCoupon(id)
	if err != nil {
		if errors.Is(err, repository.ErrCouponNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewNotFoundError("Coupon"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("Failed to get coupon"))
	}

	return c.JSON(http.StatusOK, coupon.ToResponse())
}

// UpdateCoupon updates an existing coupon
func (h *Handler) UpdateCoupon(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("Invalid coupon ID", "invalid_request_error"))
	}

	var req UpdateCouponRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("Invalid request format", "invalid_request_error"))
	}

	if err := h.validator.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewValidationErrors(err))
	}

	coupon, err := h.service.UpdateCoupon(id, req)
	if err != nil {
		if errors.Is(err, repository.ErrCouponNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewNotFoundError("Coupon"))
		}
		if errors.Is(err, repository.ErrCouponExists) {
			return c.JSON(http.StatusConflict, dto.NewError("Coupon with this code already exists", "invalid_request_error"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("Failed to update coupon"))
	}

	return c.JSON(http.StatusOK, coupon.ToResponse())
}

// DeleteCoupon deletes a coupon by ID
func (h *Handler) DeleteCoupon(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("Invalid coupon ID", "invalid_request_error"))
	}

	err = h.service.DeleteCoupon(id)
	if err != nil {
		if errors.Is(err, repository.ErrCouponNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewNotFoundError("Coupon"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("Failed to delete coupon"))
	}

	return c.NoContent(http.StatusNoContent)
}

// ListCoupons retrieves all coupons with optional status filter
func (h *Handler) ListCoupons(c echo.Context) error {
	status := c.QueryParam("status")

	var coupons []*entity.Coupon
	var err error

	if status != "" {
		couponStatus := entity.CouponStatus(status)
		coupons, err = h.service.ListCouponsByStatus(couponStatus)
	} else {
		coupons, err = h.service.ListCoupons()
	}

	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewError("Failed to list coupons"))
	}

	couponResponses := make([]entity.CouponResponse, len(coupons))
	for i, coupon := range coupons {
		couponResponses[i] = coupon.ToResponse()
	}

	response := ListCouponsResponse{
		Coupons: couponResponses,
		Total:   len(couponResponses),
	}

	return c.JSON(http.StatusOK, response)
}

// RedeemCoupon redeems a coupon for discount and/or reward items
func (h *Handler) RedeemCoupon(c echo.Context) error {
	var req RedeemCouponRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("Invalid request format", "invalid_request_error"))
	}

	if err := h.validator.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewValidationErrors(err))
	}

	response, err := h.service.RedeemCoupon(req)
	if err != nil {
		if errors.Is(err, ErrCouponNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewNotFoundError("Coupon"))
		}
		if errors.Is(err, ErrCouponNotUsable) {
			return c.JSON(http.StatusBadRequest, dto.NewError("Coupon is not usable (expired or inactive)", "invalid_request_error"))
		}
		if errors.Is(err, ErrInvalidOrderAmount) {
			return c.JSON(http.StatusBadRequest, dto.NewError("Invalid order amount", "invalid_request_error"))
		}
		// Handle specific error messages from service
		errorMsg := err.Error()
		if errorMsg == "coupon already used" {
			return c.JSON(http.StatusBadRequest, dto.NewError("Coupon has already been used", "invalid_request_error"))
		}
		if strings.Contains(errorMsg, "order amount does not meet minimum requirement") {
			return c.JSON(http.StatusBadRequest, dto.NewError(errorMsg, "invalid_request_error"))
		}
		if strings.Contains(errorMsg, "order amount is required") {
			return c.JSON(http.StatusBadRequest, dto.NewError(errorMsg, "invalid_request_error"))
		}
		if strings.Contains(errorMsg, "failed to grant reward items") {
			return c.JSON(http.StatusInternalServerError, dto.NewError("Failed to grant reward items"))
		}
		
		h.logger.Error("Failed to redeem coupon", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.NewError("Failed to redeem coupon"))
	}

	return c.JSON(http.StatusOK, response)
}

