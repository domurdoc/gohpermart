package bonus

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/domurdoc/gophermart/internal/client"
	"github.com/domurdoc/gophermart/internal/models"
)

type CheckerPool struct {
	log            *zap.SugaredLogger
	client         client.BonusClient
	orderToCheckCh chan *models.Order
	orderToSave    chan *models.Order
	poolSize       int
	wg             sync.WaitGroup
	doneCh         chan struct{}
}

func NewCheckerPool(
	log *zap.SugaredLogger,
	client client.BonusClient,
	orderToCheckCh chan *models.Order,
	orderToSaveCh chan *models.Order,
	poolSize int,
) *CheckerPool {
	pool := CheckerPool{
		log:            log,
		client:         client,
		orderToCheckCh: orderToCheckCh,
		orderToSave:    orderToSaveCh,
		poolSize:       poolSize,
		wg:             sync.WaitGroup{},
		doneCh:         make(chan struct{}),
	}
	for i := range pool.poolSize {
		pool.wg.Add(1)
		go pool.worker(i)
	}
	return &pool
}

func (p *CheckerPool) Close() error {
	close(p.doneCh)
	p.wg.Wait()
	return nil
}

func (p *CheckerPool) worker(workerID int) {
	defer p.wg.Done()
	ctx := context.Background()

	for {
		select {
		case <-p.doneCh:
			return
		case order := <-p.orderToCheckCh:
			p.log.Debugw("received an order to check", "worker_id", workerID, "order", order.Number)
			bonus, err := p.client.GetForOrder(ctx, order.Number)
			if err != nil {
				p.log.Errorw("failed to check an order", "worker_id", workerID, "order", order.Number, "error", err.Error())
				continue
			}
			order.Accrual = bonus.Accrual
			order.Status = bonus.Status
			select {
			case <-p.doneCh:
				return
			case p.orderToSave <- order:
			}
		}
	}
}
