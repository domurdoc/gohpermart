package pg

import (
	"context"
	"database/sql"
	"time"

	"github.com/domurdoc/gophermart/internal/models"
)

type PGOrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *PGOrderRepository {
	return &PGOrderRepository{db: db}
}

func (r *PGOrderRepository) CreateOrder(ctx context.Context, user *models.User, orderNumber string) (*models.Order, error) {
	order := models.Order{
		Number:     orderNumber,
		Status:     models.StatusNew,
		Accrual:    0,
		UploadedAt: time.Now().In(time.UTC),
	}
	row := r.db.QueryRowContext(
		ctx,
		`
INSERT INTO orders (number, status, user_id, accrual, uploaded_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (number) DO UPDATE SET number = orders.number
RETURNING user_id, uploaded_at = $6
`,
		order.Number,
		order.Status,
		user.ID,
		order.Accrual,
		order.UploadedAt,
		order.UploadedAt,
	)
	var (
		currentOrderUserID int
		isInsertionTime    bool
	)
	err := row.Scan(
		&currentOrderUserID,
		&isInsertionTime,
	)
	if err != nil {
		return nil, err
	}
	if currentOrderUserID != user.ID {
		return nil, models.ErrOrderCreatedByAnotherUser
	}
	if !isInsertionTime {
		return nil, models.ErrOrderAlreadyCreatedByUser
	}
	return &order, nil
}

func (r *PGOrderRepository) GetUserOrders(ctx context.Context, user *models.User) ([]*models.Order, error) {
	var orders []*models.Order

	rows, err := r.db.QueryContext(
		ctx,
		`
SELECT number, status, accrual, uploaded_at FROM orders WHERE user_id = $1
`,
		user.ID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(
			&order.Number,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}
