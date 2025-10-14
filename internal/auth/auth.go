package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/domurdoc/gophermart/internal/auth/strategy"
	"github.com/domurdoc/gophermart/internal/auth/transport"
	"github.com/domurdoc/gophermart/internal/models"
	"github.com/domurdoc/gophermart/internal/repositories"
	"github.com/domurdoc/gophermart/internal/utils"
)

type Auth struct {
	strategy  strategy.Strategy
	transport transport.Transport
	userRepo  repositories.UserRepository
}

func New(strategy strategy.Strategy, transport transport.Transport, userRepo repositories.UserRepository) *Auth {
	return &Auth{
		strategy:  strategy,
		transport: transport,
		userRepo:  userRepo,
	}
}

func (a *Auth) AuthenticateToken(ctx context.Context, r *http.Request) (*models.User, error) {
	tokenString, err := a.transport.Read(r)
	if err != nil {
		return nil, err
	}
	u, err := a.strategy.ReadToken(ctx, tokenString, a.userRepo)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (a *Auth) AuthenticateCredentials(ctx context.Context, username string, password string) (*models.User, error) {
	user, err := a.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return nil, models.ErrInvalidCredentials
		}
		return nil, err
	}
	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		return nil, models.ErrInvalidCredentials
	}
	return user, nil
}

func (a *Auth) Login(ctx context.Context, w http.ResponseWriter, u *models.User) error {
	tokenString, err := a.strategy.WriteToken(ctx, u)
	if err != nil {
		return err
	}
	return a.transport.Write(w, tokenString)
}

func (a *Auth) Register(ctx context.Context, username string, password string) (*models.User, error) {
	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}
	return a.userRepo.CreateUser(ctx, username, passwordHash)
}
