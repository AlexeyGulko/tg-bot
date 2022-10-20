package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type Message struct {
	Text   string
	UserID int64
}

type Spending struct {
	Amount   decimal.Decimal
	Category string
	Date     time.Time
}

type User struct {
	ID       int64
	Currency string
}

type Currency struct {
	Code string
	Rate decimal.Decimal
}

type CurrencyMap map[string]Currency
