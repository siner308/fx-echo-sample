package dto

// ListResponse provides a unified wrapper for list responses (Stripe style)
type ListResponse[T any] struct {
	Object   string `json:"object"`            // Always "list"
	Data     []T    `json:"data"`              // Array of items
	HasMore  bool   `json:"has_more"`          // Whether there are more items
	TotalCount int  `json:"total_count,omitempty"` // Total count if available
}

// EmptyResponse for operations that don't return data (like delete)
type EmptyResponse struct {
	Deleted bool   `json:"deleted"`
	ID      string `json:"id"`
}

// Helper functions to create standard responses

// NewList creates a Stripe-style list response
func NewList[T any](data []T, hasMore ...bool) ListResponse[T] {
	more := false
	if len(hasMore) > 0 {
		more = hasMore[0]
	}
	
	return ListResponse[T]{
		Object:     "list",
		Data:       data,
		HasMore:    more,
		TotalCount: len(data),
	}
}

// NewEmpty creates an empty success response for deletions
func NewEmpty(id string) EmptyResponse {
	return EmptyResponse{
		Deleted: true,
		ID:      id,
	}
}

// Note: For single item success responses, return the data directly
// Example: return c.JSON(http.StatusOK, item.ToResponse())
// This follows Stripe's pattern of returning objects directly