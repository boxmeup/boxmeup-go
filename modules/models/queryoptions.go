package models

import "math"

// SortType is a locked string
type SortType string

const (
	// ASC is ascending order
	ASC SortType = "ASC"
	// DSC is descending order
	DSC SortType = "DESC"
)

// SortBy is a construct for sorting query results
type SortBy struct {
	Field     string
	Direction SortType
}

// QueryLimit is a construct for limiting and paginating queries.
type QueryLimit struct {
	Limit  int
	Offset int
}

// SetPage will calculate the limit and offset based on page and size.
func (l *QueryLimit) SetPage(page int, size int) {
	if page <= 0 {
		page = 1
	}
	page--
	l.Limit = size
	l.Offset = page * size
}

// PagedResponse contains pagination meta data.
type PagedResponse struct {
	RequestTotal int `json:"request_total"`
	Total        int `json:"total"`
	Pages        int `json:"pages"`
}

// CalculatePages sets the number of pages based on a query limit.
func (r *PagedResponse) CalculatePages(limit QueryLimit) {
	calc := float64(r.Total) / float64(limit.Limit)
	r.Pages = int(math.Ceil(calc))
}
