package services

import (
	"context"

	"github.com/domurdoc/gophermart/internal/models"
	"github.com/domurdoc/gophermart/internal/repositories"
	"github.com/domurdoc/gophermart/internal/services/bonus"
	"github.com/domurdoc/gophermart/internal/utils"
)

type OrderService struct {
	orderRepo    repositories.OrderRepository
	bonusService *bonus.BonusService
}

func NewOrderService(orderRepo repositories.OrderRepository, bonusService *bonus.BonusService) *OrderService {
	return &OrderService{orderRepo: orderRepo, bonusService: bonusService}
}

func (s *OrderService) CreateOrder(ctx context.Context, user *models.User, orderNumer string) (*models.Order, error) {
	if err := ValidateOrderNumber(orderNumer); err != nil {
		return nil, err
	}
	order, err := s.orderRepo.CreateOrder(ctx, user, orderNumer)
	if err != nil {
		return nil, err
	}
	go s.bonusService.CheckOrder(order)
	return order, err
}

func (s *OrderService) GetUserOrders(ctx context.Context, user *models.User) ([]*models.Order, error) {
	return s.orderRepo.GetUserOrders(ctx, user)
}

func ValidateOrderNumber(orderNumber string) error {
	if !utils.IsLuhnNumber(orderNumber) {
		return models.ErrInvalidOrderNumber
	}
	return nil
}
