package auth

import (
	"errors"
	"net/http"

	"fxserver/pkg/dto"
	"fxserver/pkg/jwt"
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

func (h *Handler) Login(c echo.Context) error {
	var req LoginRequest
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

	response, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) || errors.Is(err, ErrInvalidCredentials) {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: "Invalid email or password",
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to login",
		})
	}

	return c.JSON(http.StatusOK, response)
}

func (h *Handler) RefreshToken(c echo.Context) error {
	var req RefreshTokenRequest
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

	response, err := h.service.RefreshToken(req.RefreshToken)
	if err != nil {
		if errors.Is(err, ErrInvalidRefreshToken) {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: "Invalid refresh token",
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to refresh token",
		})
	}

	return c.JSON(http.StatusOK, response)
}

func (h *Handler) AdminLogin(c echo.Context) error {
	var req AdminLoginRequest
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

	response, err := h.service.AdminLogin(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) || errors.Is(err, ErrInvalidCredentials) {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: "Invalid admin credentials",
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to admin login",
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