package services

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/domurdoc/gophermart/internal/client"
	"github.com/domurdoc/gophermart/internal/models"
	"github.com/domurdoc/gophermart/internal/repositories"
)

type BonusService struct {
	bonusClient               client.BonusClient
	balanceRepo               repositories.BalanceRepository
	checkOrderCh              chan *models.Order
	checkedOrderCh            chan *models.Order
	checkedOrderBatchCh       chan []*models.Order
	checkedOrderBatchMaxSize  int
	checkedOrderBatchInterval time.Duration
	userCh                    chan int
	userBatchCh               chan []int
	userBatchMaxSize          int
	userBatchInterval         time.Duration
	checkWorkers              int
	saveWorkers               int
	doneCh                    chan struct{}
	log                       *zap.SugaredLogger
}

type BonusParams struct {
	BonusClient               client.BonusClient
	BalanceRepo               repositories.BalanceRepository
	CheckedOrderBatchMaxSize  int
	CheckedOrderBatchInterval time.Duration
	UserBatchMaxSize          int
	UserBatchInterval         time.Duration
	CheckWorkers              int
	SaveWorkers               int
	Log                       *zap.SugaredLogger
}

func NewBonusService(params *BonusParams) (*BonusService, error) {
	s := BonusService{
		bonusClient:               params.BonusClient,
		balanceRepo:               params.BalanceRepo,
		checkedOrderBatchMaxSize:  params.CheckedOrderBatchMaxSize,
		checkedOrderBatchInterval: params.CheckedOrderBatchInterval,
		userBatchMaxSize:          params.UserBatchMaxSize,
		userBatchInterval:         params.UserBatchInterval,
		checkWorkers:              params.CheckWorkers,
		saveWorkers:               params.SaveWorkers,
		log:                       params.Log,
		checkOrderCh:              make(chan *models.Order),
		checkedOrderCh:            make(chan *models.Order),
		checkedOrderBatchCh:       make(chan []*models.Order),
		userCh:                    make(chan int),
		userBatchCh:               make(chan []int),
		doneCh:                    make(chan struct{}),
	}
	for range s.checkWorkers {
		go s.checkOrderWorker()
	}
	go s.batchCheckedOrders()
	for range s.saveWorkers {
		go s.saveCheckedOrderBatchWorker()
	}
	go s.batchUsersToRefresh()
	go s.refreshBalanceWorker()

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
	return nil
}

func (s *BonusService) CheckOrder(order *models.Order) {
	select {
	case <-s.doneCh:
		return
	case s.checkOrderCh <- order:
	}
}

func (s *BonusService) checkOrders(orders []*models.Order) {
	for _, order := range orders {
		select {
		case <-s.doneCh:
			return
		case s.checkOrderCh <- order:
		}
	}
}

func (s *BonusService) refreshUsers(userIDS []int) {
	for _, userID := range userIDS {
		select {
		case <-s.doneCh:
			return
		case s.userCh <- userID:
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

func (s *BonusService) checkOrderWorker() {
	ctx := context.Background()
	for {
		select {
		case <-s.doneCh:
			return
		case order := <-s.checkOrderCh:
			s.log.Debugw("received an order to check", "order", order.Number)
			bonus, err := s.bonusClient.GetForOrder(ctx, order.Number)
			if err != nil {
				s.log.Errorw("failed to check an order", "order", order.Number, "error", err.Error())
				continue
			}
			order.Accrual = bonus.Accrual
			order.Status = bonus.Status
			select {
			case <-s.doneCh:
				return
			case s.checkedOrderCh <- order:
			}
		}
	}
}

func (s *BonusService) batchCheckedOrders() {
	var batch []*models.Order

	t := time.NewTicker(s.checkedOrderBatchInterval)
	defer t.Stop()

	for {
		select {
		case <-s.doneCh:
			return
		case order := <-s.checkedOrderCh:
			s.log.Debugw("received an order to save", "order", order.Number)
			batch = append(batch, order)
			if len(batch) >= s.checkedOrderBatchMaxSize {
				select {
				case <-s.doneCh:
					return
				case s.checkedOrderBatchCh <- batch:
					batch = nil
				}
				t.Reset(s.checkedOrderBatchInterval)
			}
		case <-t.C:
			if len(batch) > 0 {
				select {
				case <-s.doneCh:
					return
				case s.checkedOrderBatchCh <- batch:
					batch = nil
				}
			}
		}
	}
}

func (s *BonusService) saveCheckedOrderBatchWorker() {
	ctx := context.Background()
	for {
		select {
		case <-s.doneCh:
			return
		case orderBatch := <-s.checkedOrderBatchCh:
			userIDS, err := s.balanceRepo.AccrueBonuses(ctx, orderBatch)
			if err != nil {
				s.log.Errorw("failed to accrue bonuses", "error", err.Error())
				continue
			}
			s.log.Debugw("bonuses accrued", "user_ids", userIDS)
			for _, userID := range userIDS {
				select {
				case <-s.doneCh:
					return
				case s.userCh <- userID:
				}
			}
		}
	}
}

func (s *BonusService) batchUsersToRefresh() {
	var batch []int

	t := time.NewTicker(s.userBatchInterval)
	defer t.Stop()

	for {
		select {
		case <-s.doneCh:
			return
		case userID := <-s.userCh:
			s.log.Debugw("received a user to refresh", "user_id", userID)
			batch = append(batch, userID)
			if len(batch) >= s.userBatchMaxSize {
				select {
				case <-s.doneCh:
					return
				case s.userBatchCh <- batch:
					batch = nil
				}
				t.Reset(s.userBatchInterval)
			}
		case <-t.C:
			if len(batch) > 0 {
				select {
				case <-s.doneCh:
					return
				case s.userBatchCh <- batch:
					batch = nil
				}
			}
		}
	}
}

func (s *BonusService) refreshBalanceWorker() {
	ctx := context.Background()
	for {
		select {
		case <-s.doneCh:
			return
		case userIDBatch := <-s.userBatchCh:
			err := s.balanceRepo.RefreshUsers(ctx, userIDBatch)
			if err != nil {
				s.log.Errorw("failed to refresh users", "error", err.Error())
			}
			s.log.Debugw("users refreshed", "user_ids", userIDBatch)
		}
	}
}
