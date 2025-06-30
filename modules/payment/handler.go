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

// ProcessPayment godoc
// @Summary Process payment and grant items
// @Description Process payment and automatically grant reward items to user
// @Tags payments
// @Accept json
// @Produce json
// @Param request body CreatePaymentRequest true "Payment request"
// @Success 200 {object} ProcessPaymentResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/payments [post]
func (h *Handler) ProcessPayment(c echo.Context) error {
	var req CreatePaymentRequest
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

	response, err := h.service.ProcessPayment(req)
	if err != nil {
		if errors.Is(err, ErrInvalidPaymentMethod) || 
		   errors.Is(err, ErrInvalidAmount) ||
		   errors.Is(err, ErrPaymentAlreadyExists) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: err.Error(),
			})
		}
		h.logger.Error("Failed to process payment", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to process payment",
		})
	}

	return c.JSON(http.StatusOK, response)
}

// GetPayment godoc
// @Summary Get payment details
// @Description Get specific payment by ID
// @Tags payments
// @Accept json
// @Produce json
// @Param id path int true "Payment ID"
// @Success 200 {object} entity.PaymentResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/payments/{id} [get]
func (h *Handler) GetPayment(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid payment ID",
		})
	}

	payment, err := h.service.GetPayment(id)
	if err != nil {
		if errors.Is(err, ErrPaymentNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "Payment not found",
			})
		}
		h.logger.Error("Failed to get payment", zap.Error(err), zap.Int("payment_id", id))
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to get payment",
		})
	}

	return c.JSON(http.StatusOK, payment.ToResponse())
}

// GetUserPayments godoc
// @Summary Get user payment history
// @Description Get payment history for a specific user
// @Tags payments
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param status query string false "Filter by payment status"
// @Success 200 {object} entity.PaymentHistoryResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/users/{id}/payments [get]
func (h *Handler) GetUserPayments(c echo.Context) error {
	idParam := c.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid user ID",
		})
	}

	status := c.QueryParam("status")
	
	var history *entity.PaymentHistoryResponse
	
	if status != "" {
		if !entity.IsValidPaymentStatus(status) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: "Invalid payment status",
			})
		}
		history, err = h.service.GetUserPaymentsByStatus(userID, entity.PaymentStatus(status))
	} else {
		history, err = h.service.GetUserPayments(userID)
	}

	if err != nil {
		h.logger.Error("Failed to get user payments", zap.Error(err), zap.Int("user_id", userID))
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to get user payments",
		})
	}

	return c.JSON(http.StatusOK, history)
}

// GetUserPaymentSummary godoc
// @Summary Get user payment summary
// @Description Get payment statistics for a specific user
// @Tags payments
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} entity.PaymentSummaryResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/users/{id}/payments/summary [get]
func (h *Handler) GetUserPaymentSummary(c echo.Context) error {
	idParam := c.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid user ID",
		})
	}

	summary, err := h.service.GetPaymentSummaryByUser(userID)
	if err != nil {
		h.logger.Error("Failed to get user payment summary", zap.Error(err), zap.Int("user_id", userID))
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to get user payment summary",
		})
	}

	return c.JSON(http.StatusOK, summary)
}

// Public APIs

// GetPaymentMethods godoc
// @Summary Get available payment methods
// @Description Get list of all available payment methods
// @Tags payments
// @Accept json
// @Produce json
// @Success 200 {object} PaymentMethodsResponse
// @Router /api/v1/payments/methods [get]
func (h *Handler) GetPaymentMethods(c echo.Context) error {
	return c.JSON(http.StatusOK, PaymentMethodsResponse{
		Methods: h.service.GetPaymentMethods(),
	})
}

// GetPaymentStatuses godoc
// @Summary Get payment status types
// @Description Get list of all payment status types with descriptions
// @Tags payments
// @Accept json
// @Produce json
// @Success 200 {object} PaymentStatusesResponse
// @Router /api/v1/payments/statuses [get]
func (h *Handler) GetPaymentStatuses(c echo.Context) error {
	return c.JSON(http.StatusOK, PaymentStatusesResponse{
		Statuses: h.service.GetPaymentStatuses(),
	})
}

// Admin APIs

// UpdatePaymentStatus godoc
// @Summary Update payment status (Admin only)
// @Description Update payment status (used by payment webhooks or admin)
// @Tags admin,payments
// @Accept json
// @Produce json
// @Param id path int true "Payment ID"
// @Param request body UpdatePaymentStatusRequest true "Status update request"
// @Success 200 {object} entity.PaymentResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/payments/{id}/status [put]
func (h *Handler) UpdatePaymentStatus(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid payment ID",
		})
	}

	var req UpdatePaymentStatusRequest
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

	payment, err := h.service.UpdatePaymentStatus(id, req)
	if err != nil {
		if errors.Is(err, ErrPaymentNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "Payment not found",
			})
		}
		if errors.Is(err, ErrInvalidPaymentStatus) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: err.Error(),
			})
		}
		h.logger.Error("Failed to update payment status", zap.Error(err), zap.Int("payment_id", id))
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to update payment status",
		})
	}

	return c.JSON(http.StatusOK, payment.ToResponse())
}

// RefundPayment godoc
// @Summary Refund payment (Admin only)
// @Description Refund a completed payment
// @Tags admin,payments
// @Accept json
// @Produce json
// @Param id path int true "Payment ID"
// @Param request body RefundPaymentRequest true "Refund request"
// @Success 200 {object} entity.PaymentResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/payments/{id}/refund [post]
func (h *Handler) RefundPayment(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid payment ID",
		})
	}

	var req RefundPaymentRequest
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

	payment, err := h.service.RefundPayment(id, req)
	if err != nil {
		if errors.Is(err, ErrPaymentNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "Payment not found",
			})
		}
		if errors.Is(err, ErrCannotRefund) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: err.Error(),
			})
		}
		h.logger.Error("Failed to refund payment", zap.Error(err), zap.Int("payment_id", id))
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to refund payment",
		})
	}

	return c.JSON(http.StatusOK, payment.ToResponse())
}

// GetAllPayments godoc
// @Summary Get all payments (Admin only)
// @Description Get list of all payments with optional filters
// @Tags admin,payments
// @Accept json
// @Produce json
// @Param status query string false "Filter by payment status"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} entity.PaymentHistoryResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/payments [get]
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
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: "Invalid payment status",
			})
		}
		history, err = h.service.GetPaymentsByStatus(entity.PaymentStatus(status))
	} else {
		history, err = h.service.GetAllPayments()
	}

	if err != nil {
		h.logger.Error("Failed to get payments", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to get payments",
		})
	}

	return c.JSON(http.StatusOK, history)
}

// GetPaymentSummary godoc
// @Summary Get payment summary (Admin only)
// @Description Get overall payment statistics
// @Tags admin,payments
// @Accept json
// @Produce json
// @Success 200 {object} entity.PaymentSummaryResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/payments/summary [get]
func (h *Handler) GetPaymentSummary(c echo.Context) error {
	summary, err := h.service.GetPaymentSummary()
	if err != nil {
		h.logger.Error("Failed to get payment summary", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to get payment summary",
		})
	}

	return c.JSON(http.StatusOK, summary)
}

// Helper function to parse validation errors
func parseValidationErrors(err error) map[string]string {
	details := make(map[string]string)
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for field, message := range validationErrors.Errors {
			details[field] = message
		}
	} else {
		details["validation"] = err.Error()
	}
	return details
}