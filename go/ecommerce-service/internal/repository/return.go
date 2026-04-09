package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/model"
)

type ReturnRepository struct {
	pool *pgxpool.Pool
}

func NewReturnRepository(pool *pgxpool.Pool) *ReturnRepository {
	return &ReturnRepository{pool: pool}
}

func (r *ReturnRepository) Create(ctx context.Context, orderID, userID uuid.UUID, itemIDs []string, reason string) (*model.Return, error) {
	itemsJSON, err := json.Marshal(itemIDs)
	if err != nil {
		return nil, fmt.Errorf("marshal itemIDs: %w", err)
	}

	var ret model.Return
	var returnedItems []byte
	err = r.pool.QueryRow(ctx,
		`INSERT INTO returns (order_id, user_id, status, reason, item_ids)
		 VALUES ($1, $2, 'requested', $3, $4)
		 RETURNING id, order_id, user_id, status, reason, item_ids, created_at, updated_at`,
		orderID, userID, reason, itemsJSON,
	).Scan(&ret.ID, &ret.OrderID, &ret.UserID, &ret.Status, &ret.Reason, &returnedItems, &ret.CreatedAt, &ret.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert return: %w", err)
	}
	if err := json.Unmarshal(returnedItems, &ret.ItemIDs); err != nil {
		return nil, fmt.Errorf("unmarshal item_ids: %w", err)
	}
	return &ret, nil
}
