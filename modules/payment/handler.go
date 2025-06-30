package payment

import (
	"errors"
	"net/http"
	"strconv"

	"fxserver/modules/payment/entity"
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

// User APIs

// ProcessPayment processes payment and grants items to user
func (h *Handler) ProcessPayment(c echo.Context) error {
	var req CreatePaymentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("Invalid request format", "invalid_request_error"))
	}

	if err := h.validator.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewValidationErrors(err))
	}

	response, err := h.service.ProcessPayment(req)
	if err != nil {
		if errors.Is(err, ErrInvalidPaymentMethod) ||
			errors.Is(err, ErrInvalidAmount) ||
			errors.Is(err, ErrPaymentAlreadyExists) {
			return c.JSON(http.StatusBadRequest, dto.NewError(err.Error(), "invalid_request_error"))
		}
		h.logger.Error("Failed to process payment", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.NewError("Failed to process payment"))
	}

	return c.JSON(http.StatusOK, response)
}

// GetPayment retrieves payment details by ID
func (h *Handler) GetPayment(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("Invalid payment ID", "invalid_request_error"))
	}

	payment, err := h.service.GetPayment(id)
	if err != nil {
		if errors.Is(err, ErrPaymentNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewNotFoundError("Payment"))
		}
		h.logger.Error("Failed to get payment", zap.Error(err), zap.Int("payment_id", id))
		return c.JSON(http.StatusInternalServerError, dto.NewError("Failed to get payment"))
	}

	return c.JSON(http.StatusOK, payment.ToResponse())
}

// GetUserPayments retrieves payment history for a specific user
func (h *Handler) GetUserPayments(c echo.Context) error {
	idParam := c.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("Invalid user ID", "invalid_request_error"))
	}

	status := c.QueryParam("status")

	var history *entity.PaymentHistoryResponse

	if status != "" {
		if !entity.IsValidPaymentStatus(status) {
			return c.JSON(http.StatusBadRequest, dto.NewError("Invalid payment status", "invalid_request_error"))
		}
		history, err = h.service.GetUserPaymentsByStatus(userID, entity.PaymentStatus(status))
	} else {
		history, err = h.service.GetUserPayments(userID)
	}

	if err != nil {
		h.logger.Error("Failed to get user payments", zap.Error(err), zap.Int("user_id", userID))
		return c.JSON(http.StatusInternalServerError, dto.NewError("Failed to get user payments"))
	}

	return c.JSON(http.StatusOK, history)
}

// GetUserPaymentSummary retrieves payment statistics for a specific user
func (h *Handler) GetUserPaymentSummary(c echo.Context) error {
	idParam := c.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("Invalid user ID", "invalid_request_error"))
	}

	summary, err := h.service.GetPaymentSummaryByUser(userID)
	if err != nil {
		h.logger.Error("Failed to get user payment summary", zap.Error(err), zap.Int("user_id", userID))
		return c.JSON(http.StatusInternalServerError, dto.NewError("Failed to get user payment summary"))
	}

	return c.JSON(http.StatusOK, summary)
}

// Public APIs

// GetPaymentMethods returns all available payment methods
func (h *Handler) GetPaymentMethods(c echo.Context) error {
	return c.JSON(http.StatusOK, PaymentMethodsResponse{
		Methods: h.service.GetPaymentMethods(),
	})
}

// GetPaymentStatuses returns all payment status types with descriptions
func (h *Handler) GetPaymentStatuses(c echo.Context) error {
	return c.JSON(http.StatusOK, PaymentStatusesResponse{
		Statuses: h.service.GetPaymentStatuses(),
	})
}

// Admin APIs

// UpdatePaymentStatus updates payment status (admin only)
func (h *Handler) UpdatePaymentStatus(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("Invalid payment ID", "invalid_request_error"))
	}

	var req UpdatePaymentStatusRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("Invalid request format", "invalid_request_error"))
	}

	if err := h.validator.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewValidationErrors(err))
	}

	payment, err := h.service.UpdatePaymentStatus(id, req)
	if err != nil {
		if errors.Is(err, ErrPaymentNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewNotFoundError("Payment"))
		}
		if errors.Is(err, ErrInvalidPaymentStatus) {
			return c.JSON(http.StatusBadRequest, dto.NewError(err.Error(), "invalid_request_error"))
		}
		h.logger.Error("Failed to update payment status", zap.Error(err), zap.Int("payment_id", id))
		return c.JSON(http.StatusInternalServerError, dto.NewError("Failed to update payment status"))
	}

	return c.JSON(http.StatusOK, payment.ToResponse())
}

// RefundPayment refunds a completed payment (admin only)
func (h *Handler) RefundPayment(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("Invalid payment ID", "invalid_request_error"))
	}

	var req RefundPaymentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("Invalid request format", "invalid_request_error"))
	}

	if err := h.validator.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewValidationErrors(err))
	}

	payment, err := h.service.RefundPayment(id, req)
	if err != nil {
		if errors.Is(err, ErrPaymentNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewNotFoundError("Payment"))
		}
		if errors.Is(err, ErrCannotRefund) {
			return c.JSON(http.StatusBadRequest, dto.NewError(err.Error(), "invalid_request_error"))
		}
		h.logger.Error("Failed to refund payment", zap.Error(err), zap.Int("payment_id", id))
		return c.JSON(http.StatusInternalServerError, dto.NewError("Failed to refund payment"))
	}

	return c.JSON(http.StatusOK, payment.ToResponse())
}

// GetAllPayments retrieves all payments with optional filters (admin only)
func (h *Handler) GetAllPayments(c echo.Context) error {
	status := c.QueryParam("status")
	startDate := c.QueryParam("start_date")
	endDate := c.QueryParam("end_date")

	var history *entity.PaymentHistoryResponse
	var err error

	if startDate != "" && endDate != "" {
		history, err = h.service.GetPaymentsByDateRange(startDate, endDate)
	} else if status != "" {
		if !entity.IsValidPaymentStatus(status) {
			return c.JSON(http.StatusBadRequest, dto.NewError("Invalid payment status", "invalid_request_error"))
		}
		history, err = h.service.GetPaymentsByStatus(entity.PaymentStatus(status))
	} else {
		history, err = h.service.GetAllPayments()
	}

	if err != nil {
		h.logger.Error("Failed to get payments", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.NewError("Failed to get payments"))
	}

	return c.JSON(http.StatusOK, history)
}

// GetPaymentSummary retrieves overall payment statistics (admin only)
func (h *Handler) GetPaymentSummary(c echo.Context) error {
	summary, err := h.service.GetPaymentSummary()
	if err != nil {
		h.logger.Error("Failed to get payment summary", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.NewError("Failed to get payment summary"))
	}

	return c.JSON(http.StatusOK, summary)
}

