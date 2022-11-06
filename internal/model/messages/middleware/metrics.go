package middleware

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/clients/tg"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
	"golang.org/x/net/context"
)

type Metrics struct {
	next tg.Message
}

func NewMetrics(next tg.Message) *Metrics {
	return &Metrics{next: next}
}

var (
	msgGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "bot",
		Subsystem: "tg_client",
		Name:      "incoming_messages_total",
		Help:      "total number of messages from tg users",
	})
	histogramResponseTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "bot",
			Subsystem: "tg_client",
			Name:      "histogram_execute_time_seconds",
			Help:      "execute command time",
			Buckets:   prometheus.ExponentialBuckets(0.01, 2, 10),
		},
		[]string{"command"},
	)
	summaryResponseTime = promauto.NewSummary(prometheus.SummaryOpts{
		Namespace: "bot",
		Subsystem: "tg_client",
		Name:      "summary_execute_time_seconds",
		Help:      "execute command time",
		Objectives: map[float64]float64{
			0.5:  0.1,
			0.9:  0.01,
			0.99: 0.01,
		},
	})
)

func (m *Metrics) IncomingMessage(ctx context.Context, msg *dto.Message) error {
	msgGauge.Inc()
	startTime := time.Now()
	err := m.next.IncomingMessage(ctx, msg)
	duration := time.Since(startTime)
	summaryResponseTime.Observe(duration.Seconds())
	histogramResponseTime.WithLabelValues(msg.Command).Observe(duration.Seconds())
	if err != nil {
		return err
	}
	return nil
}
