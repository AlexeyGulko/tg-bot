package update_rates

import (
	"sync"
	"time"

	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/logger"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type Worker struct {
	service CurrencyService
	config  Config
	ch      chan ChannelR
}

type Channel chan ChannelR

type ChannelR struct {
	T  time.Time
	Wg *sync.WaitGroup
}

type CurrencyService interface {
	UpdateRates(ctx context.Context, date time.Time) error
}

type Config interface {
	CurrencyUpdateDuration() time.Duration
}

func New(service CurrencyService, config Config, ch chan ChannelR) *Worker {
	return &Worker{service: service, config: config, ch: ch}
}

func (w *Worker) Run(ctx context.Context) {
	var wg sync.WaitGroup
	req := ChannelR{T: time.Now(), Wg: &wg}
	req.Wg.Add(1)
	go func() { w.ch <- req }()
	ticker := time.NewTicker(w.config.CurrencyUpdateDuration())

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				close(w.ch)
				logger.Info("update rates stopped")
				return
			case <-ticker.C:
				var wg sync.WaitGroup
				req := ChannelR{T: time.Now(), Wg: &wg}
				req.Wg.Add(1)
				go func() { w.ch <- req }()
			case req := <-w.ch:
				err := w.service.UpdateRates(ctx, req.T)
				if err != nil {
					logger.Error("update rates error", zap.Error(err))
				}
				req.Wg.Done()
			}
		}
	}()
}
