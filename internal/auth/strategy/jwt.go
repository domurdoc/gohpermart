package strategy

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/domurdoc/gophermart/internal/models"
	"github.com/domurdoc/gophermart/internal/repositories"
)

type JWTStrategy struct {
	secretKey string
	tokenExp  time.Duration
}

type claims struct {
	jwt.RegisteredClaims
	UserID int
}

func NewJWT(secretKey string, tokenExp time.Duration) *JWTStrategy {
	return &JWTStrategy{secretKey: secretKey, tokenExp: tokenExp}
}

func (s *JWTStrategy) WriteToken(ctx context.Context, user *models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenExp)),
		},
		UserID: user.ID,
	})
	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (s *JWTStrategy) ReadToken(ctx context.Context, tokenString string, repo repositories.UserRepository) (*models.User, error) {
	claims := &claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (any, error) {
			return []byte(s.secretKey), nil
		})
	if err != nil {
		return nil, &InvalidTokenError{err}
	}
	if !token.Valid {
		return nil, &InvalidTokenError{err}
	}
	return repo.GetUserByID(ctx, claims.UserID)
}
