package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/model"
	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/service"
)

type ReturnServiceInterface interface {
	Initiate(ctx context.Context, userID, orderID uuid.UUID, itemIDs []string, reason string) (*model.Return, error)
}

type ReturnHandler struct {
	svc ReturnServiceInterface
}

func NewReturnHandler(svc ReturnServiceInterface) *ReturnHandler {
	return &ReturnHandler{svc: svc}
}

func (h *ReturnHandler) Initiate(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}
	var req model.InitiateReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ret, err := h.svc.Initiate(c.Request.Context(), userID, orderID, req.ItemIDs, req.Reason)
	if err != nil {
		if errors.Is(err, service.ErrOrderNotOwned) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initiate return"})
		return
	}
	c.JSON(http.StatusCreated, ret)
}
