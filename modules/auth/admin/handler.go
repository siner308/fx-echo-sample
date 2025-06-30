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
			return c.JSON(http.StatusServiceUnavailable, dto.NewError("Keycloak SSO service unavailable"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("Failed to generate auth URL"))
	}

	return c.JSON(http.StatusOK, KeycloakAuthURLResponse{
		AuthURL: authURL,
	})
}

// HandleKeycloakCallback handles the OAuth2 callback from Keycloak
func (h *Handler) HandleKeycloakCallback(c echo.Context) error {
	var req KeycloakCallbackRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("Invalid request format"))
	}

	if err := h.validator.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewValidationErrors(err))
	}

	response, err := h.service.HandleKeycloakCallback(c.Request().Context(), req.Code)
	if err != nil {
		if errors.Is(err, ErrKeycloakUnavailable) {
			return c.JSON(http.StatusServiceUnavailable, dto.NewError("Keycloak SSO service unavailable"))
		}
		if errors.Is(err, ErrInvalidAuthCode) || errors.Is(err, ErrTokenExchange) {
			return c.JSON(http.StatusBadRequest, dto.NewError("Invalid authorization code"))
		}
		if errors.Is(err, ErrNotAdminUser) {
			return c.JSON(http.StatusForbidden, dto.NewError("Insufficient permissions - admin access required"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewError("Failed to complete SSO login"))
	}

	return c.JSON(http.StatusOK, response)
}

// GetAdminInfo returns the current admin user's information
func (h *Handler) GetAdminInfo(c echo.Context) error {
	// Get admin info from context (set by middleware)
	adminEmail, ok := GetAdminEmail(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.NewError("Admin user not found in context"))
	}

	adminID, ok := GetAdminID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.NewError("Admin user ID not found in context"))
	}

	// For simplicity, return admin info from JWT claims
	// In production, you might want to fetch full admin info from Keycloak
	adminRole, _ := GetAdminRole(c)
	tokenType, _ := GetTokenType(c)

	adminInfo := map[string]interface{}{
		"id":    adminID,
		"email": adminEmail,
		"role":  adminRole,
		"token_type": tokenType,
	}

	return c.JSON(http.StatusOK, adminInfo)
}


