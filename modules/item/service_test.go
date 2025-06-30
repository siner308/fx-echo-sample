package item

import (
	"errors"
	"testing"

	"fxserver/modules/item/entity"
	"fxserver/modules/item/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Mock repository for testing
type MockItemRepository struct {
	mock.Mock
}

func (m *MockItemRepository) Create(item *entity.Item) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockItemRepository) GetByID(id int) (*entity.Item, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Item), args.Error(1)
}

func (m *MockItemRepository) Update(item *entity.Item) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockItemRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockItemRepository) List() ([]*entity.Item, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Item), args.Error(1)
}

func (m *MockItemRepository) GetByType(itemType entity.ItemType) ([]*entity.Item, error) {
	args := m.Called(itemType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Item), args.Error(1)
}

func (m *MockItemRepository) AddToInventory(userID, itemID, count int, source string) error {
	args := m.Called(userID, itemID, count, source)
	return args.Error(0)
}

func (m *MockItemRepository) GetUserInventory(userID int) (*entity.UserInventory, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserInventory), args.Error(1)
}

func setupItemService(mockRepo repository.ItemRepository) Service {
	logger := zap.NewNop()
	return &service{
		repository: mockRepo,
		logger:     logger,
	}
}

func TestCreateItem(t *testing.T) {
	tests := []struct {
		name        string
		request     CreateItemRequest
		setupMock   func(*MockItemRepository)
		wantErr     bool
		wantErrType error
	}{
		{
			name: "successful item creation",
			request: CreateItemRequest{
				Name:        "Test Sword",
				Description: "A powerful sword",
				Type:        string(entity.ItemTypeEquipment),
				Rarity:      string(entity.RarityRare),
				IsActive:    true,
			},
			setupMock: func(m *MockItemRepository) {
				m.On("Create", mock.AnythingOfType("*entity.Item")).Return(nil).Run(func(args mock.Arguments) {
					item := args.Get(0).(*entity.Item)
					item.ID = 1
				})
			},
			wantErr: false,
		},
		{
			name: "invalid item type",
			request: CreateItemRequest{
				Name:        "Test Item",
				Description: "Test description",
				Type:        "invalid_type",
				Rarity:      string(entity.RarityCommon),
				IsActive:    true,
			},
			setupMock: func(m *MockItemRepository) {
				// No mock setup needed as validation should fail before repository call
			},
			wantErr:     true,
			wantErrType: ErrInvalidItemType,
		},
		{
			name: "invalid rarity",
			request: CreateItemRequest{
				Name:        "Test Item",
				Description: "Test description",
				Type:        string(entity.ItemTypeCurrency),
				Rarity:      "invalid_rarity",
				IsActive:    true,
			},
			setupMock: func(m *MockItemRepository) {
				// No mock setup needed as validation should fail before repository call
			},
			wantErr:     true,
			wantErrType: ErrInvalidRarity,
		},
		{
			name: "repository error",
			request: CreateItemRequest{
				Name:        "Test Item",
				Description: "Test description",
				Type:        string(entity.ItemTypeCurrency),
				Rarity:      string(entity.RarityCommon),
				IsActive:    true,
			},
			setupMock: func(m *MockItemRepository) {
				m.On("Create", mock.AnythingOfType("*entity.Item")).Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockItemRepository)
			tt.setupMock(mockRepo)

			service := setupItemService(mockRepo)

			item, err := service.CreateItem(tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrType != nil {
					assert.ErrorIs(t, err, tt.wantErrType)
				}
				assert.Nil(t, item)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, item)
				assert.Equal(t, tt.request.Name, item.Name)
				assert.Equal(t, tt.request.Description, item.Description)
				assert.Equal(t, entity.ItemType(tt.request.Type), item.Type)
				assert.Equal(t, entity.Rarity(tt.request.Rarity), item.Rarity)
				assert.Equal(t, tt.request.IsActive, item.IsActive)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetItem(t *testing.T) {
	testItem := &entity.Item{
		ID:          1,
		Name:        "Test Sword",
		Description: "A powerful sword",
		Type:        entity.ItemTypeEquipment,
		Rarity:      entity.RarityRare,
		IsActive:    true,
	}

	tests := []struct {
		name        string
		itemID      int
		setupMock   func(*MockItemRepository)
		wantErr     bool
		wantErrType error
	}{
		{
			name:   "successful item retrieval",
			itemID: 1,
			setupMock: func(m *MockItemRepository) {
				m.On("GetByID", 1).Return(testItem, nil)
			},
			wantErr: false,
		},
		{
			name:   "item not found",
			itemID: 999,
			setupMock: func(m *MockItemRepository) {
				m.On("GetByID", 999).Return(nil, repository.ErrItemNotFound)
			},
			wantErr:     true,
			wantErrType: ErrItemNotFound,
		},
		{
			name:   "repository error",
			itemID: 1,
			setupMock: func(m *MockItemRepository) {
				m.On("GetByID", 1).Return(nil, errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockItemRepository)
			tt.setupMock(mockRepo)

			service := setupItemService(mockRepo)

			item, err := service.GetItem(tt.itemID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrType != nil {
					assert.ErrorIs(t, err, tt.wantErrType)
				}
				assert.Nil(t, item)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, item)
				assert.Equal(t, testItem.ID, item.ID)
				assert.Equal(t, testItem.Name, item.Name)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetItems(t *testing.T) {
	testItems := []*entity.Item{
		{ID: 1, Name: "Sword", Type: entity.ItemTypeEquipment, IsActive: true},
		{ID: 2, Name: "Gold", Type: entity.ItemTypeCurrency, IsActive: true},
	}

	tests := []struct {
		name      string
		setupMock func(*MockItemRepository)
		wantErr   bool
		wantCount int
	}{
		{
			name: "successful items retrieval",
			setupMock: func(m *MockItemRepository) {
				m.On("List").Return(testItems, nil)
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name: "empty items list",
			setupMock: func(m *MockItemRepository) {
				m.On("List").Return([]*entity.Item{}, nil)
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "repository error",
			setupMock: func(m *MockItemRepository) {
				m.On("List").Return(nil, errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockItemRepository)
			tt.setupMock(mockRepo)

			service := setupItemService(mockRepo)

			items, err := service.GetItems()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, items)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, items)
				assert.Len(t, items, tt.wantCount)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetItemsByType(t *testing.T) {
	equipmentItems := []*entity.Item{
		{ID: 1, Name: "Sword", Type: entity.ItemTypeEquipment, IsActive: true},
		{ID: 2, Name: "Shield", Type: entity.ItemTypeEquipment, IsActive: true},
	}

	tests := []struct {
		name      string
		itemType  entity.ItemType
		setupMock func(*MockItemRepository)
		wantErr   bool
		wantCount int
	}{
		{
			name:     "successful items by type retrieval",
			itemType: entity.ItemTypeEquipment,
			setupMock: func(m *MockItemRepository) {
				m.On("GetByType", entity.ItemTypeEquipment).Return(equipmentItems, nil)
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:     "no items found for type",
			itemType: entity.ItemTypeCard,
			setupMock: func(m *MockItemRepository) {
				m.On("GetByType", entity.ItemTypeCard).Return([]*entity.Item{}, nil)
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:     "repository error",
			itemType: entity.ItemTypeEquipment,
			setupMock: func(m *MockItemRepository) {
				m.On("GetByType", entity.ItemTypeEquipment).Return(nil, errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockItemRepository)
			tt.setupMock(mockRepo)

			service := setupItemService(mockRepo)

			items, err := service.GetItemsByType(tt.itemType)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, items)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, items)
				assert.Len(t, items, tt.wantCount)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAddToInventory(t *testing.T) {
	tests := []struct {
		name      string
		userID    int
		itemID    int
		count     int
		source    string
		setupMock func(*MockItemRepository)
		wantErr   bool
	}{
		{
			name:   "successful inventory addition",
			userID: 1,
			itemID: 1,
			count:  5,
			source: "purchase",
			setupMock: func(m *MockItemRepository) {
				m.On("AddToInventory", 1, 1, 5, "purchase").Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "repository error",
			userID: 1,
			itemID: 1,
			count:  5,
			source: "purchase",
			setupMock: func(m *MockItemRepository) {
				m.On("AddToInventory", 1, 1, 5, "purchase").Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockItemRepository)
			tt.setupMock(mockRepo)

			service := setupItemService(mockRepo)

			err := service.AddToInventory(tt.userID, tt.itemID, tt.count, tt.source)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAddMultipleToInventory(t *testing.T) {
	rewardItems := []entity.RewardItem{
		{ItemID: 1, Count: 5},
		{ItemID: 2, Count: 10},
	}

	tests := []struct {
		name        string
		userID      int
		rewardItems []entity.RewardItem
		source      string
		setupMock   func(*MockItemRepository)
		wantErr     bool
	}{
		{
			name:        "successful multiple inventory addition",
			userID:      1,
			rewardItems: rewardItems,
			source:      "quest_reward",
			setupMock: func(m *MockItemRepository) {
				m.On("AddToInventory", 1, 1, 5, "quest_reward").Return(nil)
				m.On("AddToInventory", 1, 2, 10, "quest_reward").Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "first item fails",
			userID:      1,
			rewardItems: rewardItems,
			source:      "quest_reward",
			setupMock: func(m *MockItemRepository) {
				m.On("AddToInventory", 1, 1, 5, "quest_reward").Return(errors.New("database error"))
				// Second call should not happen due to early return
			},
			wantErr: true,
		},
		{
			name:        "empty reward items",
			userID:      1,
			rewardItems: []entity.RewardItem{},
			source:      "quest_reward",
			setupMock: func(m *MockItemRepository) {
				// No repository calls expected
			},
			wantErr: false, // Should handle empty list gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockItemRepository)
			tt.setupMock(mockRepo)

			service := setupItemService(mockRepo)

			err := service.AddMultipleToInventory(tt.userID, tt.rewardItems, tt.source)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetUserInventory(t *testing.T) {
	testInventory := &entity.UserInventory{
		UserID: 1,
		Items: []entity.InventoryItem{
			{ItemID: 1, Count: 5, LastUpdated: "2023-01-01"},
			{ItemID: 2, Count: 10, LastUpdated: "2023-01-02"},
		},
		TotalItems: 2,
	}

	tests := []struct {
		name      string
		userID    int
		setupMock func(*MockItemRepository)
		wantErr   bool
	}{
		{
			name:   "successful inventory retrieval",
			userID: 1,
			setupMock: func(m *MockItemRepository) {
				m.On("GetUserInventory", 1).Return(testInventory, nil)
			},
			wantErr: false,
		},
		{
			name:   "user has no inventory",
			userID: 2,
			setupMock: func(m *MockItemRepository) {
				m.On("GetUserInventory", 2).Return(&entity.UserInventory{
					UserID:     2,
					Items:      []entity.InventoryItem{},
					TotalItems: 0,
				}, nil)
			},
			wantErr: false,
		},
		{
			name:   "repository error",
			userID: 1,
			setupMock: func(m *MockItemRepository) {
				m.On("GetUserInventory", 1).Return(nil, errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockItemRepository)
			tt.setupMock(mockRepo)

			service := setupItemService(mockRepo)

			inventory, err := service.GetUserInventory(tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, inventory)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, inventory)
				assert.Equal(t, tt.userID, inventory.UserID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateItem(t *testing.T) {
	existingItem := &entity.Item{
		ID:          1,
		Name:        "Old Sword",
		Description: "An old sword",
		Type:        entity.ItemTypeEquipment,
		Rarity:      entity.RarityCommon,
		IsActive:    true,
	}

	tests := []struct {
		name        string
		itemID      int
		request     UpdateItemRequest
		setupMock   func(*MockItemRepository)
		wantErr     bool
		wantErrType error
	}{
		{
			name:   "successful item update",
			itemID: 1,
			request: UpdateItemRequest{
				Name:        "Updated Sword",
				Description: "An updated sword",
				Type:        string(entity.ItemTypeEquipment),
				Rarity:      string(entity.RarityRare),
				IsActive:    &[]bool{false}[0], // Helper to get pointer
			},
			setupMock: func(m *MockItemRepository) {
				m.On("GetByID", 1).Return(existingItem, nil)
				m.On("Update", mock.AnythingOfType("*entity.Item")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "item not found",
			itemID: 999,
			request: UpdateItemRequest{
				Name: "Updated Name",
			},
			setupMock: func(m *MockItemRepository) {
				m.On("GetByID", 999).Return(nil, repository.ErrItemNotFound)
			},
			wantErr:     true,
			wantErrType: ErrItemNotFound,
		},
		{
			name:   "invalid item type",
			itemID: 1,
			request: UpdateItemRequest{
				Type: "invalid_type",
			},
			setupMock: func(m *MockItemRepository) {
				m.On("GetByID", 1).Return(existingItem, nil)
			},
			wantErr:     true,
			wantErrType: ErrInvalidItemType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockItemRepository)
			tt.setupMock(mockRepo)

			service := setupItemService(mockRepo)

			item, err := service.UpdateItem(tt.itemID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrType != nil {
					assert.ErrorIs(t, err, tt.wantErrType)
				}
				assert.Nil(t, item)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, item)
				assert.Equal(t, tt.request.Name, item.Name)
				assert.Equal(t, tt.request.Description, item.Description)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDeleteItem(t *testing.T) {
	tests := []struct {
		name        string
		itemID      int
		setupMock   func(*MockItemRepository)
		wantErr     bool
		wantErrType error
	}{
		{
			name:   "successful item deletion",
			itemID: 1,
			setupMock: func(m *MockItemRepository) {
				m.On("Delete", 1).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "item not found",
			itemID: 999,
			setupMock: func(m *MockItemRepository) {
				m.On("Delete", 999).Return(repository.ErrItemNotFound)
			},
			wantErr:     true,
			wantErrType: ErrItemNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockItemRepository)
			tt.setupMock(mockRepo)

			service := setupItemService(mockRepo)

			err := service.DeleteItem(tt.itemID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrType != nil {
					assert.ErrorIs(t, err, tt.wantErrType)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetItemTypes(t *testing.T) {
	service := setupItemService(new(MockItemRepository))

	itemTypes := service.GetItemTypes()

	assert.NotNil(t, itemTypes)
	assert.Len(t, itemTypes, 6) // 6 item types defined in entity

	// Check if all expected types are present
	typeNames := make(map[string]bool)
	for _, itemType := range itemTypes {
		typeNames[itemType.Type] = true
	}

	expectedTypes := []string{
		string(entity.ItemTypeCurrency),
		string(entity.ItemTypeEquipment),
		string(entity.ItemTypeConsumable),
		string(entity.ItemTypeCard),
		string(entity.ItemTypeMaterial),
		string(entity.ItemTypeTicket),
	}

	for _, expectedType := range expectedTypes {
		assert.True(t, typeNames[expectedType], "Expected type %s not found", expectedType)
	}
}