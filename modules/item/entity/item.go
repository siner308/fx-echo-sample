package entity

import "time"

type ItemType string

const (
	ItemTypeCurrency   ItemType = "currency"   // 게임 내 화폐 (골드, 다이아몬드)
	ItemTypeEquipment  ItemType = "equipment"  // 장비 아이템 (무기, 방어구)
	ItemTypeConsumable ItemType = "consumable" // 소모품 (포션, 버프 아이템)
	ItemTypeCard       ItemType = "card"       // 수집형 카드/캐릭터
	ItemTypeMaterial   ItemType = "material"   // 제작/강화 재료
	ItemTypeTicket     ItemType = "ticket"     // 이용권 (던전 입장권, 가챠 티켓)
)

type Item struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        ItemType `json:"type"`
	Value       int      `json:"value"`       // 기본값, 타입별 의미 다름
	Rarity      string   `json:"rarity"`      // common, rare, epic, legendary
	IconURL     string   `json:"icon_url"`    // 아이템 아이콘 이미지 URL
	IsActive    bool     `json:"is_active"`   // 활성화 여부
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserInventory struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	ItemID     int       `json:"item_id"`
	Count      int       `json:"count"`       // 보유 수량
	AcquiredAt time.Time `json:"acquired_at"` // 최초 획득 시간
	Source     string    `json:"source"`      // 획득 경로 (coupon, payment, reward, admin)
	UpdatedAt  time.Time `json:"updated_at"`  // 수량 변경 시간
}

type RewardItem struct {
	ItemID int `json:"item_id"`
	Count  int `json:"count"`
}

// Item Response DTOs
type ItemResponse struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        ItemType `json:"type"`
	Value       int      `json:"value"`
	Rarity      string   `json:"rarity"`
	IconURL     string   `json:"icon_url"`
}

type InventoryResponse struct {
	ID         int          `json:"id"`
	Item       ItemResponse `json:"item"`
	Count      int          `json:"count"`
	AcquiredAt time.Time    `json:"acquired_at"`
	Source     string       `json:"source"`
	UpdatedAt  time.Time    `json:"updated_at"`
}

type UserInventoryResponse struct {
	UserID int                 `json:"user_id"`
	Items  []InventoryResponse `json:"items"`
	Total  int                 `json:"total"`
}

// Helper methods
func (i *Item) ToResponse() ItemResponse {
	return ItemResponse{
		ID:          i.ID,
		Name:        i.Name,
		Description: i.Description,
		Type:        i.Type,
		Value:       i.Value,
		Rarity:      i.Rarity,
		IconURL:     i.IconURL,
	}
}

func (ui *UserInventory) ToResponse(item *Item) InventoryResponse {
	return InventoryResponse{
		ID:         ui.ID,
		Item:       item.ToResponse(),
		Count:      ui.Count,
		AcquiredAt: ui.AcquiredAt,
		Source:     ui.Source,
		UpdatedAt:  ui.UpdatedAt,
	}
}

// GetTypeDescription returns description of what Value field means for each ItemType
func (t ItemType) GetValueDescription() string {
	switch t {
	case ItemTypeCurrency:
		return "지급할 화폐 수량"
	case ItemTypeEquipment:
		return "강화 레벨 또는 등급 (기본값: 1)"
	case ItemTypeConsumable:
		return "지급할 개수"
	case ItemTypeCard:
		return "카드 레벨 또는 등급 (기본값: 1)"
	case ItemTypeMaterial:
		return "지급할 재료 개수"
	case ItemTypeTicket:
		return "지급할 티켓 개수"
	default:
		return "알 수 없는 타입"
	}
}

// IsValidType validates if the item type is valid
func IsValidItemType(itemType string) bool {
	switch ItemType(itemType) {
	case ItemTypeCurrency, ItemTypeEquipment, ItemTypeConsumable, 
		 ItemTypeCard, ItemTypeMaterial, ItemTypeTicket:
		return true
	default:
		return false
	}
}

// IsValidRarity validates if the rarity is valid
func IsValidRarity(rarity string) bool {
	switch rarity {
	case "common", "rare", "epic", "legendary":
		return true
	default:
		return false
	}
}