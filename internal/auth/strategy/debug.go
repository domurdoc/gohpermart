package strategy

import (
	"context"
	"strconv"

	"github.com/domurdoc/gophermart/internal/models"
	"github.com/domurdoc/gophermart/internal/repositories"
)

type DebugStrategy struct{}

func NewDebug() *DebugStrategy {
	return &DebugStrategy{}
}

func (s *DebugStrategy) WriteToken(ctx context.Context, user *models.User) (string, error) {
	return strconv.Itoa(user.ID), nil
}

func (s *DebugStrategy) ReadToken(ctx context.Context, token string, repo repositories.UserRepository) (*models.User, error) {
	userID, err := strconv.Atoi(token)
	if err != nil {
		return nil, err
	}
	return repo.GetUserByID(ctx, userID)
}
