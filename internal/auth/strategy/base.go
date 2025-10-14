package strategy

import (
	"context"

	"fmt"

	"github.com/domurdoc/gophermart/internal/models"
	"github.com/domurdoc/gophermart/internal/repositories"
)

type InvalidTokenError struct {
	Err error
}

func (e *InvalidTokenError) Error() string {
	return fmt.Sprintf("invalid token: %v", e.Err)

}

func (e *InvalidTokenError) Unwrap() error {
	return e.Err
}

type Strategy interface {
	WriteToken(context.Context, *models.User) (string, error)
	ReadToken(context.Context, string, repositories.UserRepository) (*models.User, error)
}
