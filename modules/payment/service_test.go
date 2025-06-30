package payment

import (
	"errors"
	"testing"
	"time"

	itemEntity "fxserver/modules/item/entity"
	"fxserver/modules/payment/entity"
	"fxserver/modules/payment/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Mock repository for testing
type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) Create(payment *entity.Payment) error {
	args := m.Called(payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetByID(id int) (*entity.Payment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Payment), args.Error(1)
}

func (m *MockPaymentRepository) GetByExternalID(externalID string) (*entity.Payment, error) {
	args := m.Called(externalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Payment), args.Error(1)
}

func (m *MockPaymentRepository) Update(payment *entity.Payment) error {
	args := m.Called(payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) List(userID int, status entity.PaymentStatus) ([]*entity.Payment, error) {
	args := m.Called(userID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Payment), args.Error(1)
}

func (m *MockPaymentRepository) GetUserPayments(userID int) ([]*entity.Payment, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Payment), args.Error(1)
}

func setupPaymentService(mockRepo repository.Repository) Service {
	logger := zap.NewNop()
	return &service{
		repository: mockRepo,
		logger:     logger,
	}
}

func TestProcessPayment(t *testing.T) {
	tests := []struct {
		name        string
		request     CreatePaymentRequest
		setupMock   func(*MockPaymentRepository)
		wantErr     bool
		wantErrType error
	}{
		{
			name: "successful payment processing",
			request: CreatePaymentRequest{
				UserID:   1,
				Amount:   99.99,
				Currency: "USD",
				Method:   entity.PaymentMethodCard,
				ExternalID: "ext_12345",
				RewardItems: []itemEntity.RewardItem{
					{ItemID: 1, Count: 5},
					{ItemID: 2, Count: 10},
				},
			},
			setupMock: func(m *MockPaymentRepository) {
				m.On("GetPaymentByExternalID", "ext_12345").Return(nil, repository.ErrPaymentNotFound)
				m.On("Create", mock.AnythingOfType("*entity.Payment")).Return(nil).Run(func(args mock.Arguments) {
					payment := args.Get(0).(*entity.Payment)
					payment.ID = 1
				})
			},
			wantErr: false,
		},
		{
			name: "invalid payment method",
			request: CreatePaymentRequest{
				UserID:     1,
				Amount:     99.99,
				Currency:   "USD",
				Method:     "invalid_method",
				ExternalID: "ext_12345",
				RewardItems: []itemEntity.RewardItem{
					{ItemID: 1, Count: 5},
				},
			},
			setupMock: func(m *MockPaymentRepository) {
				// No repository calls expected due to early validation failure
			},
			wantErr:     true,
			wantErrType: ErrInvalidPaymentMethod,
		},
		{
			name: "invalid amount",
			request: CreatePaymentRequest{
				UserID:     1,
				Amount:     -10.0,
				Currency:   "USD",
				Method:     entity.PaymentMethodCard,
				ExternalID: "ext_12345",
				RewardItems: []itemEntity.RewardItem{
					{ItemID: 1, Count: 5},
				},
			},
			setupMock: func(m *MockPaymentRepository) {
				// No repository calls expected due to early validation failure
			},
			wantErr:     true,
			wantErrType: ErrInvalidAmount,
		},
		{
			name: "payment already exists",
			request: CreatePaymentRequest{
				UserID:     1,
				Amount:     99.99,
				Currency:   "USD",
				Method:     entity.PaymentMethodCard,
				ExternalID: "existing_ext_id",
				RewardItems: []itemEntity.RewardItem{
					{ItemID: 1, Count: 5},
				},
			},
			setupMock: func(m *MockPaymentRepository) {
				existingPayment := &entity.Payment{ID: 1, ExternalID: "existing_ext_id"}
				m.On("GetPaymentByExternalID", "existing_ext_id").Return(existingPayment, nil)
			},
			wantErr:     true,
			wantErrType: ErrPaymentAlreadyExists,
		},
		{
			name: "empty reward items",
			request: CreatePaymentRequest{
				UserID:      1,
				Amount:      99.99,
				Currency:    "USD",
				Method:      entity.PaymentMethodCard,
				ExternalID:  "ext_12345",
				RewardItems: []itemEntity.RewardItem{},
			},
			setupMock: func(m *MockPaymentRepository) {
				// No repository calls expected due to validation failure
			},
			wantErr: true,
		},
		{
			name: "invalid reward item",
			request: CreatePaymentRequest{
				UserID:     1,
				Amount:     99.99,
				Currency:   "USD",
				Method:     entity.PaymentMethodCard,
				ExternalID: "ext_12345",
				RewardItems: []itemEntity.RewardItem{
					{ItemID: 0, Count: 5}, // Invalid ItemID
				},
			},
			setupMock: func(m *MockPaymentRepository) {
				m.On("GetPaymentByExternalID", "ext_12345").Return(nil, repository.ErrPaymentNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPaymentRepository)
			tt.setupMock(mockRepo)

			service := setupPaymentService(mockRepo)

			response, err := service.ProcessPayment(tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrType != nil {
					assert.ErrorIs(t, err, tt.wantErrType)
				}
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.NotZero(t, response.PaymentID)
				assert.Equal(t, entity.PaymentStatusPending, response.Status)
				assert.NotEmpty(t, response.Message)
				assert.Equal(t, tt.request.RewardItems, response.RewardItems)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUpdatePaymentStatus(t *testing.T) {
	existingPayment := &entity.Payment{
		ID:         1,
		UserID:     1,
		Amount:     99.99,
		Currency:   "USD",
		Status:     entity.PaymentStatusPending,
		Method:     entity.PaymentMethodCard,
		ExternalID: "ext_12345",
		RewardItems: []itemEntity.RewardItem{
			{ItemID: 1, Count: 5},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name        string
		paymentID   int
		request     UpdatePaymentStatusRequest
		setupMock   func(*MockPaymentRepository)
		wantErr     bool
		wantErrType error
	}{
		{
			name:      "successful status update to completed",
			paymentID: 1,
			request: UpdatePaymentStatusRequest{
				Status: entity.PaymentStatusCompleted,
			},
			setupMock: func(m *MockPaymentRepository) {
				m.On("GetByID", 1).Return(existingPayment, nil)
				m.On("Update", mock.AnythingOfType("*entity.Payment")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "successful status update to failed",
			paymentID: 1,
			request: UpdatePaymentStatusRequest{
				Status:        entity.PaymentStatusFailed,
				FailureReason: "Insufficient funds",
			},
			setupMock: func(m *MockPaymentRepository) {
				m.On("GetByID", 1).Return(existingPayment, nil)
				m.On("Update", mock.AnythingOfType("*entity.Payment")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "payment not found",
			paymentID: 999,
			request: UpdatePaymentStatusRequest{
				Status: entity.PaymentStatusCompleted,
			},
			setupMock: func(m *MockPaymentRepository) {
				m.On("GetByID", 999).Return(nil, repository.ErrPaymentNotFound)
			},
			wantErr:     true,
			wantErrType: ErrPaymentNotFound,
		},
		{
			name:      "invalid payment status",
			paymentID: 1,
			request: UpdatePaymentStatusRequest{
				Status: "invalid_status",
			},
			setupMock: func(m *MockPaymentRepository) {
				m.On("GetByID", 1).Return(existingPayment, nil)
			},
			wantErr:     true,
			wantErrType: ErrInvalidPaymentStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPaymentRepository)
			tt.setupMock(mockRepo)

			service := setupPaymentService(mockRepo)

			payment, err := service.UpdatePaymentStatus(tt.paymentID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrType != nil {
					assert.ErrorIs(t, err, tt.wantErrType)
				}
				assert.Nil(t, payment)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, payment)
				assert.Equal(t, tt.request.Status, payment.Status)
				if tt.request.Status == entity.PaymentStatusCompleted {
					assert.NotNil(t, payment.ProcessedAt)
				}
				if tt.request.FailureReason != "" {
					assert.Equal(t, tt.request.FailureReason, payment.FailureReason)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetPayment(t *testing.T) {
	testPayment := &entity.Payment{
		ID:         1,
		UserID:     1,
		Amount:     99.99,
		Currency:   "USD",
		Status:     entity.PaymentStatusCompleted,
		Method:     entity.PaymentMethodCard,
		ExternalID: "ext_12345",
	}

	tests := []struct {
		name        string
		paymentID   int
		setupMock   func(*MockPaymentRepository)
		wantErr     bool
		wantErrType error
	}{
		{
			name:      "successful payment retrieval",
			paymentID: 1,
			setupMock: func(m *MockPaymentRepository) {
				m.On("GetByID", 1).Return(testPayment, nil)
			},
			wantErr: false,
		},
		{
			name:      "payment not found",
			paymentID: 999,
			setupMock: func(m *MockPaymentRepository) {
				m.On("GetByID", 999).Return(nil, repository.ErrPaymentNotFound)
			},
			wantErr:     true,
			wantErrType: ErrPaymentNotFound,
		},
		{
			name:      "repository error",
			paymentID: 1,
			setupMock: func(m *MockPaymentRepository) {
				m.On("GetByID", 1).Return(nil, errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPaymentRepository)
			tt.setupMock(mockRepo)

			service := setupPaymentService(mockRepo)

			payment, err := service.GetPayment(tt.paymentID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrType != nil {
					assert.ErrorIs(t, err, tt.wantErrType)
				}
				assert.Nil(t, payment)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, payment)
				assert.Equal(t, testPayment.ID, payment.ID)
				assert.Equal(t, testPayment.UserID, payment.UserID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetPayments(t *testing.T) {
	testPayments := []*entity.Payment{
		{ID: 1, UserID: 1, Status: entity.PaymentStatusCompleted},
		{ID: 2, UserID: 1, Status: entity.PaymentStatusPending},
	}

	tests := []struct {
		name      string
		query     GetPaymentsQuery
		setupMock func(*MockPaymentRepository)
		wantErr   bool
		wantCount int
	}{
		{
			name: "get all payments for user",
			query: GetPaymentsQuery{
				UserID: 1,
			},
			setupMock: func(m *MockPaymentRepository) {
				m.On("List", 1, entity.PaymentStatus("")).Return(testPayments, nil)
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name: "get payments by status",
			query: GetPaymentsQuery{
				UserID: 1,
				Status: entity.PaymentStatusCompleted,
			},
			setupMock: func(m *MockPaymentRepository) {
				completedPayments := []*entity.Payment{testPayments[0]}
				m.On("List", 1, entity.PaymentStatusCompleted).Return(completedPayments, nil)
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name: "repository error",
			query: GetPaymentsQuery{
				UserID: 1,
			},
			setupMock: func(m *MockPaymentRepository) {
				m.On("List", 1, entity.PaymentStatus("")).Return(nil, errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPaymentRepository)
			tt.setupMock(mockRepo)

			service := setupPaymentService(mockRepo)

			payments, err := service.GetPayments(tt.query)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, payments)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, payments)
				assert.Len(t, payments, tt.wantCount)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestRefundPayment(t *testing.T) {
	completedPayment := &entity.Payment{
		ID:         1,
		UserID:     1,
		Amount:     99.99,
		Currency:   "USD",
		Status:     entity.PaymentStatusCompleted,
		Method:     entity.PaymentMethodCard,
		ExternalID: "ext_12345",
		ProcessedAt: &[]time.Time{time.Now()}[0],
	}

	pendingPayment := &entity.Payment{
		ID:         2,
		UserID:     1,
		Amount:     99.99,
		Currency:   "USD",
		Status:     entity.PaymentStatusPending,
		Method:     entity.PaymentMethodCard,
		ExternalID: "ext_67890",
	}

	tests := []struct {
		name        string
		paymentID   int
		request     RefundPaymentRequest
		setupMock   func(*MockPaymentRepository)
		wantErr     bool
		wantErrType error
	}{
		{
			name:      "successful refund",
			paymentID: 1,
			request: RefundPaymentRequest{
				Reason: "Customer requested refund",
			},
			setupMock: func(m *MockPaymentRepository) {
				m.On("GetByID", 1).Return(completedPayment, nil)
				m.On("Update", mock.AnythingOfType("*entity.Payment")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "payment not found",
			paymentID: 999,
			request: RefundPaymentRequest{
				Reason: "Customer requested refund",
			},
			setupMock: func(m *MockPaymentRepository) {
				m.On("GetByID", 999).Return(nil, repository.ErrPaymentNotFound)
			},
			wantErr:     true,
			wantErrType: ErrPaymentNotFound,
		},
		{
			name:      "payment not completed",
			paymentID: 2,
			request: RefundPaymentRequest{
				Reason: "Customer requested refund",
			},
			setupMock: func(m *MockPaymentRepository) {
				m.On("GetByID", 2).Return(pendingPayment, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPaymentRepository)
			tt.setupMock(mockRepo)

			service := setupPaymentService(mockRepo)

			payment, err := service.RefundPayment(tt.paymentID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrType != nil {
					assert.ErrorIs(t, err, tt.wantErrType)
				}
				assert.Nil(t, payment)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, payment)
				assert.Equal(t, entity.PaymentStatusRefunded, payment.Status)
				assert.NotNil(t, payment.RefundedAt)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetUserPaymentHistory(t *testing.T) {
	testPayments := []*entity.Payment{
		{ID: 1, UserID: 1, Amount: 99.99, Status: entity.PaymentStatusCompleted},
		{ID: 2, UserID: 1, Amount: 49.99, Status: entity.PaymentStatusCompleted},
		{ID: 3, UserID: 1, Amount: 19.99, Status: entity.PaymentStatusFailed},
	}

	tests := []struct {
		name      string
		userID    int
		setupMock func(*MockPaymentRepository)
		wantErr   bool
		wantCount int
	}{
		{
			name:   "successful payment history retrieval",
			userID: 1,
			setupMock: func(m *MockPaymentRepository) {
				m.On("GetUserPayments", 1).Return(testPayments, nil)
			},
			wantErr:   false,
			wantCount: 3,
		},
		{
			name:   "user with no payments",
			userID: 2,
			setupMock: func(m *MockPaymentRepository) {
				m.On("GetUserPayments", 2).Return([]*entity.Payment{}, nil)
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:   "repository error",
			userID: 1,
			setupMock: func(m *MockPaymentRepository) {
				m.On("GetUserPayments", 1).Return(nil, errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPaymentRepository)
			tt.setupMock(mockRepo)

			service := setupPaymentService(mockRepo)

			history, err := service.GetUserPaymentHistory(tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, history)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, history)
				assert.Len(t, history.Payments, tt.wantCount)
				assert.Equal(t, tt.wantCount, history.Total)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetPaymentMethods(t *testing.T) {
	service := setupPaymentService(new(MockPaymentRepository))

	methods := service.GetPaymentMethods()

	assert.NotNil(t, methods)
	assert.Len(t, methods.Methods, 5) // 5 payment methods defined

	// Check if all expected methods are present
	methodTypes := make(map[entity.PaymentMethod]bool)
	for _, method := range methods.Methods {
		methodTypes[method.Method] = true
	}

	expectedMethods := []entity.PaymentMethod{
		entity.PaymentMethodCard,
		entity.PaymentMethodBank,
		entity.PaymentMethodPaypal,
		entity.PaymentMethodApple,
		entity.PaymentMethodGoogle,
	}

	for _, expectedMethod := range expectedMethods {
		assert.True(t, methodTypes[expectedMethod], "Expected method %s not found", expectedMethod)
	}
}

func TestGetPaymentStatuses(t *testing.T) {
	service := setupPaymentService(new(MockPaymentRepository))

	statuses := service.GetPaymentStatuses()

	assert.NotNil(t, statuses)
	assert.Len(t, statuses.Statuses, 6) // 6 payment statuses defined

	// Check if all expected statuses are present
	statusTypes := make(map[entity.PaymentStatus]bool)
	for _, status := range statuses.Statuses {
		statusTypes[status.Status] = true
	}

	expectedStatuses := []entity.PaymentStatus{
		entity.PaymentStatusPending,
		entity.PaymentStatusProcessing,
		entity.PaymentStatusCompleted,
		entity.PaymentStatusFailed,
		entity.PaymentStatusCancelled,
		entity.PaymentStatusRefunded,
	}

	for _, expectedStatus := range expectedStatuses {
		assert.True(t, statusTypes[expectedStatus], "Expected status %s not found", expectedStatus)
	}
}

// Integration test for payment flow
func TestPaymentFlow(t *testing.T) {
	mockRepo := new(MockPaymentRepository)

	// Setup mocks for complete payment flow
	mockRepo.On("GetPaymentByExternalID", "flow_test_ext_id").Return(nil, repository.ErrPaymentNotFound)
	mockRepo.On("Create", mock.AnythingOfType("*entity.Payment")).Return(nil).Run(func(args mock.Arguments) {
		payment := args.Get(0).(*entity.Payment)
		payment.ID = 1
	})
	mockRepo.On("GetByID", 1).Return(&entity.Payment{
		ID:         1,
		UserID:     1,
		Amount:     99.99,
		Currency:   "USD",
		Status:     entity.PaymentStatusPending,
		Method:     entity.PaymentMethodCard,
		ExternalID: "flow_test_ext_id",
		RewardItems: []itemEntity.RewardItem{
			{ItemID: 1, Count: 5},
		},
	}, nil)
	mockRepo.On("Update", mock.AnythingOfType("*entity.Payment")).Return(nil)

	service := setupPaymentService(mockRepo)

	// Step 1: Process payment
	processRequest := CreatePaymentRequest{
		UserID:     1,
		Amount:     99.99,
		Currency:   "USD",
		Method:     entity.PaymentMethodCard,
		ExternalID: "flow_test_ext_id",
		RewardItems: []itemEntity.RewardItem{
			{ItemID: 1, Count: 5},
		},
	}

	processResponse, err := service.ProcessPayment(processRequest)
	assert.NoError(t, err)
	assert.NotNil(t, processResponse)
	assert.Equal(t, entity.PaymentStatusPending, processResponse.Status)

	// Step 2: Update payment status to completed
	updateRequest := UpdatePaymentStatusRequest{
		Status: entity.PaymentStatusCompleted,
	}

	updatedPayment, err := service.UpdatePaymentStatus(processResponse.PaymentID, updateRequest)
	assert.NoError(t, err)
	assert.NotNil(t, updatedPayment)
	assert.Equal(t, entity.PaymentStatusCompleted, updatedPayment.Status)

	mockRepo.AssertExpectations(t)
}