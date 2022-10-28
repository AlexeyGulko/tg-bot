package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Message struct {
	Text   string
	UserID int64
}

type Spending struct {
	ID       uuid.UUID
	UserID   uuid.UUID
	Amount   decimal.Decimal
	Category string
	Date     time.Time
	Created  time.Time
	Updated  time.Time
}

type User struct {
	ID          uuid.UUID
	TgID        int64
	Currency    string
	MonthBudget decimal.Decimal
	Created     time.Time
	Updated     time.Time
}

type Currency struct {
	ID        uuid.UUID
	Code      string
	Rate      decimal.Decimal
	TimeStamp time.Time
	Created   time.Time
	Updated   time.Time
}

type CurrencyMap map[string]Currency
