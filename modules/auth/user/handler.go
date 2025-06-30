package user

import (
	"net/http"

	"fxserver/pkg/dto"
	"fxserver/pkg/validator"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Handler struct {
	authService Service
	validator   validator.Validator
	logger      *zap.Logger
}

type HandlerParam struct {
	fx.In
	AuthService Service
	Validator   validator.Validator
	Logger      *zap.Logger
}

func NewHandler(p HandlerParam) *Handler {
	return &Handler{
		authService: p.AuthService,
		validator:   p.Validator,
		logger:      p.Logger,
	}
}

// Login godoc
// @Summary User login
// @Description Authenticate user and return tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *Handler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("Invalid request format", "invalid_request_error"))
	}

	if err := h.validator.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewValidationErrors(err))
	}

	response, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		if err == ErrInvalidCredentials {
			return c.JSON(http.StatusUnauthorized, dto.NewAuthError("Invalid email or password"))
		}
		h.logger.Error("Login failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.NewError("Login failed"))
	}

	return c.JSON(http.StatusOK, response)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} RefreshResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/auth/refresh [post]
func (h *Handler) RefreshToken(c echo.Context) error {
	var req RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("Invalid request format", "invalid_request_error"))
	}

	if err := h.validator.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewValidationErrors(err))
	}

	response, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		if err == ErrInvalidRefreshToken {
			return c.JSON(http.StatusUnauthorized, dto.NewAuthError("Invalid refresh token"))
		}
		h.logger.Error("Token refresh failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.NewError("Token refresh failed"))
	}

	return c.JSON(http.StatusOK, response)
}

