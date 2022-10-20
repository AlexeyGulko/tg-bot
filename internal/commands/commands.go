package commands

import (
	"time"

	"github.com/shopspring/decimal"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/model/messages"
	"golang.org/x/net/context"
)

type Storage interface {
	Add(int64, Command)
	Get(int64) (Command, bool)
	Delete(int64)
}

type CommandError struct {
	Retry bool
	text  string
}

func (e CommandError) DoRetry() bool {
	return e.Retry
}

func (e CommandError) Error() string {
	return e.text
}

type SpendingStorage interface {
	Add(int64, dto.Spending)
	Get(int64) ([]dto.Spending, bool)
	GetReportByCategory(int64, time.Time) map[string][]dto.Spending
}

type UserStorage interface {
	Add(dto.User)
	Get(int64) (dto.User, bool)
}

type MessageSender interface {
	SendMessage(text string, userId int64, markup interface{}) error
}

type Config interface {
	Currencies() []string
	DefaultCurrency() string
}

type CurrencyService interface {
	ConvertFrom(context.Context, string, decimal.Decimal, time.Time) (decimal.Decimal, error)
	ConvertTo(context.Context, string, decimal.Decimal, time.Time) (decimal.Decimal, error)
}

type Command struct {
	Finished    bool
	NextCommand messages.Command
	Retry       bool
	CallBack    func(ctx context.Context, message dto.Message) messages.CommandError
}

func (c Command) DoRetry() bool {
	return c.Retry
}

func (c Command) Execute(ctx context.Context, message dto.Message) messages.CommandError {
	return c.CallBack(ctx, message)
}

func (c Command) SetNext(comm Command) Command {
	c.NextCommand = comm
	return c
}

func (c Command) Next() (messages.Command, bool) {
	return c.NextCommand, c.NextCommand != nil
}
