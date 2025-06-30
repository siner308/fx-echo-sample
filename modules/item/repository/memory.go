package repository

import (
	"fmt"
	"sync"
	"time"

	"fxserver/modules/item/entity"
)

type memoryRepository struct {
	items        map[int]*entity.Item
	inventories  map[string]*entity.UserInventory // key: "userID:itemID"
	itemCounter  int
	invCounter   int
	mu           sync.RWMutex
}

func NewMemoryRepository() Repository {
	repo := &memoryRepository{
		items:       make(map[int]*entity.Item),
		inventories: make(map[string]*entity.UserInventory),
		itemCounter: 0,
		invCounter:  0,
	}
	
	// Initialize with some default items
	repo.initializeDefaultItems()
	
	return repo
}

func (r *memoryRepository) initializeDefaultItems() {
	defaultItems := []*entity.Item{
		{
			Name:        "골드",
			Description: "기본 게임 화폐",
			Type:        entity.ItemTypeCurrency,
			Value:       1,
			Rarity:      "common",
			IconURL:     "/icons/gold.png",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "다이아몬드",
			Description: "프리미엄 화폐",
			Type:        entity.ItemTypeCurrency,
			Value:       1,
			Rarity:      "epic",
			IconURL:     "/icons/diamond.png",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "체력 포션",
			Description: "HP를 회복하는 포션",
			Type:        entity.ItemTypeConsumable,
			Value:       1,
			Rarity:      "common",
			IconURL:     "/icons/hp_potion.png",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "전설의 검",
			Description: "강력한 공격력을 가진 전설 등급 검",
			Type:        entity.ItemTypeEquipment,
			Value:       1,
			Rarity:      "legendary",
			IconURL:     "/icons/legendary_sword.png",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "던전 입장권",
			Description: "던전에 입장할 수 있는 티켓",
			Type:        entity.ItemTypeTicket,
			Value:       1,
			Rarity:      "common",
			IconURL:     "/icons/dungeon_ticket.png",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, item := range defaultItems {
		r.itemCounter++
		item.ID = r.itemCounter
		r.items[item.ID] = item
	}
}

// ItemRepository implementation
func (r *memoryRepository) GetItem(id int) (*entity.Item, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, exists := r.items[id]
	if !exists {
		return nil, fmt.Errorf("item with id %d not found", id)
	}
	return item, nil
}

func (r *memoryRepository) GetItems() ([]*entity.Item, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]*entity.Item, 0, len(r.items))
	for _, item := range r.items {
		if item.IsActive {
			items = append(items, item)
		}
	}
	return items, nil
}

func (r *memoryRepository) GetItemsByType(itemType entity.ItemType) ([]*entity.Item, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []*entity.Item
	for _, item := range r.items {
		if item.IsActive && item.Type == itemType {
			items = append(items, item)
		}
	}
	return items, nil
}

func (r *memoryRepository) CreateItem(item *entity.Item) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.itemCounter++
	item.ID = r.itemCounter
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()
	r.items[item.ID] = item
	return nil
}

func (r *memoryRepository) UpdateItem(item *entity.Item) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.items[item.ID]; !exists {
		return fmt.Errorf("item with id %d not found", item.ID)
	}

	item.UpdatedAt = time.Now()
	r.items[item.ID] = item
	return nil
}

func (r *memoryRepository) DeleteItem(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, exists := r.items[id]
	if !exists {
		return fmt.Errorf("item with id %d not found", id)
	}

	item.IsActive = false
	item.UpdatedAt = time.Now()
	return nil
}

// InventoryRepository implementation
func (r *memoryRepository) GetUserInventory(userID int) ([]*entity.UserInventory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var inventories []*entity.UserInventory
	for _, inv := range r.inventories {
		if inv.UserID == userID && inv.Count > 0 {
			inventories = append(inventories, inv)
		}
	}
	return inventories, nil
}

func (r *memoryRepository) GetUserInventoryItem(userID, itemID int) (*entity.UserInventory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := fmt.Sprintf("%d:%d", userID, itemID)
	inventory, exists := r.inventories[key]
	if !exists {
		return nil, fmt.Errorf("inventory item not found for user %d, item %d", userID, itemID)
	}
	return inventory, nil
}

func (r *memoryRepository) AddToInventory(userID, itemID int, count int, source string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Verify item exists
	if _, exists := r.items[itemID]; !exists {
		return fmt.Errorf("item with id %d not found", itemID)
	}

	key := fmt.Sprintf("%d:%d", userID, itemID)
	
	if existing, exists := r.inventories[key]; exists {
		// Update existing inventory
		existing.Count += count
		existing.UpdatedAt = time.Now()
	} else {
		// Create new inventory entry
		r.invCounter++
		r.inventories[key] = &entity.UserInventory{
			ID:         r.invCounter,
			UserID:     userID,
			ItemID:     itemID,
			Count:      count,
			AcquiredAt: time.Now(),
			Source:     source,
			UpdatedAt:  time.Now(),
		}
	}

	return nil
}

func (r *memoryRepository) UpdateInventoryCount(userID, itemID int, count int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%d:%d", userID, itemID)
	inventory, exists := r.inventories[key]
	if !exists {
		return fmt.Errorf("inventory item not found for user %d, item %d", userID, itemID)
	}

	inventory.Count = count
	inventory.UpdatedAt = time.Now()
	return nil
}

func (r *memoryRepository) RemoveFromInventory(userID, itemID int, count int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%d:%d", userID, itemID)
	inventory, exists := r.inventories[key]
	if !exists {
		return fmt.Errorf("inventory item not found for user %d, item %d", userID, itemID)
	}

	if inventory.Count < count {
		return fmt.Errorf("insufficient item count: have %d, trying to remove %d", inventory.Count, count)
	}

	inventory.Count -= count
	inventory.UpdatedAt = time.Now()
	return nil
}

func (r *memoryRepository) AddMultipleToInventory(userID int, items []entity.RewardItem, source string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Verify all items exist first
	for _, item := range items {
		if _, exists := r.items[item.ItemID]; !exists {
			return fmt.Errorf("item with id %d not found", item.ItemID)
		}
	}

	// Add all items
	for _, item := range items {
		key := fmt.Sprintf("%d:%d", userID, item.ItemID)
		
		if existing, exists := r.inventories[key]; exists {
			existing.Count += item.Count
			existing.UpdatedAt = time.Now()
		} else {
			r.invCounter++
			r.inventories[key] = &entity.UserInventory{
				ID:         r.invCounter,
				UserID:     userID,
				ItemID:     item.ItemID,
				Count:      item.Count,
				AcquiredAt: time.Now(),
				Source:     source,
				UpdatedAt:  time.Now(),
			}
		}
	}

	return nil
}