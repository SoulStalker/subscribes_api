package handler

import (
	"net/http"
	"time"

	"github.com/SoulStalker/subscribes_api/internal/domain"
	"github.com/SoulStalker/subscribes_api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	_ "github.com/SoulStalker/subscribes_api/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type SubscriptionHandler struct {
	service *service.SubscriptionService
	logger  *zap.Logger
}

func NewHandler(service *service.SubscriptionService, logger *zap.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		service: service,
		logger:  logger,
	}
}

func (h *SubscriptionHandler) InitRoutes(mode string) *gin.Engine {
	gin.SetMode(mode)
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(h.loggingMiddleware())

	api := router.Group("/api/v1")

	{
		subs := api.Group("/subscriptions")
		{
			subs.POST("", h.create)
			subs.GET("/:id", h.getByID)
			subs.GET("", h.list)
			subs.PUT("/:id", h.update)
			subs.DELETE("/:id", h.delete)
			subs.GET("/total-cost", h.totalCost)
		}
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}

func (h *SubscriptionHandler) loggingMiddleware() gin.HandlerFunc {
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

// @Summary Create subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body CreateSubscriptionRequest true "Subscription data"
// @Success 201 {object} SubscriptionResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/subscriptions [post]
func (h *SubscriptionHandler) create(c *gin.Context) {
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

// @Summary Get subscription by ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} SubscriptionResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/subscriptions/{id} [get]
func (h *SubscriptionHandler) getByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return
	}

	sub, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "subscription not found"})
		return
	}

	c.JSON(http.StatusOK, toResponse(sub))
}

// @Summary List subscriptions
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "Filter by user ID"
// @Param service_name query string false "Filter by service name"
// @Success 200 {array} SubscriptionResponse
// @Router /api/v1/subscriptions [get]
func (h *SubscriptionHandler) list(c *gin.Context) {
	filter := domain.SubscriptionFilter{}

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user id"})
			return
		}

		filter.UserID = &userID
	}

	if serviceName := c.Query("service_name"); serviceName != "" {
		filter.ServiceName = &serviceName
	}

	subs, err := h.service.List(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("failed to list subscriptions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	resp := make([]SubscriptionResponse, len(subs))
	for i, sub := range subs {
		resp[i] = toResponse(&sub)
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary Update subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Param subscription body UpdateSubscriptionRequest true "Updated data"
// @Success 200 {object} SubscriptionResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/subscriptions/{id} [put]
func (h *SubscriptionHandler) update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return
	}

	var req UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	startDate, _ := time.Parse("01-2006", req.StartDate)
	var endDate *time.Time
	if req.EndDate != "" {
		ed, _ := time.Parse("01-2006", req.EndDate)
		endDate = &ed
	}

	sub := &domain.Subscription{
		ID:          id,
		ServiceName: req.ServiceName,
		Price:       req.Price,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	if err := h.service.Update(c.Request.Context(), sub); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, toResponse(sub))
}

// @Summary Delete subscription
// @Tags subscriptions
// @Param id path string true "Subscription ID"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/subscriptions/{id} [delete]
func (h *SubscriptionHandler) delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "subscription not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary Calculate total cost
// @Tags subscriptions
// @Produce json
// @Param start_period query string true "Start period YYYY-MM-DD"
// @Param end_period query string true "End period YYYY-MM-DD"
// @Param user_id query string false "Filter by user ID"
// @Param service_name query string false "Filter by service name"
// @Success 200 {object} TotalCostResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/subscriptions/total-cost [get]
func (h *SubscriptionHandler) totalCost(c *gin.Context) {
	var req TotalCostRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	startPeriod, err := time.Parse("2006-01-02", req.StartPeriod)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid start_period"})
		return
	}

	endPeriod, err := time.Parse("2006-01-02", req.EndPeriod)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid end_period"})
		return
	}

	filter := domain.SubscriptionFilter{
		StartPeriod: &startPeriod,
		EndPeriod:   &endPeriod,
	}

	if req.UserID != nil {
		userID, err := uuid.Parse(*req.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user_id"})
			return
		}
		filter.UserID = &userID
	}

	if req.ServiceName != nil {
		filter.ServiceName = req.ServiceName
	}

	total, err := h.service.TotalCost(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, TotalCostResponse{Total: total})
}
