package item

import (
	"errors"
	"fmt"

	"fxserver/modules/item/entity"
	"fxserver/modules/item/repository"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var (
	ErrItemNotFound     = errors.New("item not found")
	ErrInvalidItemType  = errors.New("invalid item type")
	ErrInvalidRarity    = errors.New("invalid rarity")
	ErrInsufficientItem = errors.New("insufficient item count")
)

type Service interface {
	// Item master operations (Admin)
	CreateItem(req CreateItemRequest) (*entity.Item, error)
	UpdateItem(id int, req UpdateItemRequest) (*entity.Item, error)
	DeleteItem(id int) error
	GetItem(id int) (*entity.Item, error)
	GetItems() ([]*entity.Item, error)
	GetItemsByType(itemType entity.ItemType) ([]*entity.Item, error)

	// Inventory operations
	GetUserInventory(userID int) (*entity.UserInventoryResponse, error)
	AddToInventory(userID, itemID int, count int, source string) error
	RemoveFromInventory(userID, itemID int, count int) error
	AddMultipleToInventory(userID int, items []entity.RewardItem, source string) error

	// Utility
	GetItemTypes() []ItemTypeInfo
}

type service struct {
	repository repository.Repository
	logger     *zap.Logger
}

type ServiceParam struct {
	fx.In
	Repository repository.Repository
	Logger     *zap.Logger
}

func NewService(p ServiceParam) Service {
	return &service{
		repository: p.Repository,
		logger:     p.Logger,
	}
}

// Item master operations
func (s *service) CreateItem(req CreateItemRequest) (*entity.Item, error) {
	// Validate item type
	if !entity.IsValidItemType(string(req.Type)) {
		return nil, ErrInvalidItemType
	}

	// Validate rarity
	if !entity.IsValidRarity(req.Rarity) {
		return nil, ErrInvalidRarity
	}

	item := &entity.Item{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Value:       req.Value,
		Rarity:      req.Rarity,
		IconURL:     req.IconURL,
		IsActive:    true,
	}

	if err := s.repository.CreateItem(item); err != nil {
		s.logger.Error("Failed to create item", zap.Error(err), zap.String("name", req.Name))
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	s.logger.Info("Item created successfully", 
		zap.Int("item_id", item.ID), 
		zap.String("name", item.Name),
		zap.String("type", string(item.Type)))

	return item, nil
}

func (s *service) UpdateItem(id int, req UpdateItemRequest) (*entity.Item, error) {
	// Get existing item
	item, err := s.repository.GetItem(id)
	if err != nil {
		return nil, ErrItemNotFound
	}

	// Update fields if provided
	if req.Name != "" {
		item.Name = req.Name
	}
	if req.Description != "" {
		item.Description = req.Description
	}
	if req.Type != "" {
		if !entity.IsValidItemType(string(req.Type)) {
			return nil, ErrInvalidItemType
		}
		item.Type = req.Type
	}
	if req.Value > 0 {
		item.Value = req.Value
	}
	if req.Rarity != "" {
		if !entity.IsValidRarity(req.Rarity) {
			return nil, ErrInvalidRarity
		}
		item.Rarity = req.Rarity
	}
	if req.IconURL != "" {
		item.IconURL = req.IconURL
	}

	if err := s.repository.UpdateItem(item); err != nil {
		s.logger.Error("Failed to update item", zap.Error(err), zap.Int("item_id", id))
		return nil, fmt.Errorf("failed to update item: %w", err)
	}

	s.logger.Info("Item updated successfully", zap.Int("item_id", item.ID))
	return item, nil
}

func (s *service) DeleteItem(id int) error {
	// Check if item exists
	_, err := s.repository.GetItem(id)
	if err != nil {
		return ErrItemNotFound
	}

	if err := s.repository.DeleteItem(id); err != nil {
		s.logger.Error("Failed to delete item", zap.Error(err), zap.Int("item_id", id))
		return fmt.Errorf("failed to delete item: %w", err)
	}

	s.logger.Info("Item deleted successfully", zap.Int("item_id", id))
	return nil
}

func (s *service) GetItem(id int) (*entity.Item, error) {
	item, err := s.repository.GetItem(id)
	if err != nil {
		return nil, ErrItemNotFound
	}
	return item, nil
}

func (s *service) GetItems() ([]*entity.Item, error) {
	return s.repository.GetItems()
}

func (s *service) GetItemsByType(itemType entity.ItemType) ([]*entity.Item, error) {
	if !entity.IsValidItemType(string(itemType)) {
		return nil, ErrInvalidItemType
	}
	return s.repository.GetItemsByType(itemType)
}

// Inventory operations
func (s *service) GetUserInventory(userID int) (*entity.UserInventoryResponse, error) {
	inventories, err := s.repository.GetUserInventory(userID)
	if err != nil {
		s.logger.Error("Failed to get user inventory", zap.Error(err), zap.Int("user_id", userID))
		return nil, fmt.Errorf("failed to get user inventory: %w", err)
	}

	var inventoryResponses []entity.InventoryResponse
	for _, inv := range inventories {
		item, err := s.repository.GetItem(inv.ItemID)
		if err != nil {
			s.logger.Warn("Item not found for inventory", 
				zap.Int("item_id", inv.ItemID), 
				zap.Int("user_id", userID))
			continue
		}
		inventoryResponses = append(inventoryResponses, inv.ToResponse(item))
	}

	return &entity.UserInventoryResponse{
		UserID: userID,
		Items:  inventoryResponses,
		Total:  len(inventoryResponses),
	}, nil
}

func (s *service) AddToInventory(userID, itemID int, count int, source string) error {
	// Verify item exists
	_, err := s.repository.GetItem(itemID)
	if err != nil {
		return ErrItemNotFound
	}

	if count <= 0 {
		return errors.New("count must be greater than 0")
	}

	if err := s.repository.AddToInventory(userID, itemID, count, source); err != nil {
		s.logger.Error("Failed to add item to inventory", 
			zap.Error(err), 
			zap.Int("user_id", userID), 
			zap.Int("item_id", itemID),
			zap.Int("count", count))
		return fmt.Errorf("failed to add item to inventory: %w", err)
	}

	s.logger.Info("Item added to inventory", 
		zap.Int("user_id", userID), 
		zap.Int("item_id", itemID),
		zap.Int("count", count),
		zap.String("source", source))

	return nil
}

func (s *service) RemoveFromInventory(userID, itemID int, count int) error {
	if count <= 0 {
		return errors.New("count must be greater than 0")
	}

	if err := s.repository.RemoveFromInventory(userID, itemID, count); err != nil {
		s.logger.Error("Failed to remove item from inventory", 
			zap.Error(err), 
			zap.Int("user_id", userID), 
			zap.Int("item_id", itemID),
			zap.Int("count", count))
		return fmt.Errorf("failed to remove item from inventory: %w", err)
	}

	s.logger.Info("Item removed from inventory", 
		zap.Int("user_id", userID), 
		zap.Int("item_id", itemID),
		zap.Int("count", count))

	return nil
}

func (s *service) AddMultipleToInventory(userID int, items []entity.RewardItem, source string) error {
	if len(items) == 0 {
		return errors.New("no items to add")
	}

	// Validate all items and counts
	for _, item := range items {
		if item.Count <= 0 {
			return fmt.Errorf("invalid count %d for item %d", item.Count, item.ItemID)
		}
		// Verify item exists
		if _, err := s.repository.GetItem(item.ItemID); err != nil {
			return fmt.Errorf("item %d not found", item.ItemID)
		}
	}

	if err := s.repository.AddMultipleToInventory(userID, items, source); err != nil {
		s.logger.Error("Failed to add multiple items to inventory", 
			zap.Error(err), 
			zap.Int("user_id", userID), 
			zap.Int("item_count", len(items)))
		return fmt.Errorf("failed to add multiple items to inventory: %w", err)
	}

	s.logger.Info("Multiple items added to inventory", 
		zap.Int("user_id", userID), 
		zap.Int("item_count", len(items)),
		zap.String("source", source))

	return nil
}

func (s *service) GetItemTypes() []ItemTypeInfo {
	return GetItemTypes()
}