package item

import (
	"errors"
	"net/http"
	"strconv"

	"fxserver/modules/item/entity"
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

// Public APIs

// GetItems godoc
// @Summary Get all items
// @Description Get list of all available items
// @Tags items
// @Accept json
// @Produce json
// @Param type query string false "Filter by item type"
// @Success 200 {object} ListItemsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/items [get]
func (h *Handler) GetItems(c echo.Context) error {
	itemType := c.QueryParam("type")
	
	var items []*entity.Item
	var err error

	if itemType != "" {
		if !entity.IsValidItemType(itemType) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: "Invalid item type",
			})
		}
		items, err = h.service.GetItemsByType(entity.ItemType(itemType))
	} else {
		items, err = h.service.GetItems()
	}

	if err != nil {
		h.logger.Error("Failed to get items", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to get items",
		})
	}

	itemResponses := make([]entity.ItemResponse, len(items))
	for i, item := range items {
		itemResponses[i] = item.ToResponse()
	}

	return c.JSON(http.StatusOK, ListItemsResponse{
		Items: itemResponses,
		Total: len(itemResponses),
	})
}

// GetItem godoc
// @Summary Get item by ID
// @Description Get specific item by ID
// @Tags items
// @Accept json
// @Produce json
// @Param id path int true "Item ID"
// @Success 200 {object} entity.ItemResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/items/{id} [get]
func (h *Handler) GetItem(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid item ID",
		})
	}

	item, err := h.service.GetItem(id)
	if err != nil {
		if errors.Is(err, ErrItemNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "Item not found",
			})
		}
		h.logger.Error("Failed to get item", zap.Error(err), zap.Int("item_id", id))
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to get item",
		})
	}

	return c.JSON(http.StatusOK, item.ToResponse())
}

// GetItemTypes godoc
// @Summary Get all item types
// @Description Get list of all available item types with descriptions
// @Tags items
// @Accept json
// @Produce json
// @Success 200 {object} ItemTypesResponse
// @Router /api/v1/items/types [get]
func (h *Handler) GetItemTypes(c echo.Context) error {
	return c.JSON(http.StatusOK, ItemTypesResponse{
		Types: h.service.GetItemTypes(),
	})
}

// GetUserInventory godoc
// @Summary Get user inventory
// @Description Get inventory for a specific user
// @Tags inventory
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} entity.UserInventoryResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/users/{id}/inventory [get]
func (h *Handler) GetUserInventory(c echo.Context) error {
	idParam := c.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid user ID",
		})
	}

	inventory, err := h.service.GetUserInventory(userID)
	if err != nil {
		h.logger.Error("Failed to get user inventory", zap.Error(err), zap.Int("user_id", userID))
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to get user inventory",
		})
	}

	return c.JSON(http.StatusOK, inventory)
}

// Admin APIs

// CreateItem godoc
// @Summary Create new item (Admin only)
// @Description Create a new item in the system
// @Tags admin,items
// @Accept json
// @Produce json
// @Param request body CreateItemRequest true "Create item request"
// @Success 201 {object} entity.ItemResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/items [post]
func (h *Handler) CreateItem(c echo.Context) error {
	var req CreateItemRequest
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

	item, err := h.service.CreateItem(req)
	if err != nil {
		if errors.Is(err, ErrInvalidItemType) || errors.Is(err, ErrInvalidRarity) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: err.Error(),
			})
		}
		h.logger.Error("Failed to create item", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to create item",
		})
	}

	return c.JSON(http.StatusCreated, item.ToResponse())
}

// UpdateItem godoc
// @Summary Update item (Admin only)
// @Description Update an existing item
// @Tags admin,items
// @Accept json
// @Produce json
// @Param id path int true "Item ID"
// @Param request body UpdateItemRequest true "Update item request"
// @Success 200 {object} entity.ItemResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/items/{id} [put]
func (h *Handler) UpdateItem(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid item ID",
		})
	}

	var req UpdateItemRequest
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

	item, err := h.service.UpdateItem(id, req)
	if err != nil {
		if errors.Is(err, ErrItemNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "Item not found",
			})
		}
		if errors.Is(err, ErrInvalidItemType) || errors.Is(err, ErrInvalidRarity) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: err.Error(),
			})
		}
		h.logger.Error("Failed to update item", zap.Error(err), zap.Int("item_id", id))
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to update item",
		})
	}

	return c.JSON(http.StatusOK, item.ToResponse())
}

// DeleteItem godoc
// @Summary Delete item (Admin only)
// @Description Soft delete an item (mark as inactive)
// @Tags admin,items
// @Accept json
// @Produce json
// @Param id path int true "Item ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/items/{id} [delete]
func (h *Handler) DeleteItem(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid item ID",
		})
	}

	if err := h.service.DeleteItem(id); err != nil {
		if errors.Is(err, ErrItemNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "Item not found",
			})
		}
		h.logger.Error("Failed to delete item", zap.Error(err), zap.Int("item_id", id))
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to delete item",
		})
	}

	return c.NoContent(http.StatusNoContent)
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