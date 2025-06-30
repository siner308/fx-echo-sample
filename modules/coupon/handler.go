package coupon

import (
	"errors"
	"net/http"
	"strconv"

	"fxserver/modules/coupon/entity"
	"fxserver/modules/coupon/repository"
	"fxserver/pkg/dto"
	"fxserver/pkg/validator"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Handler struct {
	service   Service
	validator *validator.Validator
	logger    *zap.Logger
}

func NewHandler(service Service, validator *validator.Validator, logger *zap.Logger) *Handler {
	return &Handler{
		service:   service,
		validator: validator,
		logger:    logger,
	}
}

func (h *Handler) CreateCoupon(c echo.Context) error {
	var req CreateCouponRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request format",
		})
	}

	if err := h.validator.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Validation failed",
			Details: parseValidationErrors(err),
		})
	}

	coupon, err := h.service.CreateCoupon(req)
	if err != nil {
		if errors.Is(err, repository.ErrCouponExists) {
			return c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error: "Coupon with this code already exists",
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to create coupon",
		})
	}

	return c.JSON(http.StatusCreated, coupon.ToResponse())
}

func (h *Handler) GetCoupon(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid coupon ID",
		})
	}

	coupon, err := h.service.GetCoupon(id)
	if err != nil {
		if errors.Is(err, repository.ErrCouponNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "Coupon not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to get coupon",
		})
	}

	return c.JSON(http.StatusOK, coupon.ToResponse())
}

func (h *Handler) UpdateCoupon(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid coupon ID",
		})
	}

	var req UpdateCouponRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request format",
		})
	}

	if err := h.validator.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Validation failed",
			Details: parseValidationErrors(err),
		})
	}

	coupon, err := h.service.UpdateCoupon(id, req)
	if err != nil {
		if errors.Is(err, repository.ErrCouponNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "Coupon not found",
			})
		}
		if errors.Is(err, repository.ErrCouponExists) {
			return c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error: "Coupon with this code already exists",
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to update coupon",
		})
	}

	return c.JSON(http.StatusOK, coupon.ToResponse())
}

func (h *Handler) DeleteCoupon(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid coupon ID",
		})
	}

	err = h.service.DeleteCoupon(id)
	if err != nil {
		if errors.Is(err, repository.ErrCouponNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "Coupon not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to delete coupon",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

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
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to list coupons",
		})
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

func (h *Handler) UseCoupon(c echo.Context) error {
	var req UseCouponRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request format",
		})
	}

	if err := h.validator.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Validation failed",
			Details: parseValidationErrors(err),
		})
	}

	response, err := h.service.UseCoupon(req)
	if err != nil {
		if errors.Is(err, repository.ErrCouponNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "Coupon not found",
			})
		}
		if errors.Is(err, repository.ErrCouponNotUsable) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: "Coupon is not usable (expired or inactive)",
			})
		}
		if errors.Is(err, repository.ErrCouponAlreadyUsed) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: "Coupon has already been used",
			})
		}
		if errors.Is(err, ErrInvalidOrderAmount) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: "Invalid order amount",
			})
		}
		if err.Error() == "order amount does not meet minimum requirement" {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: "Order amount does not meet minimum requirement",
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to use coupon",
		})
	}

	return c.JSON(http.StatusOK, response)
}

func parseValidationErrors(err error) map[string]string {
	details := make(map[string]string)
	if err != nil {
		details["validation"] = err.Error()
	}
	return details
}
