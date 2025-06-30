package reward

import (
	"fmt"
	"time"

	"fxserver/modules/item"
	"fxserver/modules/item/entity"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Service interface {
	// Core reward granting
	GrantRewards(req GrantRewardRequest) (*GrantRewardResponse, error)
	BulkGrantRewards(req BulkGrantRewardRequest) (*BulkGrantRewardResponse, error)
	
	// Helper methods for other services
	GrantItemsToUser(userID int, items []entity.RewardItem, source, description string) error
	ValidateRewardItems(items []entity.RewardItem) error
}

type service struct {
	itemService item.Service
	logger      *zap.Logger
}

type ServiceParam struct {
	fx.In
	ItemService item.Service
	Logger      *zap.Logger
}

func NewService(p ServiceParam) Service {
	return &service{
		itemService: p.ItemService,
		logger:      p.Logger,
	}
}

func (s *service) GrantRewards(req GrantRewardRequest) (*GrantRewardResponse, error) {
	// Validate reward source
	if !IsValidRewardSource(req.Source) {
		return &GrantRewardResponse{
			UserID:  req.UserID,
			Items:   req.Items,
			Source:  req.Source,
			Success: false,
			Message: "Invalid reward source",
		}, fmt.Errorf("invalid reward source: %s", req.Source)
	}

	// Validate reward items
	if err := s.ValidateRewardItems(req.Items); err != nil {
		return &GrantRewardResponse{
			UserID:  req.UserID,
			Items:   req.Items,
			Source:  req.Source,
			Success: false,
			Message: err.Error(),
		}, err
	}

	// Grant items to user
	if err := s.GrantItemsToUser(req.UserID, req.Items, req.Source, req.Description); err != nil {
		s.logger.Error("Failed to grant rewards to user", 
			zap.Error(err),
			zap.Int("user_id", req.UserID),
			zap.String("source", req.Source))
		
		return &GrantRewardResponse{
			UserID:  req.UserID,
			Items:   req.Items,
			Source:  req.Source,
			Success: false,
			Message: err.Error(),
		}, err
	}

	s.logger.Info("Rewards granted successfully", 
		zap.Int("user_id", req.UserID),
		zap.String("source", req.Source),
		zap.Int("item_count", len(req.Items)),
		zap.String("description", req.Description))

	return &GrantRewardResponse{
		UserID:      req.UserID,
		Items:       req.Items,
		Source:      req.Source,
		Description: req.Description,
		GrantedAt:   time.Now().Format(time.RFC3339),
		Success:     true,
		Message:     "Rewards granted successfully",
	}, nil
}

func (s *service) BulkGrantRewards(req BulkGrantRewardRequest) (*BulkGrantRewardResponse, error) {
	// Validate reward source
	if !IsValidRewardSource(req.Source) {
		return nil, fmt.Errorf("invalid reward source: %s", req.Source)
	}

	// Validate reward items
	if err := s.ValidateRewardItems(req.Items); err != nil {
		return nil, err
	}

	results := make([]GrantRewardResponse, len(req.UserIDs))
	successCount := 0
	failureCount := 0

	// Grant rewards to each user
	for i, userID := range req.UserIDs {
		err := s.GrantItemsToUser(userID, req.Items, req.Source, req.Description)
		
		results[i] = GrantRewardResponse{
			UserID:      userID,
			Items:       req.Items,
			Source:      req.Source,
			Description: req.Description,
			GrantedAt:   time.Now().Format(time.RFC3339),
		}

		if err != nil {
			results[i].Success = false
			results[i].Message = err.Error()
			failureCount++
			
			s.logger.Warn("Failed to grant rewards to user in bulk operation", 
				zap.Error(err),
				zap.Int("user_id", userID),
				zap.String("source", req.Source))
		} else {
			results[i].Success = true
			results[i].Message = "Rewards granted successfully"
			successCount++
		}
	}

	s.logger.Info("Bulk reward grant completed", 
		zap.String("source", req.Source),
		zap.Int("total_users", len(req.UserIDs)),
		zap.Int("success_count", successCount),
		zap.Int("failure_count", failureCount),
		zap.String("description", req.Description))

	return &BulkGrantRewardResponse{
		TotalUsers:   len(req.UserIDs),
		SuccessCount: successCount,
		FailureCount: failureCount,
		Results:      results,
		Items:        req.Items,
		Source:       req.Source,
		Description:  req.Description,
	}, nil
}

func (s *service) GrantItemsToUser(userID int, items []entity.RewardItem, source, description string) error {
	// Validate inputs
	if userID <= 0 {
		return fmt.Errorf("invalid user ID: %d", userID)
	}

	if len(items) == 0 {
		return fmt.Errorf("no items to grant")
	}

	if source == "" {
		return fmt.Errorf("reward source is required")
	}

	// Validate all items exist before granting any
	if err := s.ValidateRewardItems(items); err != nil {
		return err
	}

	// Grant all items using item service
	if err := s.itemService.AddMultipleToInventory(userID, items, source); err != nil {
		return fmt.Errorf("failed to add items to inventory: %w", err)
	}

	return nil
}

func (s *service) ValidateRewardItems(items []entity.RewardItem) error {
	if len(items) == 0 {
		return fmt.Errorf("no reward items provided")
	}

	// Check for duplicate item IDs
	itemIDMap := make(map[int]bool)
	
	for i, item := range items {
		// Validate item fields
		if item.ItemID <= 0 {
			return fmt.Errorf("invalid item ID at index %d: %d", i, item.ItemID)
		}
		
		if item.Count <= 0 {
			return fmt.Errorf("invalid item count at index %d: %d", i, item.Count)
		}

		// Check for duplicates
		if itemIDMap[item.ItemID] {
			return fmt.Errorf("duplicate item ID found: %d", item.ItemID)
		}
		itemIDMap[item.ItemID] = true

		// Verify item exists in the system
		if _, err := s.itemService.GetItem(item.ItemID); err != nil {
			return fmt.Errorf("item with ID %d not found", item.ItemID)
		}
	}

	return nil
}