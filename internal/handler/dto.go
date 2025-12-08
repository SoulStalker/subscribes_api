package handler

import "time"

type CreateSubscriptionRequest struct {
	ServiceName string `json:"service_name" binding:"required" example:"Yandex Plus"`
	Price       int    `json:"price" binding:"required,min=0" example:"400"`
	UserID      string `json:"user_id" binding:"required,uuid" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string `json:"start_date" binding:"required" example:"07-2025"`
	EndDate     string `json:"end_date,omitempty" example:"12-2025"`
}

type UpdateSubscriptionRequest struct {
	ServiceName string `json:"service_name" binding:"required"`
	Price       int    `json:"price" binding:"required,min=0"`
	StartDate   string `json:"start_date" binding:"required"`
	EndDate     string `json:"end_date,omitempty"`
}

type SubscriptionResponse struct {
	ID          string    `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	ServiceName string    `json:"service_name" example:"Yandex Plus"`
	Price       int       `json:"price" example:"400"`
	UserID      string    `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string    `json:"start_date" example:"2025-07-01"`
	EndDate     *string   `json:"end_date,omitempty" example:"2025-12-01"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TotalCostRequest struct {
	StartPeriod string  `form:"start_period" binding:"required" example:"2025-01-01"`
	EndPeriod   string  `form:"end_period" binding:"required" example:"2025-12-31"`
	UserID      *string `form:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	ServiceName *string `form:"service_name" example:"Yandex"`
}

type TotalCostResponse struct {
	Total int `json:"total" example:"12000"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"invalid request"`
}
