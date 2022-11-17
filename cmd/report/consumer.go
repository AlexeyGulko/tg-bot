package main

import (
	"context"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/logger"
)

var (
	KafkaTopic         = "reporter"
	KafkaConsumerGroup = "reporter-group"
	Assignor           = "range"
	lagGauge           = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "bot",
		Subsystem: "reporter",
		Name:      "queue_lag_total",
		Help:      "queue lag",
	})
)

// Consumer represents a Sarama consumer group consumer.
type Consumer struct {
	handler *Handler
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	logger.Info("consumer - setup")
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited.
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	logger.Info("consumer - cleanup")
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		consumer.handler.generateReport(session.Context(), message)
		session.MarkMessage(message, "")
		lag := claim.HighWaterMarkOffset() - message.Offset
		lagGauge.Add(float64(lag))
	}

	return nil
}

func startConsumerGroup(ctx context.Context, brokerList []string, consumerGroupHandler Consumer) error {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_8_0_0
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	switch Assignor {
	case "sticky":
		cfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategySticky}
	case "round-robin":
		cfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategyRoundRobin}
	case "range":
		cfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategyRange}
	default:
		return fmt.Errorf("unrecognized consumer group partition assignor: %s", Assignor)
	}

	// Create consumer group
	consumerGroup, err := sarama.NewConsumerGroup(brokerList, KafkaConsumerGroup, cfg)
	if err != nil {
		return fmt.Errorf("starting consumer group: %w", err)
	}

	err = consumerGroup.Consume(ctx, []string{KafkaTopic}, &consumerGroupHandler)
	if err != nil {
		return fmt.Errorf("consuming via handler: %w", err)
	}
	return nil
}
