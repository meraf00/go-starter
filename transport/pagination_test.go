package transport

import (
	"encoding/json"
	"testing"
)

func TestCalculateMetadata(t *testing.T) {
	tests := []struct {
		name       string
		totalCount int
		limit      int
		offset     int
		want       PaginationMetadata
	}{
		{
			name:       "First page of many",
			totalCount: 100,
			limit:      10,
			offset:     0,
			want: PaginationMetadata{
				TotalCount:  100,
				PageSize:    10,
				CurrentPage: 1,
				TotalPages:  10,
				HasNext:     true,
				HasPrev:     false,
			},
		},
		{
			name:       "Middle page",
			totalCount: 100,
			limit:      10,
			offset:     20,
			want: PaginationMetadata{
				TotalCount:  100,
				PageSize:    10,
				CurrentPage: 3,
				TotalPages:  10,
				HasNext:     true,
				HasPrev:     true,
			},
		},
		{
			name:       "Last page exactly",
			totalCount: 100,
			limit:      10,
			offset:     90,
			want: PaginationMetadata{
				TotalCount:  100,
				PageSize:    10,
				CurrentPage: 10,
				TotalPages:  10,
				HasNext:     false,
				HasPrev:     true,
			},
		},
		{
			name:       "Empty results",
			totalCount: 0,
			limit:      10,
			offset:     0,
			want: PaginationMetadata{
				TotalCount:  0,
				PageSize:    10,
				CurrentPage: 1,
				TotalPages:  0,
				HasNext:     false,
				HasPrev:     false,
			},
		},
		{
			name:       "Partial last page",
			totalCount: 25,
			limit:      10,
			offset:     20,
			want: PaginationMetadata{
				TotalCount:  25,
				PageSize:    10,
				CurrentPage: 3,
				TotalPages:  3,
				HasNext:     false,
				HasPrev:     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateMetadata(tt.totalCount, tt.limit, tt.offset)
			if got != tt.want {
				t.Errorf("CalculateMetadata() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestPaginatedResponse(t *testing.T) {
	items := []string{"a", "b"}
	totalCount := 10
	limit := 2
	offset := 0

	resp := PaginatedResponse(items, totalCount, limit, offset)

	// Check if it's a map
	m, ok := resp.(map[string]any)
	if !ok {
		t.Fatal("PaginatedResponse should return a map")
	}

	// Check items
	if _, ok := m["items"]; !ok {
		t.Error("Missing 'items' in response")
	}

	// Check meta
	meta, ok := m["meta"].(PaginationMetadata)
	if !ok {
		t.Error("Missing or invalid 'meta' in response")
	}

	if meta.TotalCount != totalCount {
		t.Errorf("Expected total_count %d, got %d", totalCount, meta.TotalCount)
	}
}

func TestPaginationIntUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    paginationInt
		wantErr bool
	}{
		{"Number", `10`, 10, false},
		{"String number", `"20"`, 20, false},
		{"Invalid string", `"abc"`, 0, true},
		{"Empty string", `""`, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p paginationInt
			err := json.Unmarshal([]byte(tt.input), &p)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && p != tt.want {
				t.Errorf("UnmarshalJSON() = %v, want %v", p, tt.want)
			}
		})
	}
}
