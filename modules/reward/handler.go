package reward

import (
	"net/http"

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

// GrantReward godoc
// @Summary Grant reward to user (Admin only)
// @Description Grant items to a specific user
// @Tags admin,rewards
// @Accept json
// @Produce json
// @Param request body GrantRewardRequest true "Grant reward request"
// @Success 200 {object} GrantRewardResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/rewards/grant [post]
func (h *Handler) GrantReward(c echo.Context) error {
	var req GrantRewardRequest
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

	response, err := h.service.GrantRewards(req)
	if err != nil {
		// Even if there's an error, we might have a partial response
		if response != nil && !response.Success {
			return c.JSON(http.StatusBadRequest, response)
		}
		
		h.logger.Error("Failed to grant reward", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to grant reward",
		})
	}

	return c.JSON(http.StatusOK, response)
}

// BulkGrantReward godoc
// @Summary Grant rewards to multiple users (Admin only)
// @Description Grant items to multiple users at once
// @Tags admin,rewards
// @Accept json
// @Produce json
// @Param request body BulkGrantRewardRequest true "Bulk grant reward request"
// @Success 200 {object} BulkGrantRewardResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/rewards/bulk-grant [post]
func (h *Handler) BulkGrantReward(c echo.Context) error {
	var req BulkGrantRewardRequest
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

	// Additional validation for bulk operations
	if len(req.UserIDs) > 1000 {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Cannot grant rewards to more than 1000 users at once",
		})
	}

	response, err := h.service.BulkGrantRewards(req)
	if err != nil {
		h.logger.Error("Failed to bulk grant rewards", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to bulk grant rewards",
		})
	}

	// Return 207 Multi-Status if there were partial failures
	statusCode := http.StatusOK
	if response.FailureCount > 0 && response.SuccessCount > 0 {
		statusCode = http.StatusMultiStatus
	} else if response.FailureCount > 0 && response.SuccessCount == 0 {
		statusCode = http.StatusBadRequest
	}

	return c.JSON(statusCode, response)
}

// GetRewardSources godoc
// @Summary Get available reward sources
// @Description Get list of all available reward sources with descriptions
// @Tags rewards
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/v1/rewards/sources [get]
func (h *Handler) GetRewardSources(c echo.Context) error {
	sources := map[string]string{
		RewardSourceAdmin:        GetRewardSourceDescription(RewardSourceAdmin),
		RewardSourceCoupon:       GetRewardSourceDescription(RewardSourceCoupon),
		RewardSourcePayment:      GetRewardSourceDescription(RewardSourcePayment),
		RewardSourceEvent:        GetRewardSourceDescription(RewardSourceEvent),
		RewardSourceCompensation: GetRewardSourceDescription(RewardSourceCompensation),
		RewardSourceDaily:        GetRewardSourceDescription(RewardSourceDaily),
		RewardSourceAchievement:  GetRewardSourceDescription(RewardSourceAchievement),
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"sources": sources,
	})
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