package client

import (
	"context"

	"github.com/domurdoc/gophermart/internal/models"
)

type DebugBonusClient struct {
}

func (c *DebugBonusClient) GetForOrder(ctx context.Context, orderNumber string) (*models.Bonus, error) {
	bonus := models.Bonus{
		OrderNumber: orderNumber,
		Status:      models.StatusProcessed,
		Accrual:     100,
	}
	return &bonus, nil
}

func (c *DebugBonusClient) Close() error {
	return nil
}
