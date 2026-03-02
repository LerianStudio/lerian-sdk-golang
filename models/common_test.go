package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ptr returns a pointer to the given value — handy for optional fields in tests.
func ptr[T any](v T) *T { return &v }

func TestListOptionsJSON(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	opts := ListOptions{
		Limit:     25,
		Page:      2,
		Cursor:    "abc123",
		SortBy:    "createdAt",
		SortOrder: "desc",
		StartDate: &start,
		EndDate:   &end,
		Filters:   map[string]string{"status": "active"},
	}

	data, err := json.Marshal(opts)
	require.NoError(t, err)

	raw := make(map[string]json.RawMessage)
	require.NoError(t, json.Unmarshal(data, &raw))

	// Verify camelCase keys are present.
	for _, key := range []string{"limit", "page", "cursor", "sortBy", "sortOrder", "startDate", "endDate", "filters"} {
		assert.Contains(t, raw, key, "expected camelCase key %q in JSON output", key)
	}

	// Round-trip back.
	var decoded ListOptions
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, opts.Limit, decoded.Limit)
	assert.Equal(t, opts.Page, decoded.Page)
	assert.Equal(t, opts.Cursor, decoded.Cursor)
	assert.Equal(t, opts.SortBy, decoded.SortBy)
	assert.Equal(t, opts.SortOrder, decoded.SortOrder)
	assert.True(t, opts.StartDate.Equal(*decoded.StartDate))
	assert.True(t, opts.EndDate.Equal(*decoded.EndDate))
	assert.Equal(t, opts.Filters, decoded.Filters)
}

func TestListOptionsOmitempty(t *testing.T) {
	data, err := json.Marshal(ListOptions{})
	require.NoError(t, err)
	assert.JSONEq(t, `{}`, string(data))
}

func TestListResponseGeneric(t *testing.T) {
	resp := ListResponse[string]{
		Items: []string{"a", "b"},
		Pagination: Pagination{
			Total: 2,
			Limit: 10,
		},
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded ListResponse[string]
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, resp.Items, decoded.Items)
	assert.Equal(t, resp.Pagination.Total, decoded.Pagination.Total)
	assert.Equal(t, resp.Pagination.Limit, decoded.Pagination.Limit)
}

func TestListResponseWithInt(t *testing.T) {
	resp := ListResponse[int]{
		Items: []int{1, 2, 3},
		Pagination: Pagination{
			Total: 3,
			Limit: 50,
		},
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded ListResponse[int]
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, []int{1, 2, 3}, decoded.Items)
	assert.Equal(t, 3, decoded.Pagination.Total)
	assert.Equal(t, 50, decoded.Pagination.Limit)
}

func TestPaginationJSON(t *testing.T) {
	t.Run("all fields", func(t *testing.T) {
		p := Pagination{
			Total:      100,
			Page:       3,
			Limit:      25,
			NextCursor: "next_abc",
			PrevCursor: "prev_xyz",
		}

		data, err := json.Marshal(p)
		require.NoError(t, err)

		raw := make(map[string]json.RawMessage)
		require.NoError(t, json.Unmarshal(data, &raw))

		assert.Contains(t, raw, "total")
		assert.Contains(t, raw, "page")
		assert.Contains(t, raw, "limit")
		assert.Contains(t, raw, "nextCursor")
		assert.Contains(t, raw, "prevCursor")
	})

	t.Run("required fields only", func(t *testing.T) {
		p := Pagination{
			Total: 42,
			Limit: 10,
		}

		data, err := json.Marshal(p)
		require.NoError(t, err)

		raw := make(map[string]json.RawMessage)
		require.NoError(t, json.Unmarshal(data, &raw))

		assert.Contains(t, raw, "total")
		assert.Contains(t, raw, "limit")
		assert.NotContains(t, raw, "page")
		assert.NotContains(t, raw, "nextCursor")
		assert.NotContains(t, raw, "prevCursor")
	})
}

func TestStatusJSON(t *testing.T) {
	desc := "Account is active"
	s := Status{
		Code:        "ACTIVE",
		Description: &desc,
	}

	data, err := json.Marshal(s)
	require.NoError(t, err)

	raw := make(map[string]json.RawMessage)
	require.NoError(t, json.Unmarshal(data, &raw))

	assert.Contains(t, raw, "code")
	assert.Contains(t, raw, "description")

	var decoded Status
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, "ACTIVE", decoded.Code)
	require.NotNil(t, decoded.Description)
	assert.Equal(t, "Account is active", *decoded.Description)
}

func TestStatusNilDescription(t *testing.T) {
	s := Status{Code: "ACTIVE"}

	data, err := json.Marshal(s)
	require.NoError(t, err)

	raw := make(map[string]json.RawMessage)
	require.NoError(t, json.Unmarshal(data, &raw))

	assert.Contains(t, raw, "code")
	assert.NotContains(t, raw, "description", "nil description must be omitted from JSON")
}

func TestMetadataJSON(t *testing.T) {
	m := Metadata{
		"key":    "value",
		"nested": map[string]any{"a": 1},
	}

	data, err := json.Marshal(m)
	require.NoError(t, err)

	var decoded Metadata
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "value", decoded["key"])

	nested, ok := decoded["nested"].(map[string]any)
	require.True(t, ok, "nested should unmarshal as map[string]any")
	// JSON numbers unmarshal as float64 by default.
	assert.Equal(t, float64(1), nested["a"])
}

func TestMetadataEmpty(t *testing.T) {
	m := Metadata{}

	data, err := json.Marshal(m)
	require.NoError(t, err)
	assert.JSONEq(t, `{}`, string(data))
}

func TestMetadata_SetOnNil(t *testing.T) {
	var m Metadata // nil
	result := m.Set("key", "value")

	require.NotNil(t, result, "Set on nil Metadata must return a new map")
	assert.Equal(t, "value", result["key"])
}

func TestMetadata_SetOnExisting(t *testing.T) {
	m := Metadata{"existing": "data"}
	result := m.Set("newKey", 42)

	assert.Equal(t, "data", result["existing"], "existing entries must be preserved")
	assert.Equal(t, 42, result["newKey"], "new entry must be present")
}

func TestMetadata_GetOnNil(t *testing.T) {
	var m Metadata // nil
	v, ok := m.Get("anything")

	assert.Nil(t, v)
	assert.False(t, ok, "Get on nil Metadata must return false")
}

func TestMetadata_GetExisting(t *testing.T) {
	m := Metadata{"color": "blue"}
	v, ok := m.Get("color")

	assert.Equal(t, "blue", v)
	assert.True(t, ok)
}

func TestMetadata_GetMissing(t *testing.T) {
	m := Metadata{"color": "blue"}
	v, ok := m.Get("missing")

	assert.Nil(t, v)
	assert.False(t, ok, "Get with missing key must return false")
}

func TestAddressJSON(t *testing.T) {
	addr := Address{
		Line1:   "123 Main St",
		Line2:   ptr("Apt 4B"),
		ZipCode: "10001",
		City:    "New York",
		State:   "NY",
		Country: "US",
	}

	data, err := json.Marshal(addr)
	require.NoError(t, err)

	raw := make(map[string]json.RawMessage)
	require.NoError(t, json.Unmarshal(data, &raw))

	for _, key := range []string{"line1", "line2", "zipCode", "city", "state", "country"} {
		assert.Contains(t, raw, key, "expected key %q in JSON output", key)
	}
}

func TestAddressOptionalLine2(t *testing.T) {
	addr := Address{
		Line1:   "456 Oak Ave",
		ZipCode: "90210",
		City:    "Beverly Hills",
		State:   "CA",
		Country: "US",
	}

	data, err := json.Marshal(addr)
	require.NoError(t, err)

	raw := make(map[string]json.RawMessage)
	require.NoError(t, json.Unmarshal(data, &raw))

	assert.NotContains(t, raw, "line2", "nil Line2 must be omitted from JSON")
	assert.Contains(t, raw, "line1")
	assert.Contains(t, raw, "zipCode")
}

func TestAddressRoundTrip(t *testing.T) {
	original := Address{
		Line1:   "789 Pine Rd",
		Line2:   ptr("Suite 100"),
		ZipCode: "30301",
		City:    "Atlanta",
		State:   "GA",
		Country: "US",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Address
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, original.Line1, decoded.Line1)
	require.NotNil(t, decoded.Line2)
	assert.Equal(t, *original.Line2, *decoded.Line2)
	assert.Equal(t, original.ZipCode, decoded.ZipCode)
	assert.Equal(t, original.City, decoded.City)
	assert.Equal(t, original.State, decoded.State)
	assert.Equal(t, original.Country, decoded.Country)
}
