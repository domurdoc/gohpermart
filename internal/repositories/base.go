package repositories

import (
	"context"
	"fmt"

	"github.com/domurdoc/gophermart/internal/models"
)

type BalanceConflictError struct {
	UserIDS []int
}

func (e *BalanceConflictError) Error() string {
	return fmt.Sprintf("failed to update users: %v", e.UserIDS)
}

type BalanceRepository interface {
	GetUserBalance(context.Context, *models.User) (*models.Balance, error)
	Withdraw(context.Context, *models.User, string, float64) (*models.Withdrawal, error)
	GetUserWithdrawals(context.Context, *models.User) ([]*models.Withdrawal, error)
	AccrueBonuses(context.Context, []*models.Order) ([]int, error)
	GetOrdersToCheck(context.Context) ([]*models.Order, error)
	GetUsersToRefresh(context.Context) ([]int, error)
	RefreshUsers(context.Context, []int) error
}

type OrderRepository interface {
	CreateOrder(context.Context, *models.User, string) (*models.Order, error)
	GetUserOrders(context.Context, *models.User) ([]*models.Order, error)
}

type UserRepository interface {
	CreateUser(context.Context, string, string) (*models.User, error)
	GetUserByID(context.Context, int) (*models.User, error)
	GetUserByUsername(context.Context, string) (*models.User, error)
}
