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

// GetItems returns all items, optionally filtered by type
func (h *Handler) GetItems(c echo.Context) error {
	itemType := c.QueryParam("type")
	
	var items []*entity.Item
	var err error

	if itemType != "" {
		if !entity.IsValidItemType(itemType) {
			return c.JSON(http.StatusBadRequest, dto.NewError("Invalid item type", "invalid_request_error"))
		}
		items, err = h.service.GetItemsByType(entity.ItemType(itemType))
	} else {
		items, err = h.service.GetItems()
	}

	if err != nil {
		h.logger.Error("Failed to get items", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.NewError("Failed to get items"))
	}

	itemResponses := make([]entity.ItemResponse, len(items))
	for i, item := range items {
		itemResponses[i] = item.ToResponse()
	}

	return c.JSON(http.StatusOK, dto.NewList(itemResponses))
}

// GetItem returns a specific item by ID
func (h *Handler) GetItem(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("Invalid item ID", "invalid_request_error"))
	}

	item, err := h.service.GetItem(id)
	if err != nil {
		if errors.Is(err, ErrItemNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewNotFoundError("Item"))
		}
		h.logger.Error("Failed to get item", zap.Error(err), zap.Int("item_id", id))
		return c.JSON(http.StatusInternalServerError, dto.NewError("Failed to get item"))
	}

	return c.JSON(http.StatusOK, item.ToResponse())
}

// GetItemTypes returns all available item types
func (h *Handler) GetItemTypes(c echo.Context) error {
	return c.JSON(http.StatusOK, ItemTypesResponse{
		Types: h.service.GetItemTypes(),
	})
}

// GetUserInventory returns inventory for a specific user
func (h *Handler) GetUserInventory(c echo.Context) error {
	idParam := c.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("Invalid user ID", "invalid_request_error"))
	}

	inventory, err := h.service.GetUserInventory(userID)
	if err != nil {
		h.logger.Error("Failed to get user inventory", zap.Error(err), zap.Int("user_id", userID))
		return c.JSON(http.StatusInternalServerError, dto.NewError("Failed to get user inventory"))
	}

	return c.JSON(http.StatusOK, inventory)
}

// Admin APIs

// CreateItem creates a new item (admin only)
func (h *Handler) CreateItem(c echo.Context) error {
	var req CreateItemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("Invalid request format", "invalid_request_error"))
	}

	if err := h.validator.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewValidationErrors(err))
	}

	item, err := h.service.CreateItem(req)
	if err != nil {
		if errors.Is(err, ErrInvalidItemType) || errors.Is(err, ErrInvalidRarity) {
			return c.JSON(http.StatusBadRequest, dto.NewError(err.Error()))
		}
		h.logger.Error("Failed to create item", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, dto.NewError("Failed to create item"))
	}

	return c.JSON(http.StatusCreated, item.ToResponse())
}

// UpdateItem updates an existing item (admin only)
func (h *Handler) UpdateItem(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("Invalid item ID", "invalid_request_error"))
	}

	var req UpdateItemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("Invalid request format", "invalid_request_error"))
	}

	if err := h.validator.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewValidationErrors(err))
	}

	item, err := h.service.UpdateItem(id, req)
	if err != nil {
		if errors.Is(err, ErrItemNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewNotFoundError("Item"))
		}
		if errors.Is(err, ErrInvalidItemType) || errors.Is(err, ErrInvalidRarity) {
			return c.JSON(http.StatusBadRequest, dto.NewError(err.Error()))
		}
		h.logger.Error("Failed to update item", zap.Error(err), zap.Int("item_id", id))
		return c.JSON(http.StatusInternalServerError, dto.NewError("Failed to update item"))
	}

	return c.JSON(http.StatusOK, item.ToResponse())
}

// DeleteItem soft deletes an item (admin only)
func (h *Handler) DeleteItem(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewError("Invalid item ID", "invalid_request_error"))
	}

	if err := h.service.DeleteItem(id); err != nil {
		if errors.Is(err, ErrItemNotFound) {
			return c.JSON(http.StatusNotFound, dto.NewNotFoundError("Item"))
		}
		h.logger.Error("Failed to delete item", zap.Error(err), zap.Int("item_id", id))
		return c.JSON(http.StatusInternalServerError, dto.NewError("Failed to delete item"))
	}

	return c.JSON(http.StatusOK, dto.NewEmpty(strconv.Itoa(id)))
}

