package repository

import "fxserver/modules/item/entity"

type ItemRepository interface {
	// Item master data operations
	GetItem(id int) (*entity.Item, error)
	GetItems() ([]*entity.Item, error)
	GetItemsByType(itemType entity.ItemType) ([]*entity.Item, error)
	CreateItem(item *entity.Item) error
	UpdateItem(item *entity.Item) error
	DeleteItem(id int) error
}

type InventoryRepository interface {
	// User inventory operations
	GetUserInventory(userID int) ([]*entity.UserInventory, error)
	GetUserInventoryItem(userID, itemID int) (*entity.UserInventory, error)
	AddToInventory(userID, itemID int, count int, source string) error
	UpdateInventoryCount(userID, itemID int, count int) error
	RemoveFromInventory(userID, itemID int, count int) error
	
	// Batch operations for reward system
	AddMultipleToInventory(userID int, items []entity.RewardItem, source string) error
}

type Repository interface {
	ItemRepository
	InventoryRepository
}