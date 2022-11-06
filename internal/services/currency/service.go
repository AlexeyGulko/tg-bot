package currency

import (
	"database/sql"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/shopspring/decimal"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/logger"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/workers/update_rates"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type Service struct {
	RatesClient RatesClient
	config      config
	storage     Storage
	ch          update_rates.Channel
}

type RatesClient interface {
	GetExchangeRates(ctx context.Context, date time.Time) ([]dto.Currency, error)
}

func New(RatesClient RatesClient, config config, storage Storage, updateRatesCh update_rates.Channel) *Service {
	return &Service{RatesClient: RatesClient, config: config, storage: storage, ch: updateRatesCh}
}

type config interface {
	Currencies() []string
	DefaultCurrency() string
}

type Storage interface {
	Get(ctx context.Context, time time.Time, code string) (*dto.Currency, error)
	Add(ctx context.Context, currency dto.Currency) error
	AddBulk(ctx context.Context, currencies []dto.Currency) error
}

func (s *Service) UpdateRates(ctx context.Context, date time.Time) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "currency service update rates")
	defer span.Finish()
	year, month, day := date.Date()
	date = time.Date(year, month, day, 0, 0, 0, 0, date.Location())

	var code string
	for _, item := range s.config.Currencies() {
		if item != s.config.DefaultCurrency() {
			code = item
			break
		}
	}

	_, err := s.storage.Get(ctx, date, code)

	if err != sql.ErrNoRows && err != nil {
		return err
	}

	logger.Info("update rates on date", zap.String("date", date.Format("02 01 2006")))
	allRates, err := s.RatesClient.GetExchangeRates(ctx, date)
	if err != nil {
		return err
	}

	availableCurrencies := make(map[string]interface{}, len(s.config.Currencies()))

	for _, v := range s.config.Currencies() {
		availableCurrencies[v] = struct{}{}
	}

	filtered := make([]dto.Currency, 0, len(s.config.Currencies()))
	for _, v := range allRates {
		if _, has := availableCurrencies[v.Code]; !has {
			continue
		}
		filtered = append(filtered, v)
	}

	err = s.storage.AddBulk(ctx, filtered)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) ConvertTo(ctx context.Context, code string, amount decimal.Decimal, time time.Time) (decimal.Decimal, error) {
	if s.config.DefaultCurrency() == code {
		return amount, nil
	}
	rate, err := s.GetRate(ctx, code, time)
	if err != nil {
		return decimal.Decimal{}, err
	}
	return amount.Div(rate.Rate), nil
}

func (s *Service) ConvertFrom(ctx context.Context, code string, amount decimal.Decimal, time time.Time) (decimal.Decimal, error) {
	if s.config.DefaultCurrency() == code {
		return amount, nil
	}
	rate, err := s.GetRate(ctx, code, time)
	if err != nil {
		return decimal.Decimal{}, err
	}
	return amount.Mul(rate.Rate), nil
}

func (s *Service) GetRate(ctx context.Context, code string, date time.Time) (*dto.Currency, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "currency srv get rate")
	defer span.Finish()
	year, month, day := date.Date()
	date = time.Date(year, month, day, 0, 0, 0, 0, date.Location())
	rate, err := s.storage.Get(ctx, date, code)

	if err == sql.ErrNoRows {
		var wg sync.WaitGroup
		r := update_rates.ChannelR{T: date, Wg: &wg}
		r.Wg.Add(1)
		s.ch <- r
		r.Wg.Wait()

		rate, _ = s.storage.Get(ctx, date, code)
	}

	return rate, nil
}
