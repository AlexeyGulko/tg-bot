package Queue

import (
	"github.com/Shopify/sarama"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/logger"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

var successGauge = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: "bot",
	Subsystem: "tg_bot",
	Name:      "success_msg_total",
	Help:      "queue success messages",
})

type Worker struct {
	producer sarama.AsyncProducer
}

func New(brokerList []string) (*Worker, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0
	// So we can know the partition and offset of messages.
	config.Producer.Return.Successes = true

	producer, err := sarama.NewAsyncProducer(brokerList, config)
	if err != nil {
		return nil, err
	}

	return &Worker{producer: producer}, nil
}

func (w Worker) MessageChannel() chan<- *sarama.ProducerMessage {
	return w.producer.Input()
}

func (w Worker) Run(ctx context.Context) {
	// We will log to STDOUT if we're not able to produce messages.
	go func() {
		for err := range w.producer.Errors() {
			logger.Error("Failed to write message:", zap.Error(err))
		}
	}()

	go func() {
		logger.Info("start send in queue")
		for {
			select {
			case <-ctx.Done():
				logger.Info("stop sending in queue")
				err := w.producer.Close()
				if err != nil {
					logger.Error("Error in queue prosucer", zap.Error(err))
				}
				return
			case successMsg := <-w.producer.Successes():
				successGauge.Inc()
				logger.Info("Successful to write message:", zap.Int64("offset", successMsg.Offset))
			}
		}
	}()
}
