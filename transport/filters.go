package transport

import (
	"fmt"
	"net/http"
	"strings"
)

// FilterOp represents a comparison operator for filtering.
type FilterOp string

const (
	OpEq       FilterOp = "eq"
	OpNeq      FilterOp = "neq"
	OpGt       FilterOp = "gt"
	OpGte      FilterOp = "gte"
	OpLt       FilterOp = "lt"
	OpLte      FilterOp = "lte"
	OpLike     FilterOp = "like"
	OpContains FilterOp = "contains"
)

var validOps = map[FilterOp]bool{
	OpEq: true, OpNeq: true,
	OpGt: true, OpGte: true,
	OpLt: true, OpLte: true,
	OpLike: true, OpContains: true,
}

const partDelimiter = "|"

// Filter represents a single filter condition parsed from the query string.
type Filter struct {
	Field string   `json:"field"`
	Op    FilterOp `json:"op"`
	Value string   `json:"value"`
}

// SortDir represents a sort direction.
type SortDir string

const (
	SortAsc  SortDir = "ASC"
	SortDesc SortDir = "DESC"
)

// Sort represents a parsed sort directive.
type Sort struct {
	Field string  `json:"field"`
	Dir   SortDir `json:"dir"`
}

// ParseFilters parses the "filter" query parameter.
// Format: filter=field:op:value|field:op:value
// Example: filter=category_id:eq:abc123|price:gte:100|price:lte:500
func ParseFilters(r *http.Request) ([]Filter, error) {
	raw := r.URL.Query().Get("filter")
	if raw == "" {
		return nil, nil
	}

	parts := strings.Split(raw, partDelimiter)
	filters := make([]Filter, 0, len(parts))

	for _, part := range parts {
		segments := strings.SplitN(part, ":", 3)
		if len(segments) != 3 {
			return nil, BadRequest(
				fmt.Sprintf("invalid filter format: %q, expected field:op:value", part),
				"filter_validation", nil,
			)
		}

		field := strings.TrimSpace(segments[0])
		op := FilterOp(strings.TrimSpace(segments[1]))
		value := strings.TrimSpace(segments[2])

		if field == "" || value == "" {
			return nil, BadRequest(
				fmt.Sprintf("filter field and value must not be empty: %q", part),
				"filter_validation", nil,
			)
		}

		if !validOps[op] {
			return nil, BadRequest(
				fmt.Sprintf("invalid filter operator: %q, expected one of eq,neq,gt,gte,lt,lte,like,contains", op),
				"filter_validation", nil,
			)
		}

		filters = append(filters, Filter{Field: field, Op: op, Value: value})
	}

	return filters, nil
}

// ParseSort parses the "sort" query parameter.
// Format: sort=field (ascending) or sort=-field (descending)
// Example: sort=-price
func ParseSort(r *http.Request) *Sort {
	raw := r.URL.Query().Get("sort")
	if raw == "" {
		return nil
	}

	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "-") {
		return &Sort{Field: raw[1:], Dir: SortDesc}
	}
	return &Sort{Field: raw, Dir: SortAsc}
}
