package currency

import (
	"time"

	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
)

type Storage struct {
	store map[DateKey]dto.Currency
}

type DateKey struct {
	date time.Time
	code string
}

func NewStorage() *Storage {
	return &Storage{store: make(map[DateKey]dto.Currency)}
}

func (s *Storage) Add(date time.Time, currency dto.Currency) {
	key := DateKey{date: date, code: currency.Code}
	s.store[key] = currency
}

func (s *Storage) Get(date time.Time, code string) (dto.Currency, bool) {
	key := DateKey{date: date, code: code}
	val, has := s.store[key]
	return val, has
}
