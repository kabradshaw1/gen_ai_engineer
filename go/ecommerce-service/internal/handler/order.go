package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/model"
	"github.com/kabradshaw1/portfolio/go/pkg/apperror"
)

type OrderServiceInterface interface {
	Checkout(ctx context.Context, userID uuid.UUID) (*model.Order, error)
	GetOrder(ctx context.Context, orderID uuid.UUID) (*model.Order, error)
	ListOrders(ctx context.Context, userID uuid.UUID) ([]model.Order, error)
}

type OrderHandler struct {
	svc OrderServiceInterface
}

func NewOrderHandler(svc OrderServiceInterface) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) Checkout(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("userId"))
	if err != nil {
		_ = c.Error(apperror.BadRequest("INVALID_ID", "invalid user ID"))
		return
	}

	order, err := h.svc.Checkout(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) List(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("userId"))
	if err != nil {
		_ = c.Error(apperror.BadRequest("INVALID_ID", "invalid user ID"))
		return
	}

	orders, err := h.svc.ListOrders(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

func (h *OrderHandler) GetByID(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("userId"))
	if err != nil {
		_ = c.Error(apperror.BadRequest("INVALID_ID", "invalid user ID"))
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apperror.BadRequest("INVALID_ID", "invalid order ID"))
		return
	}

	order, err := h.svc.GetOrder(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Ownership check: 404 (not 403) to avoid leaking existence of other users' orders.
	if order.UserID != userID {
		_ = c.Error(apperror.NotFound("ORDER_NOT_FOUND", "order not found"))
		return
	}

	c.JSON(http.StatusOK, order)
}
