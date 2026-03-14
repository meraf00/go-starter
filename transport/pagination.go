package transport

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type paginationInt int

type numeric interface {
	~int | ~int32 | ~int64 | ~uint | ~uint64
}

type Pagination struct {
	Page     paginationInt `json:"page,omitempty"`
	PageSize paginationInt `json:"page_size,omitempty"`
}

type PaginationMetadata struct {
	TotalCount  int  `json:"total_count"`
	PageSize    int  `json:"page_size"`
	CurrentPage int  `json:"current_page"`
	TotalPages  int  `json:"total_pages"`
	HasNext     bool `json:"has_next"`
	HasPrev     bool `json:"has_prev"`
}

func (c paginationInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(c)
}

// UnmarshalJSON allows parsing both strings and numbers into PaginationInt
func (c *paginationInt) UnmarshalJSON(b []byte) error {

	var num int
	if err := json.Unmarshal(b, &num); err == nil {
		*c = paginationInt(num)
		return nil
	}

	// Fallback: try parsing as string
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}

	n, err := strconv.Atoi(str)
	if err != nil {
		return fmt.Errorf("invalid pagination value: %s", str)
	}
	*c = paginationInt(n)
	return nil
}

func CalculateMetadata(totalCount, limit, offset int) PaginationMetadata {
	totalPages := 0
	if limit > 0 {
		totalPages = (totalCount + limit - 1) / limit
	}

	currentPage := 1
	if limit > 0 {
		currentPage = (offset / limit) + 1
	}

	return PaginationMetadata{
		TotalCount:  totalCount,
		PageSize:    limit,
		CurrentPage: currentPage,
		TotalPages:  totalPages,
		HasNext:     currentPage < totalPages,
		HasPrev:     currentPage > 1,
	}
}

func PaginatedResponse[T any](items []T, totalCount, limit, offset int) any {
	if len(items) == 0 {
		items = []T{}
	}
	return map[string]any{
		"items": items,
		"meta":  CalculateMetadata(totalCount, limit, offset),
	}
}

func LimitOffset[T numeric](page, pageSize T) (int, int) {
	limit := 100
	offset := 0

	if pageSize <= 0 {
		limit = 10
	} else if pageSize > 100 {
		limit = 100
	} else {
		limit = int(pageSize)
	}

	if page <= 0 {
		offset = 0
	} else {
		offset = (int(page) - 1) * limit
	}

	return limit, offset
}

func ParseLimitOffset(page, pageSize string) (int, int, error) {
	limit := 100
	offset := 0
	var err error

	if pageSize != "" {
		limit, err = strconv.Atoi(pageSize)
		if err != nil || limit <= 0 {
			return 0, 0, BadRequest("limit must be a positive integer", "query_validation", nil)
		}
		if limit > 100 {
			limit = 100
		}
	}

	if page != "" {
		page, err := strconv.Atoi(page)
		if err != nil || page < 1 {
			return 0, 0, BadRequest("page must be a positive integer", "query_validation", nil)
		}
		offset = (page - 1) * limit
	}

	return limit, offset, nil
}
