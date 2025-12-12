package domain

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID  `json:"id"`
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	UserID      uuid.UUID  `json:"user_id"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type SubscriptionFilter struct {
	UserID      *uuid.UUID
	ServiceName *string
	StartPeriod *time.Time
	EndPeriod   *time.Time
}

type Pagination struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	SortBy   string `form:"sort_by"`
	SortDir  string `form:"sort_dir"`
}

func (p *Pagination) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

func (p *Pagination) Validate() {
	if p.Page == 0 {
		p.Page = 1
	}
	if p.PageSize == 0 {
		p.PageSize = 20
	}
	if p.SortBy == "" {
		p.SortBy = "created_at"
	}
	if p.SortDir == "" {
		p.SortDir = "DESC"
	}
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}
