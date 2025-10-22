package app

import (
	"go.uber.org/zap"

	"github.com/domurdoc/gophermart/internal/auth"
	"github.com/domurdoc/gophermart/internal/auth/strategy"
	"github.com/domurdoc/gophermart/internal/auth/transport"
	"github.com/domurdoc/gophermart/internal/client"
	"github.com/domurdoc/gophermart/internal/config"
	"github.com/domurdoc/gophermart/internal/logger"
	"github.com/domurdoc/gophermart/internal/services"
	"github.com/domurdoc/gophermart/internal/services/bonus"
	"github.com/domurdoc/gophermart/internal/utils"
)

type Services struct {
	Log     *zap.SugaredLogger
	Auth    *auth.Auth
	Client  client.BonusClient
	Bonus   *bonus.BonusService
	Order   *services.OrderService
	Balance *services.BalanceService

	options *config.ServicesOptions
	repos   *Repositories
	closer  *utils.Closer
}

func NewServices(options *config.ServicesOptions, repos *Repositories) (*Services, error) {
	s := Services{options: options, repos: repos, closer: utils.NewCloser()}

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
	return s.closer.Close()
}

func (s *Services) initLog() error {
	log, err := logger.New(s.options.LogLevel)
	if err != nil {
		return err
	}
	s.Log = log
	s.closer.Register(log.Sync)
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
	s.closer.Register(s.Client.Close)
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
	params := bonus.BonusParams{
		BonusClient:            s.Client,
		BalanceRepo:            s.repos.Balance,
		SaverBatchMaxSize:      s.options.SaverBatchMaxSize,
		SaverBatchInterval:     s.options.SaverBatchInterval,
		RefresherBatchMaxSize:  s.options.RefresherBatchMaxSize,
		RefresherBatchInterval: s.options.RefresherBatchInterval,
		CheckerPoolSize:        s.options.CheckerPoolSize,
		SaverPoolSize:          s.options.SaverPoolSize,
		Log:                    s.Log,
	}
	bonusService, err := bonus.NewBonusService(&params)
	if err != nil {
		return err
	}
	s.Bonus = bonusService
	s.closer.Register(s.Bonus.Close)
	return nil
}
