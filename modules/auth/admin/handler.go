package admin

import (
	"errors"
	"net/http"

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

// GetKeycloakAuthURL returns the Keycloak SSO login URL
func (h *Handler) GetKeycloakAuthURL(c echo.Context) error {
	authURL, err := h.service.GetKeycloakAuthURL()
	if err != nil {
		if errors.Is(err, ErrKeycloakUnavailable) {
			return c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{
				Error: "Keycloak SSO service unavailable",
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to generate auth URL",
		})
	}

	return c.JSON(http.StatusOK, KeycloakAuthURLResponse{
		AuthURL: authURL,
	})
}

// HandleKeycloakCallback handles the OAuth2 callback from Keycloak
func (h *Handler) HandleKeycloakCallback(c echo.Context) error {
	var req KeycloakCallbackRequest
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

	response, err := h.service.HandleKeycloakCallback(c.Request().Context(), req.Code)
	if err != nil {
		if errors.Is(err, ErrKeycloakUnavailable) {
			return c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{
				Error: "Keycloak SSO service unavailable",
			})
		}
		if errors.Is(err, ErrInvalidAuthCode) || errors.Is(err, ErrTokenExchange) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: "Invalid authorization code",
			})
		}
		if errors.Is(err, ErrNotAdminUser) {
			return c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error: "Insufficient permissions - admin access required",
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to complete SSO login",
		})
	}

	return c.JSON(http.StatusOK, response)
}

// Legacy admin login (deprecated)
func (h *Handler) AdminLogin(c echo.Context) error {
	return c.JSON(http.StatusGone, dto.ErrorResponse{
		Error: "Legacy admin login is deprecated. Please use Keycloak SSO instead.",
		Details: map[string]string{
			"sso_auth_url": "/auth/admin/sso/auth-url",
			"callback_url": "/auth/admin/sso/callback",
		},
	})
}

func parseValidationErrors(err error) map[string]string {
	details := make(map[string]string)
	if err != nil {
		details["validation"] = err.Error()
	}
	return details
}