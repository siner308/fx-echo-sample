package user

import (
	"errors"
	"net/http"
	"strconv"

	userauth "fxserver/modules/auth/user"
	"fxserver/modules/user/entity"
	"fxserver/modules/user/repository"
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

func (h *Handler) CreateUser(c echo.Context) error {
	var req CreateUserRequest
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

	user, err := h.service.CreateUser(req)
	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			return c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error: "User with this email already exists",
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to create user",
		})
	}

	return c.JSON(http.StatusCreated, user.ToResponse())
}

func (h *Handler) GetUser(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid user ID",
		})
	}

	user, err := h.service.GetUser(id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "User not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to get user",
		})
	}

	return c.JSON(http.StatusOK, user.ToResponse())
}

func (h *Handler) GetMyInfo(c echo.Context) error {
	// Get user ID from JWT token context (set by middleware)
	userID, ok := userauth.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
		})
	}

	user, err := h.service.GetMyInfo(userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "User not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to get user info",
		})
	}

	return c.JSON(http.StatusOK, user.ToResponse())
}

func (h *Handler) UpdateUser(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid user ID",
		})
	}

	var req UpdateUserRequest
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

	user, err := h.service.UpdateUser(id, req)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "User not found",
			})
		}
		if errors.Is(err, repository.ErrUserExists) {
			return c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error: "User with this email already exists",
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to update user",
		})
	}

	return c.JSON(http.StatusOK, user.ToResponse())
}

func (h *Handler) DeleteUser(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid user ID",
		})
	}

	err = h.service.DeleteUser(id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "User not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to delete user",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) ListUsers(c echo.Context) error {
	users, err := h.service.ListUsers()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to list users",
		})
	}

	userResponses := make([]entity.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	response := ListUsersResponse{
		Users: userResponses,
		Total: len(userResponses),
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