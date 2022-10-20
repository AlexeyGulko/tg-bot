package currency

import (
	"log"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/workers/update_rates"
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
	Get(time time.Time, code string) (dto.Currency, bool)
	Add(time time.Time, currency dto.Currency)
}

func (s *Service) UpdateRates(ctx context.Context, date time.Time) error {
	year, month, day := date.Date()
	date = time.Date(year, month, day, 0, 0, 0, 0, date.Location())

	var code string
	for _, item := range s.config.Currencies() {
		if item != s.config.DefaultCurrency() {
			code = item
			break
		}
	}

	_, has := s.storage.Get(date, code)
	if has {
		return nil
	}
	log.Printf("update rates on date %s", date.Format("02 01 2006"))
	allRates, err := s.RatesClient.GetExchangeRates(ctx, date)
	if err != nil {
		return err
	}

	availableCurrencies := make(map[string]interface{}, len(s.config.Currencies()))

	for _, v := range s.config.Currencies() {
		availableCurrencies[v] = struct{}{}
	}

	for _, v := range allRates {
		if _, has := availableCurrencies[v.Code]; !has {
			continue
		}

		s.storage.Add(date, v)
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
	return amount.Div(rate), nil
}

func (s *Service) ConvertFrom(ctx context.Context, code string, amount decimal.Decimal, time time.Time) (decimal.Decimal, error) {
	if s.config.DefaultCurrency() == code {
		return amount, nil
	}
	rate, err := s.GetRate(ctx, code, time)
	if err != nil {
		return decimal.Decimal{}, err
	}
	return amount.Mul(rate), nil
}

func (s *Service) GetRate(ctx context.Context, code string, date time.Time) (decimal.Decimal, error) {
	year, month, day := date.Date()
	date = time.Date(year, month, day, 0, 0, 0, 0, date.Location())
	rate, has := s.storage.Get(date, code)

	if !has {
		var wg sync.WaitGroup
		r := update_rates.ChannelR{T: date, Wg: &wg}
		r.Wg.Add(1)
		s.ch <- r
		r.Wg.Wait()

		rate, _ = s.storage.Get(date, code)
	}

	return rate.Rate, nil
}
