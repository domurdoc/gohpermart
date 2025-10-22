package client

import (
	"context"
	"errors"

	"github.com/domurdoc/gophermart/internal/models"
)

var ErrOrderNotRegistered = errors.New("order not registered")

type BonusClient interface {
	GetForOrder(context.Context, string) (*models.Bonus, error)
	Close() error
}
