package bonus

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/domurdoc/gophermart/internal/client"
	"github.com/domurdoc/gophermart/internal/models"
	"github.com/domurdoc/gophermart/internal/repositories"
	"github.com/domurdoc/gophermart/internal/utils"
)

type BonusParams struct {
	Log                    *zap.SugaredLogger
	BonusClient            client.BonusClient
	BalanceRepo            repositories.BalanceRepository
	CheckerPoolSize        int
	SaverPoolSize          int
	SaverBatchMaxSize      int
	SaverBatchInterval     time.Duration
	RefresherBatchMaxSize  int
	RefresherBatchInterval time.Duration
}

type BonusService struct {
	saverPool       *SaverPool
	checkerPool     *CheckerPool
	refresherPool   *RefresherPool
	balanceRepo     repositories.BalanceRepository
	log             *zap.SugaredLogger
	orderToCheckCh  chan *models.Order
	orderToSaveCh   chan *models.Order
	userToRefreshCh chan int
	doneCh          chan struct{}
	closer          *utils.Closer
}

func NewBonusService(params *BonusParams) (*BonusService, error) {
	s := BonusService{
		closer:          utils.NewCloser(),
		balanceRepo:     params.BalanceRepo,
		log:             params.Log,
		orderToCheckCh:  make(chan *models.Order),
		orderToSaveCh:   make(chan *models.Order),
		userToRefreshCh: make(chan int),
		doneCh:          make(chan struct{}),
	}

	s.checkerPool = NewCheckerPool(
		params.Log,
		params.BonusClient,
		s.orderToCheckCh,
		s.orderToSaveCh,
		params.CheckerPoolSize,
	)
	s.closer.Register(s.checkerPool.Close)

	s.saverPool = NewSaverPool(
		params.Log,
		s.orderToSaveCh,
		s.userToRefreshCh,
		params.SaverPoolSize,
		params.SaverBatchMaxSize,
		params.SaverBatchInterval,
		params.BalanceRepo,
	)
	s.closer.Register(s.saverPool.Close)

	s.refresherPool = NewRefresherPool(
		params.Log,
		s.userToRefreshCh,
		params.RefresherBatchMaxSize,
		params.RefresherBatchInterval,
		params.BalanceRepo,
	)
	s.closer.Register(s.refresherPool.Close)

	if err := s.loadOrdersToCheck(); err != nil {
		s.Close()
		return nil, err
	}
	if err := s.loadUsersToRefresh(); err != nil {
		s.Close()
		return nil, err
	}
	return &s, nil
}

func (s *BonusService) Close() error {
	close(s.doneCh)
	return s.closer.Close()
}

func (s *BonusService) CheckOrder(order *models.Order) {
	select {
	case <-s.doneCh:
		return
	case s.orderToCheckCh <- order:
	}
}

func (s *BonusService) checkOrders(orders []*models.Order) {
	for _, order := range orders {
		select {
		case <-s.doneCh:
			return
		case s.orderToCheckCh <- order:
		}
	}
}

func (s *BonusService) refreshUsers(userIDS []int) {
	for _, userID := range userIDS {
		select {
		case <-s.doneCh:
			return
		case s.userToRefreshCh <- userID:
		}
	}
}

func (s *BonusService) loadOrdersToCheck() error {
	orders, err := s.balanceRepo.GetOrdersToCheck(context.TODO())
	if err != nil {
		return err
	}
	s.log.Infow("loading orders to check", "count", len(orders))
	s.checkOrders(orders)
	return nil
}

func (s *BonusService) loadUsersToRefresh() error {
	userIDS, err := s.balanceRepo.GetUsersToRefresh(context.TODO())
	if err != nil {
		return err
	}
	s.log.Infow("loading users to refresh", "count", len(userIDS))
	s.refreshUsers(userIDS)
	return nil
}
