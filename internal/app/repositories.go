package app

import (
	"database/sql"

	"github.com/domurdoc/gophermart/internal/config"
	"github.com/domurdoc/gophermart/internal/repositories"
	"github.com/domurdoc/gophermart/internal/repositories/pg"
	"github.com/domurdoc/gophermart/internal/utils"
	"github.com/domurdoc/gophermart/migrations"
)

type Repositories struct {
	User    repositories.UserRepository
	Order   repositories.OrderRepository
	Balance repositories.BalanceRepository

	options *config.RepositoriesOptions
	db      *sql.DB
	closer  *utils.Closer
}

func NewRepositories(options *config.RepositoriesOptions) (*Repositories, error) {
	r := Repositories{options: options, closer: utils.NewCloser()}
	if err := r.initDB(); err != nil {
		return nil, err
	}
	if err := r.initUserRepository(); err != nil {
		r.Close()
		return nil, err
	}
	if err := r.initBalanceRepository(); err != nil {
		r.Close()
		return nil, err
	}
	if err := r.initOrderRepository(); err != nil {
		r.Close()
		return nil, err
	}
	return &r, nil
}

func (r *Repositories) Close() error {
	return r.closer.Close()
}

func (r *Repositories) initDB() error {
	db, err := config.NewPostgresDB(r.options.DatabaseURI)
	if err != nil {
		return err
	}
	if err := migrations.Migrate(db); err != nil {
		return err
	}
	r.db = db
	r.closer.Register(r.db.Close)
	return nil
}

func (r *Repositories) initUserRepository() error {
	r.User = pg.NewUserRepository(r.db)
	return nil
}

func (r *Repositories) initBalanceRepository() error {
	r.Balance = pg.NewBalanceRepository(r.db)
	return nil
}

func (r *Repositories) initOrderRepository() error {
	r.Order = pg.NewOrderRepository(r.db)
	return nil
}
