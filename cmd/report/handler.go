package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	api "gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/gen/proto/go"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/logger"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type Handler struct {
	cfg Config
}

type Config interface {
	DefaultCurrency() string
	BotHost() string
	BotGrpcPort() int64
}

func NewHandler(cfg Config) *Handler {
	return &Handler{cfg: cfg}
}

func (h *Handler) generateReport(ctx context.Context, msg *sarama.ConsumerMessage) {
	request := &api.GenerateReportRequest{}
	if err := proto.Unmarshal(msg.Value, request); err != nil {
		logger.Error("cannot unmarshal report request", zap.Error(err))
	}

	id, err := uuid.FromBytes(request.UserId)
	if err != nil {
		logger.Error("converting from byte", zap.Error(err))
	}

	report, err := spendingStorage.GetReportByCategory(
		ctx,
		id,
		request.Currency,
		time.Unix(request.Start, 0),
		time.Unix(request.End, 0),
	)

	if err != nil {
		logger.Error("generating report error", zap.Error(err))
	}

	output, err := h.formatSpending(report, request.GetPeriod())
	if err != nil {
		logger.Error("error formating report", zap.Error(err))
	}

	client.SendReport(ctx, &api.ReportRequest{UserId: request.UserId, Report: output})
}

func (h *Handler) formatSpending(report dto.SpendingReport, period string) (string, error) {
	res := "У тебя нет трат за " + period
	if len(report) == 0 {
		return res, nil
	}
	currency := h.cfg.DefaultCurrency()

	res = "Траты за " + period
	for i, spendings := range report {
		sum := decimal.Decimal{}
		for _, v := range spendings {
			amount := v.Amount
			if h.cfg.DefaultCurrency() != v.Currency {
				currency = v.Currency
				rate := v.Rate
				amount = amount.Div(rate)
			}
			sum = sum.Add(amount)
		}

		res += fmt.Sprintf("\n%s : %s %s", i, sum.RoundBank(2).String(), currency)
	}
	return res, nil
}
