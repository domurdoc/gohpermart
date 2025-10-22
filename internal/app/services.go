package app

import (
	"errors"

	"go.uber.org/zap"

	"github.com/domurdoc/gophermart/internal/auth"
	"github.com/domurdoc/gophermart/internal/auth/strategy"
	"github.com/domurdoc/gophermart/internal/auth/transport"
	"github.com/domurdoc/gophermart/internal/client"
	"github.com/domurdoc/gophermart/internal/config"
	"github.com/domurdoc/gophermart/internal/logger"
	"github.com/domurdoc/gophermart/internal/services"
)

type Services struct {
	Log     *zap.SugaredLogger
	Auth    *auth.Auth
	Client  client.BonusClient
	Bonus   *services.BonusService
	Order   *services.OrderService
	Balance *services.BalanceService

	options *config.ServicesOptions
	repos   *Repositories
}

func NewServices(options *config.ServicesOptions, repos *Repositories) (*Services, error) {
	s := Services{options: options, repos: repos}

	if err := s.initLog(); err != nil {
		s.Close()
		return nil, err
	}
	if err := s.initAuth(); err != nil {
		s.Close()
		return nil, err
	}
	if err := s.initBalanceService(); err != nil {
		s.Close()
		return nil, err
	}
	if err := s.initBonusClient(); err != nil {
		s.Close()
		return nil, err
	}
	if err := s.initBonusService(); err != nil {
		s.Close()
		return nil, err
	}
	if err := s.initOrderService(); err != nil {
		s.Close()
		return nil, err
	}
	return &s, nil
}

func (s *Services) Close() error {
	var errs []error

	if s.Bonus != nil {
		errs = append(errs, s.Bonus.Close())
	}
	if s.Client != nil {
		errs = append(errs, s.Client.Close())
	}
	if s.Log != nil {
		errs = append(errs, s.Log.Sync())
	}

	return errors.Join(errs...)
}

func (s *Services) initLog() error {
	log, err := logger.New(s.options.LogLevel)
	if err != nil {
		return err
	}
	s.Log = log
	return nil
}

func (s *Services) initAuth() error {
	strategy := strategy.NewJWT(
		s.options.JWTSecret,
		s.options.JWTDuration,
	)
	transport := transport.NewCookie(
		s.options.CookieName,
		int(s.options.CookieMaxAge.Seconds()),
		false,
	)
	s.Auth = auth.New(strategy, transport, s.repos.User)
	return nil
}

func (s *Services) initBonusClient() error {
	if s.options.DebugClient {
		s.Client = &client.DebugBonusClient{}
		return nil
	}
	params := client.HTTPBonusParams{
		URL:              s.options.AccrualSystemAddress,
		PoolSize:         s.options.PoolSize,
		PoolTimeout:      s.options.PoolTimeout,
		MaxReries:        s.options.MaxReries,
		RetryWaitTime:    s.options.RetryWaitTime,
		RetryMaxWaitTime: s.options.RetryMaxWaitTime,
	}
	s.Client = client.NewBonusClient(&params)
	return nil
}

func (s *Services) initBalanceService() error {
	s.Balance = services.NewBalanceService(s.repos.Balance)
	return nil
}

func (s *Services) initOrderService() error {
	s.Order = services.NewOrderService(s.repos.Order, s.Bonus)
	return nil
}

func (s *Services) initBonusService() error {
	params := services.BonusParams{
		BonusClient:               s.Client,
		BalanceRepo:               s.repos.Balance,
		CheckedOrderBatchMaxSize:  s.options.CheckedOrderBatchMaxSize,
		CheckedOrderBatchInterval: s.options.CheckedOrderBatchInterval,
		UserBatchMaxSize:          s.options.UserBatchMaxSize,
		UserBatchInterval:         s.options.UserBatchInterval,
		CheckWorkers:              s.options.CheckWorkers,
		SaveWorkers:               s.options.SaveWorkers,
		Log:                       s.Log,
	}
	bonusService, err := services.NewBonusService(&params)
	if err != nil {
		return err
	}
	s.Bonus = bonusService
	return nil
}
