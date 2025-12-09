package handler

import (
	"net/http"
	"time"

	"github.com/SoulStalker/subscribes_api/internal/domain"
	"github.com/SoulStalker/subscribes_api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Handler struct {
	service *service.SubscriptionService
	logger  *zap.Logger
}

func NewHandler(service *service.SubscriptionService, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) InitRoutes(mode string) *gin.Engine {
	gin.SetMode(mode)
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(h.loggingMiddleware())

	api := router.Group("/api/v1")

	{
		subs := api.Group("/subscriptions")
		{
			subs.POST("", h.create)
			// subs.GET("/:id", h.getByID)
			// subs.GET("", h.list)
			// subs.PUT("/:id", h.update)
			// subs.DELETE("/:id", h.delete)
			// subs.GET("/total-cost", h.totalCost)
		}
	}

	return router
}

func (h *Handler) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		h.logger.Info("request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
		)
	}
}

func (h *Handler) create(c *gin.Context) {
	var req CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user_id"})
		return
	}

	startDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid start_date format"})
		return
	}

	var endDate *time.Time
	if req.EndDate != "" {
		ed, err := time.Parse("01-2006", req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid end_date format"})
			return
		}
		endDate = &ed
	}

	sub := &domain.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	if err := h.service.Create(c.Request.Context(), sub); err != nil {
		h.logger.Error("failed to create subscription", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toResponse(sub))
}

func toResponse(sub *domain.Subscription) SubscriptionResponse {
	resp := SubscriptionResponse{
		ID:          sub.ID.String(),
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		UserID:      sub.UserID.String(),
		StartDate:   sub.StartDate.Format("2006-01-02"),
		CreatedAt:   sub.CreatedAt,
		UpdatedAt:   sub.UpdatedAt,
	}

	if sub.EndDate != nil {
		endStr := sub.EndDate.Format("2006-01-02")
		resp.EndDate = &endStr
	}

	return resp
}
