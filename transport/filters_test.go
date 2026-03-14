package transport

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestParseFilters(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		want    []Filter
		wantErr bool
	}{
		{
			name:  "Single filter",
			query: "filter=name:eq:test",
			want:  []Filter{{Field: "name", Op: OpEq, Value: "test"}},
		},
		{
			name:  "Multiple filters",
			query: "filter=name:eq:test|status:eq:active",
			want: []Filter{
				{Field: "name", Op: OpEq, Value: "test"},
				{Field: "status", Op: OpEq, Value: "active"},
			},
		},
		{
			name:  "Numeric and JSONB filters",
			query: "filter=price:gte:100|attributes.color:eq:Red",
			want: []Filter{
				{Field: "price", Op: OpGte, Value: "100"},
				{Field: "attributes.color", Op: OpEq, Value: "Red"},
			},
		},
		{
			name:  "JSONB key filter",
			query: `filter=attributes.color:eq:Red`,
			want: []Filter{
				{Field: "attributes.color", Op: OpEq, Value: "Red"},
			},
		},
		{
			name:  "JSONB contains filter",
			query: `filter=attributes:contains:{"color":"Red"}`,
			want: []Filter{
				{Field: "attributes", Op: OpContains, Value: `{"color":"Red"}`},
			},
		},
		{
			name:  "Numeric and JSONB contains filter",
			query: `filter=price:gte:100|attributes:contains:{"color":"Red","size":"xl"}`,
			want: []Filter{
				{Field: "price", Op: OpGte, Value: "100"},
				{Field: "attributes", Op: OpContains, Value: `{"color":"Red","size":"xl"}`},
			},
		},
		{
			name:  "No filters",
			query: "page=1&limit=10",
			want:  nil,
		},
		{
			name:    "Invalid format (missing parts)",
			query:   "filter=name:eq",
			wantErr: true,
		},
		{
			name:    "Invalid operator",
			query:   "filter=name:invalid:test",
			wantErr: true,
		},
		{
			name:    "Empty field",
			query:   "filter=:eq:test",
			wantErr: true,
		},
		{
			name:    "Empty value",
			query:   "filter=name:eq:",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, _ := url.Parse("http://localhost?" + tt.query)
			r := &http.Request{URL: u}

			got, err := ParseFilters(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFilters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseFilters() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseSort(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  *Sort
	}{
		{
			name:  "Ascending sort",
			query: "sort=price",
			want:  &Sort{Field: "price", Dir: SortAsc},
		},
		{
			name:  "Descending sort",
			query: "sort=-price",
			want:  &Sort{Field: "price", Dir: SortDesc},
		},
		{
			name:  "No sort",
			query: "page=1",
			want:  nil,
		},
		{
			name:  "Whitespace sort",
			query: "sort= -name ",
			want:  &Sort{Field: "name", Dir: SortDesc},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, _ := url.Parse("http://localhost?" + tt.query)
			r := &http.Request{URL: u}

			got := ParseSort(r)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseSort() = %v, want %v", got, tt.want)
			}
		})
	}
}
