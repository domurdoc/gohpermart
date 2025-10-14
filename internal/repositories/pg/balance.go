package pg

import (
	"context"
	"database/sql"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/domurdoc/gophermart/internal/models"
	"github.com/domurdoc/gophermart/internal/repositories"
)

type PGBalanceRepository struct {
	db *sql.DB
}

func NewBalanceRepository(db *sql.DB) *PGBalanceRepository {
	return &PGBalanceRepository{db: db}
}

func (r *PGBalanceRepository) GetUserBalance(ctx context.Context, user *models.User) (*models.Balance, error) {
	row := r.db.QueryRowContext(
		ctx,
		"SELECT current, withdrawn FROM users WHERE id = $1",
		user.ID,
	)
	var balance models.Balance
	err := row.Scan(
		&balance.Current,
		&balance.Withdrawn,
	)
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

func (r *PGBalanceRepository) Withdraw(ctx context.Context, user *models.User, orderNumber string, sum float64) (*models.Withdrawal, error) {
	withdrawal := models.Withdrawal{
		OrderNumber: orderNumber,
		Sum:         sum,
		ProcessedAt: time.Now().In(time.UTC),
	}

	var balance models.Balance

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	row := tx.QueryRowContext(
		ctx,
		"SELECT current, withdrawn FROM users WHERE id = $1 FOR UPDATE",
		user.ID,
	)
	err = row.Scan(
		&balance.Current,
		&balance.Withdrawn,
	)
	if err != nil {
		return nil, err
	}
	if balance.Current < withdrawal.Sum {
		return nil, models.ErrNotEnoughBalance
	}
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO withdrawals (user_id, order_number, sum, processed_at) VALUES($1, $2, $3, $4)",
		user.ID,
		withdrawal.OrderNumber,
		withdrawal.Sum,
		withdrawal.ProcessedAt,
	)
	if err != nil {
		return nil, err
	}
	_, err = tx.ExecContext(
		ctx,
		"UPDATE users SET current = $1, withdrawn = $2 WHERE id = $3",
		balance.Current-withdrawal.Sum,
		balance.Withdrawn+withdrawal.Sum,
		user.ID,
	)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return &withdrawal, nil
}

func (r *PGBalanceRepository) GetUserWithdrawals(ctx context.Context, user *models.User) ([]*models.Withdrawal, error) {
	var withdrawals []*models.Withdrawal

	rows, err := r.db.QueryContext(
		ctx,
		"SELECT order_number, sum, processed_at FROM withdrawals WHERE user_id = $1",
		user.ID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var withdrawal models.Withdrawal
		if err := rows.Scan(
			&withdrawal.OrderNumber,
			&withdrawal.Sum,
			&withdrawal.ProcessedAt,
		); err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, &withdrawal)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return withdrawals, nil
}

func (r *PGBalanceRepository) AccrueBonuses(ctx context.Context, orders []*models.Order) ([]int, error) {
	updateOrdersQueryTemplate := `
UPDATE
    orders o
SET
    accrual = tmp.accrual,
    status = tmp.status,
    version = tmp.version + 1
FROM
    (
        VALUES
            %s
    ) AS tmp (number, accrual, status, version)
WHERE
    o.number = tmp.number
    AND o.version = tmp.version
RETURNING o.user_id
`
	args := make([]any, 0, len(orders)*4)
	placeholders := make([]string, len(orders))
	for i, order := range orders {
		args = append(args, order.Number, order.Accrual, order.Status, order.Version)
		placeholders[i] = fmt.Sprintf("($%d::VARCHAR, $%d::NUMERIC, $%d::VARCHAR, $%d::INTEGER)", i*4+1, i*4+2, i*4+3, i*4+4)
	}

	updateOrdersQuery := fmt.Sprintf(updateOrdersQueryTemplate, strings.Join(placeholders, ","))

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, updateOrdersQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userIDSMap := make(map[int]struct{})
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDSMap[userID] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	userIDS := slices.Collect(maps.Keys(userIDSMap))

	updateUsersQueryTemplate := `
UPDATE
	users
SET
	is_stale = TRUE
WHERE
	id IN (%s)
`
	args = make([]any, len(userIDS))
	placeholders = make([]string, len(userIDS))
	for i, userID := range userIDS {
		args[i] = userID
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	updateUsersQuery := fmt.Sprintf(updateUsersQueryTemplate, strings.Join(placeholders, ","))
	_, err = tx.ExecContext(
		ctx,
		updateUsersQuery,
		args...,
	)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return userIDS, nil
}

func (r *PGBalanceRepository) GetOrdersToCheck(ctx context.Context) ([]*models.Order, error) {
	getOrdersToCheckQuery := `
SELECT number, status, accrual, uploaded_at, version
FROM orders WHERE status IN ('NEW', 'PROCESSING')
`
	rows, err := r.db.QueryContext(ctx, getOrdersToCheckQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(
			&order.Number,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
			&order.Version,
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

func (r *PGBalanceRepository) GetUsersToRefresh(ctx context.Context) ([]int, error) {
	query := `
SELECT id FROM users WHERE is_stale = TRUE
`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDS []int
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDS = append(userIDS, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return userIDS, nil
}

func (r *PGBalanceRepository) RefreshUsers(ctx context.Context, userIDS []int) error {
	query := `
UPDATE users
SET
	current = COALESCE((SELECT SUM(o.accrual) FROM orders o WHERE o.user_id = id), 0) - COALESCE((SELECT SUM(w1.sum) FROM withdrawals w1 WHERE w1.user_id = id), 0),
	withdrawn = COALESCE((SELECT SUM(w.sum) FROM withdrawals w WHERE w.user_id = id), 0),
	is_stale = FALSE
WHERE
	id IN (SELECT u1.id FROM users u1 WHERE u1.id IN (%s) FOR UPDATE SKIP LOCKED)
RETURNING id
`
	args := make([]any, len(userIDS))
	placeholders := make([]string, len(userIDS))
	for i, userID := range userIDS {
		args[i] = userID
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query = fmt.Sprintf(query, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	updatedUserIDS := make([]int, 0, len(userIDS))
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			return err
		}
		updatedUserIDS = append(updatedUserIDS, userID)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(userIDS) != len(updatedUserIDS) {
		skippedUserIDS := getSkippedUserIDS(userIDS, updatedUserIDS)
		return &repositories.BalanceConflictError{UserIDS: skippedUserIDS}
	}
	return nil
}

func getSkippedUserIDS(updatedUserIDS []int, userIDS []int) []int {
	m := make(map[int]struct{})
	var skipped []int
	for _, userID := range updatedUserIDS {
		m[userID] = struct{}{}
	}
	for _, userID := range userIDS {
		if _, ok := m[userID]; !ok {
			skipped = append(skipped, userID)
		}
	}
	return skipped
}
