package models

import "time"

// CursorListOptions configures cursor-based pagination and filtering for list
// operations.
type CursorListOptions struct {
	Limit     int               `json:"limit,omitempty"`
	Cursor    string            `json:"cursor,omitempty"`
	SortBy    string            `json:"sortBy,omitempty"`
	SortOrder string            `json:"sortOrder,omitempty"`
	StartDate *time.Time        `json:"startDate,omitempty"`
	EndDate   *time.Time        `json:"endDate,omitempty"`
	Filters   map[string]string `json:"filters,omitempty"`
}

// PageListOptions configures page-based pagination and filtering for list
// operations that do not support cursor semantics.
type PageListOptions struct {
	PageNumber int               `json:"pageNumber,omitempty"`
	PageSize   int               `json:"pageSize,omitempty"`
	SortOrder  string            `json:"sortOrder,omitempty"`
	Filters    map[string]string `json:"filters,omitempty"`
}

// ListResponse is the generic paginated response envelope.
type ListResponse[T any] struct {
	Items      []T        `json:"items"`
	Pagination Pagination `json:"pagination,omitempty"`
}

// Pagination holds pagination metadata returned by list endpoints.
type Pagination struct {
	Total      int    `json:"total"`
	Page       int    `json:"page,omitempty"`
	Limit      int    `json:"limit"`
	NextCursor string `json:"nextCursor,omitempty"`
	PrevCursor string `json:"prevCursor,omitempty"`
}

// Status represents an entity status with an optional description.
type Status struct {
	Code        string  `json:"code"`
	Description *string `json:"description,omitempty"`
}

// Metadata represents key-value pairs attached to entities.
// When deserialized from JSON, Metadata may be nil if the field was null or absent.
// Always use the Set method or check for nil before direct assignment to avoid panics.
type Metadata map[string]any

// Set safely assigns a key-value pair. Returns a new Metadata map if the receiver is nil.
// Usage: org.Metadata = org.Metadata.Set("key", "value")
func (m Metadata) Set(key string, value any) Metadata {
	if m == nil {
		m = make(Metadata)
	}

	m[key] = value

	return m
}

// Get safely retrieves a value by key. Returns (value, true) if found, (nil, false) otherwise.
// Safe to call on nil Metadata.
func (m Metadata) Get(key string) (any, bool) {
	if m == nil {
		return nil, false
	}

	v, ok := m[key]

	return v, ok
}
