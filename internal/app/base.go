package app

import (
	"database/sql"
	"errors"

	"go.uber.org/zap"

	"github.com/domurdoc/gophermart/internal/auth"
	"github.com/domurdoc/gophermart/internal/auth/strategy"
	"github.com/domurdoc/gophermart/internal/auth/transport"
	"github.com/domurdoc/gophermart/internal/client"
	"github.com/domurdoc/gophermart/internal/config"
	"github.com/domurdoc/gophermart/internal/logger"
	"github.com/domurdoc/gophermart/internal/repositories"
	"github.com/domurdoc/gophermart/internal/repositories/pg"
	"github.com/domurdoc/gophermart/internal/services"
	"github.com/domurdoc/gophermart/migrations"
)

type App struct {
	Options *config.Options

	Log *zap.SugaredLogger

	DB *sql.DB

	UserRepo    repositories.UserRepository
	OrderRepo   repositories.OrderRepository
	BalanceRepo repositories.BalanceRepository

	BonusClient client.BonusClient

	Auth           *auth.Auth
	BonusService   *services.BonusService
	OrderService   *services.OrderService
	BalanceService *services.BalanceService
}

func New() (*App, error) {
	a := &App{Options: config.New()}

	if err := a.initLog(); err != nil {
		a.Close()
		return nil, err
	}
	if err := a.initDB(); err != nil {
		a.Close()
		return nil, err
	}
	if err := a.initUserRepository(); err != nil {
		a.Close()
		return nil, err
	}
	if err := a.initBalanceRepository(); err != nil {
		a.Close()
		return nil, err
	}
	if err := a.initOrderRepository(); err != nil {
		a.Close()
		return nil, err
	}
	if err := a.initAuth(); err != nil {
		a.Close()
		return nil, err
	}
	if err := a.initBalanceService(); err != nil {
		a.Close()
		return nil, err
	}
	if err := a.initBonusClient(); err != nil {
		a.Close()
		return nil, err
	}
	if err := a.initBonusService(); err != nil {
		a.Close()
		return nil, err
	}
	if err := a.initOrderService(); err != nil {
		a.Close()
		return nil, err
	}
	return a, nil
}

func (a *App) Close() error {
	var errs []error

	if a.BonusService != nil {
		errs = append(errs, a.BonusService.Close())
	}
	if a.BonusClient != nil {
		errs = append(errs, a.BonusClient.Close())
	}
	if a.DB != nil {
		errs = append(errs, a.DB.Close())
	}
	if a.Log != nil {
		errs = append(errs, a.Log.Sync())
	}

	return errors.Join(errs...)
}

func (a *App) initDB() error {
	db, err := config.NewPostgresDB(a.Options.DatabaseURI)
	if err != nil {
		return err
	}
	if err := migrations.Migrate(db); err != nil {
		return err
	}
	a.DB = db
	return nil
}

func (a *App) initUserRepository() error {
	a.UserRepo = pg.NewUserRepository(a.DB)
	return nil
}

func (a *App) initBalanceRepository() error {
	a.BalanceRepo = pg.NewBalanceRepository(a.DB)
	return nil
}

func (a *App) initOrderRepository() error {
	a.OrderRepo = pg.NewOrderRepository(a.DB)
	return nil
}

func (a *App) initLog() error {
	log, err := logger.New(a.Options.LogLevel)
	if err != nil {
		return err
	}
	a.Log = log
	return nil
}

func (a *App) initAuth() error {
	strategy := strategy.NewJWT(
		a.Options.JWTSecret,
		a.Options.JWTDuration,
	)
	transport := transport.NewCookie(
		a.Options.CookieName,
		int(a.Options.CookieMaxAge.Seconds()),
		false,
	)
	a.Auth = auth.New(strategy, transport, a.UserRepo)
	return nil
}

func (a *App) initBalanceService() error {
	a.BalanceService = services.NewBalanceService(a.BalanceRepo)
	return nil
}

func (a *App) initOrderService() error {
	a.OrderService = services.NewOrderService(a.OrderRepo, a.BonusService)
	return nil
}

func (a *App) initBonusClient() error {
	if a.Options.DebugClient {
		a.BonusClient = &client.DebugBonusClient{}
		return nil
	}
	params := client.HTTPBonusParams{
		URL:              a.Options.AccrualSystemAddress,
		PoolSize:         a.Options.PoolSize,
		PoolTimeout:      a.Options.PoolTimeout,
		MaxReries:        a.Options.MaxReries,
		RetryWaitTime:    a.Options.RetryWaitTime,
		RetryMaxWaitTime: a.Options.RetryMaxWaitTime,
	}
	a.BonusClient = client.NewBonusClient(&params)
	return nil
}

func (a *App) initBonusService() error {
	params := services.BonusParams{
		BonusClient:               a.BonusClient,
		BalanceRepo:               a.BalanceRepo,
		CheckedOrderBatchMaxSize:  a.Options.CheckedOrderBatchMaxSize,
		CheckedOrderBatchInterval: a.Options.CheckedOrderBatchInterval,
		UserBatchMaxSize:          a.Options.UserBatchMaxSize,
		UserBatchInterval:         a.Options.UserBatchInterval,
		CheckWorkers:              a.Options.CheckWorkers,
		SaveWorkers:               a.Options.SaveWorkers,
		Log:                       a.Log,
	}
	bonusService, err := services.NewBonusService(&params)
	if err != nil {
		return err
	}
	a.BonusService = bonusService
	return nil
}
