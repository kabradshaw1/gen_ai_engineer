package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/model"
)

var ErrOrderNotOwned = errors.New("order not owned by user")

type ReturnRepositoryInterface interface {
	Create(ctx context.Context, orderID, userID uuid.UUID, itemIDs []string, reason string) (*model.Return, error)
}

type OrderLookup interface {
	GetOrder(ctx context.Context, orderID uuid.UUID) (*model.Order, error)
}

type ReturnService struct {
	returns ReturnRepositoryInterface
	orders  OrderLookup
}

func NewReturnService(returns ReturnRepositoryInterface, orders OrderLookup) *ReturnService {
	return &ReturnService{returns: returns, orders: orders}
}

// Initiate verifies the order belongs to userID before creating the return.
func (s *ReturnService) Initiate(ctx context.Context, userID, orderID uuid.UUID, itemIDs []string, reason string) (*model.Return, error) {
	order, err := s.orders.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.UserID != userID {
		return nil, ErrOrderNotOwned
	}
	return s.returns.Create(ctx, orderID, userID, itemIDs, reason)
}
