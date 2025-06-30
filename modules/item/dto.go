package item

import "fxserver/modules/item/entity"

// Item management DTOs (Admin only)
type CreateItemRequest struct {
	Name        string           `json:"name" validate:"required,min=2,max=100"`
	Description string           `json:"description" validate:"required,min=5,max=500"`
	Type        entity.ItemType  `json:"type" validate:"required"`
	Value       int              `json:"value" validate:"required,gt=0"`
	Rarity      string           `json:"rarity" validate:"required,oneof=common rare epic legendary"`
	IconURL     string           `json:"icon_url" validate:"omitempty,url"`
}

type UpdateItemRequest struct {
	Name        string           `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Description string           `json:"description,omitempty" validate:"omitempty,min=5,max=500"`
	Type        entity.ItemType  `json:"type,omitempty"`
	Value       int              `json:"value,omitempty" validate:"omitempty,gt=0"`
	Rarity      string           `json:"rarity,omitempty" validate:"omitempty,oneof=common rare epic legendary"`
	IconURL     string           `json:"icon_url,omitempty" validate:"omitempty,url"`
}

// Inventory DTOs
type GetInventoryRequest struct {
	UserID int `json:"user_id" validate:"required,gt=0"`
}

// Response DTOs
type ListItemsResponse struct {
	Items []entity.ItemResponse `json:"items"`
	Total int                   `json:"total"`
}

type ItemTypesResponse struct {
	Types []ItemTypeInfo `json:"types"`
}

type ItemTypeInfo struct {
	Type        entity.ItemType `json:"type"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
}

// Helper function to get all item types with descriptions
func GetItemTypes() []ItemTypeInfo {
	return []ItemTypeInfo{
		{
			Type:        entity.ItemTypeCurrency,
			Name:        "화폐",
			Description: entity.ItemTypeCurrency.GetValueDescription(),
		},
		{
			Type:        entity.ItemTypeEquipment,
			Name:        "장비",
			Description: entity.ItemTypeEquipment.GetValueDescription(),
		},
		{
			Type:        entity.ItemTypeConsumable,
			Name:        "소모품",
			Description: entity.ItemTypeConsumable.GetValueDescription(),
		},
		{
			Type:        entity.ItemTypeCard,
			Name:        "카드",
			Description: entity.ItemTypeCard.GetValueDescription(),
		},
		{
			Type:        entity.ItemTypeMaterial,
			Name:        "재료",
			Description: entity.ItemTypeMaterial.GetValueDescription(),
		},
		{
			Type:        entity.ItemTypeTicket,
			Name:        "티켓",
			Description: entity.ItemTypeTicket.GetValueDescription(),
		},
	}
}