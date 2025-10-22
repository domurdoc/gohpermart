package bonus

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/domurdoc/gophermart/internal/models"
	"github.com/domurdoc/gophermart/internal/repositories"
)

type SaverPool struct {
	log                *zap.SugaredLogger
	orderToSaveCh      chan *models.Order
	userToRefreshCh    chan int
	poolSize           int
	maxBatchSize       int
	batchInterval      time.Duration
	balanceRepo        repositories.BalanceRepository
	orderBatchToSaveCh chan []*models.Order
	wg                 sync.WaitGroup
	doneCh             chan struct{}
}

func NewSaverPool(
	log *zap.SugaredLogger,
	orderToSaveCh chan *models.Order,
	userToRefreshCh chan int,
	poolSize int,
	maxBatchSize int,
	batchInterval time.Duration,
	balanceRepo repositories.BalanceRepository,
) *SaverPool {
	p := SaverPool{
		log:                log,
		orderToSaveCh:      orderToSaveCh,
		userToRefreshCh:    userToRefreshCh,
		poolSize:           poolSize,
		maxBatchSize:       maxBatchSize,
		batchInterval:      batchInterval,
		balanceRepo:        balanceRepo,
		orderBatchToSaveCh: make(chan []*models.Order),
		wg:                 sync.WaitGroup{},
		doneCh:             make(chan struct{}),
	}
	p.wg.Add(1)
	go p.batcher()
	for i := range p.poolSize {
		p.wg.Add(1)
		go p.worker(i)
	}
	return &p
}

func (p *SaverPool) Close() error {
	close(p.doneCh)
	p.wg.Wait()
	return nil
}

func (p *SaverPool) batcher() {
	defer p.wg.Done()
	var batch []*models.Order

	t := time.NewTicker(p.batchInterval)
	defer t.Stop()

	for {
		select {
		case <-p.doneCh:
			return
		case order := <-p.orderToSaveCh:
			p.log.Debugw("received an order to save", "order", order.Number)
			batch = append(batch, order)
			if len(batch) >= p.maxBatchSize {
				select {
				case <-p.doneCh:
					return
				case p.orderBatchToSaveCh <- batch:
					batch = nil
				}
				t.Reset(p.batchInterval)
			}
		case <-t.C:
			if len(batch) > 0 {
				select {
				case <-p.doneCh:
					return
				case p.orderBatchToSaveCh <- batch:
					batch = nil
				}
			}
		}
	}
}

func (p *SaverPool) worker(workerID int) {
	defer p.wg.Done()
	ctx := context.Background()

	for {
		select {
		case <-p.doneCh:
			return
		case orderBatch := <-p.orderBatchToSaveCh:
			userIDS, err := p.balanceRepo.AccrueBonuses(ctx, orderBatch)
			if err != nil {
				p.log.Errorw("failed to accrue bonuses", "worker_id", workerID, "error", err.Error())
				continue
			}
			p.log.Debugw("bonuses accrued", "worker_id", workerID, "user_ids", userIDS)
			for _, userID := range userIDS {
				select {
				case <-p.doneCh:
					return
				case p.userToRefreshCh <- userID:
				}
			}
		}
	}
}
