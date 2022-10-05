package dto

import "time"

type Spending struct {
	Amount   int64
	Category string
	Date     time.Time
}

func (m Spending) IsEmpty() bool {
	return m == Spending{}
}
