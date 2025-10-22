package services

import (
	"context"

	"github.com/domurdoc/gophermart/internal/models"
	"github.com/domurdoc/gophermart/internal/repositories"
)

type BalanceService struct {
	balanceRepo repositories.BalanceRepository
}

func NewBalanceService(balanceRepo repositories.BalanceRepository) *BalanceService {
	return &BalanceService{balanceRepo: balanceRepo}
}

func (s *BalanceService) GetUserBalance(ctx context.Context, user *models.User) (*models.Balance, error) {
	return s.balanceRepo.GetUserBalance(ctx, user)
}

func (s *BalanceService) Withdraw(ctx context.Context, user *models.User, orderNumber string, sum float64) (*models.Withdrawal, error) {
	if err := ValidateOrderNumber(orderNumber); err != nil {
		return nil, err
	}
	return s.balanceRepo.Withdraw(ctx, user, orderNumber, sum)
}

func (s *BalanceService) GetUserWithdrawals(ctx context.Context, user *models.User) ([]*models.Withdrawal, error) {
	return s.balanceRepo.GetUserWithdrawals(ctx, user)
}
