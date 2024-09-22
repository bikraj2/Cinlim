package data

import (
	"math"
	"strings"

	"cinlim.bikraj.net/internal/validator"
)

type Filter struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
}

type PageMetaData struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

func ValidateFilters(v *validator.Validator, f Filter) {
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "must be lesser than 10 million")
	v.Check(f.PageSize > 0, "page_size", "Must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "Must be lesser than one Hundred")

	// Check if the sort paramter is in the SortSafeList
	v.Check(validator.In(f.Sort, f.SortSafeList...), "sort", "Invalid sort value")
}

func (f Filter) sortColumn() string {
	for _, safeValue := range f.SortSafeList {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	panic("unsafe sort parameter: " + f.Sort)
}
func (f Filter) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	} else {
		return "ASC"
	}
}
func (f Filter) limit() int {
	return f.PageSize
}
func (f Filter) offset() int {
	return (f.PageSize - 1) * f.Page
}
func calculateMetadata(totalRecords, page, page_size int) PageMetaData {
	if totalRecords == 0 {
		return PageMetaData{}
	}
	return PageMetaData{
		CurrentPage:  page,
		PageSize:     page_size,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(page_size))),
		TotalRecords: totalRecords,
	}
}
