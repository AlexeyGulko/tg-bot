package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/logger"
	"go.uber.org/zap"
)

type Worker struct {
	cfg Config
}

type Config interface {
	Port() int64
}

func New(cfg Config) *Worker {
	return &Worker{cfg: cfg}
}

func (w *Worker) Run(ctx context.Context) {
	http.Handle("/metrics", promhttp.Handler())

	go func() {
		logger.Info("start serve http")
		for {
			select {
			case <-ctx.Done():
				logger.Info("listen http stopped")
				return
			default:
				err := http.ListenAndServe(fmt.Sprintf(":%d", w.cfg.Port()), nil)
				if err != nil {
					logger.Fatal("error starting http server", zap.Error(err))
					return
				}
			}
		}
	}()
}
