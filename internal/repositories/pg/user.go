package pg

import (
	"context"
	"database/sql"
	"errors"

	"github.com/domurdoc/gophermart/internal/models"
)

type PGUserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *PGUserRepository {
	return &PGUserRepository{db: db}
}

func (r *PGUserRepository) CreateUser(ctx context.Context, username string, passwordHash string) (*models.User, error) {
	var user models.User

	row := r.db.QueryRowContext(
		ctx,
		`
INSERT INTO users (username, password_hash)
VALUES ($1, $2)
ON CONFLICT DO NOTHING
RETURNING id, username, password_hash
`,
		username,
		passwordHash,
	)
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrUsernameExists
		}
		return nil, err
	}
	return &user, nil
}

func (r *PGUserRepository) GetUserByID(ctx context.Context, userID int) (*models.User, error) {
	return r.getUser(
		ctx,
		`
SELECT id, username, password_hash FROM users WHERE id = $1
`,
		userID,
	)
}

func (r *PGUserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	return r.getUser(
		ctx,
		`
SELECT id, username, password_hash FROM users WHERE LOWER(username) = LOWER($1)
`,
		username,
	)
}

func (r *PGUserRepository) getUser(ctx context.Context, query string, uniqueValue any) (*models.User, error) {
	var user models.User

	row := r.db.QueryRowContext(ctx, query, uniqueValue)
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}
