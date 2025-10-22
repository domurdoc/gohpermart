package bonus

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/domurdoc/gophermart/internal/repositories"
)

type RefresherPool struct {
	log                  *zap.SugaredLogger
	userToRefreshCh      chan int
	maxBatchSize         int
	batchInterval        time.Duration
	balanceRepo          repositories.BalanceRepository
	userBatchToRefreshCh chan []int
	wg                   sync.WaitGroup
	doneCh               chan struct{}
}

func NewRefresherPool(
	log *zap.SugaredLogger,
	userToRefreshCh chan int,
	maxBatchSize int,
	batchInterval time.Duration,
	balanceRepo repositories.BalanceRepository,
) *RefresherPool {
	p := RefresherPool{
		log:                  log,
		userToRefreshCh:      userToRefreshCh,
		maxBatchSize:         maxBatchSize,
		batchInterval:        batchInterval,
		balanceRepo:          balanceRepo,
		userBatchToRefreshCh: make(chan []int),
		wg:                   sync.WaitGroup{},
		doneCh:               make(chan struct{}),
	}
	p.wg.Add(1)
	go p.batcher()
	p.wg.Add(1)
	go p.worker()
	return &p
}

func (p *RefresherPool) Close() error {
	close(p.doneCh)
	p.wg.Wait()
	return nil
}

func (p *RefresherPool) batcher() {
	defer p.wg.Done()
	var batch []int

	t := time.NewTicker(p.batchInterval)
	defer t.Stop()

	for {
		select {
		case <-p.doneCh:
			return
		case userID := <-p.userToRefreshCh:
			p.log.Debugw("received a user to refresh", "user_id", userID)
			batch = append(batch, userID)
			if len(batch) >= p.maxBatchSize {
				select {
				case <-p.doneCh:
					return
				case p.userBatchToRefreshCh <- batch:
					batch = nil
				}
				t.Reset(p.batchInterval)
			}
		case <-t.C:
			if len(batch) > 0 {
				select {
				case <-p.doneCh:
					return
				case p.userBatchToRefreshCh <- batch:
					batch = nil
				}
			}
		}
	}
}

func (p *RefresherPool) worker() {
	defer p.wg.Done()
	ctx := context.Background()
	for {
		select {
		case <-p.doneCh:
			return
		case userIDBatch := <-p.userBatchToRefreshCh:
			err := p.balanceRepo.RefreshUsers(ctx, userIDBatch)
			if err != nil {
				p.log.Errorw("failed to refresh users", "error", err.Error())
			}
			p.log.Debugw("users refreshed", "user_ids", userIDBatch)
		}
	}
}
