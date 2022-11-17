package commands

import (
	"fmt"
	"go/types"
	"regexp"
	"time"

	"github.com/google/uuid"
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
	retry bool
	err   error
}

func NewError(err error, retry bool) *CommandError {
	return &CommandError{err: err, retry: retry}
}

func (e *CommandError) DoRetry() bool {
	return e.retry
}

func (e *CommandError) Error() string {
	return e.err.Error()
}

type SpendingStorage interface {
	Add(context.Context, *dto.Spending) error
	GetReportByCategory(ctx context.Context, UserID uuid.UUID, currency string, start time.Time, end time.Time) (dto.SpendingReport, error)
	GetSpendingAmount(ctx context.Context, UserID uuid.UUID, start time.Time, end time.Time) (decimal.Decimal, error)
}

type UserStorage interface {
	Add(context.Context, dto.User) error
	Get(ctx context.Context, tgId int64) (*dto.User, error)
	GetOrCreate(context.Context, dto.User) (*dto.User, error)
	Update(context.Context, *dto.User) error
}

type MessageSender interface {
	SendMessage(ctx context.Context, text string, userId int64, markup interface{}) error
}

type Config interface {
	Currencies() []string
	DefaultCurrency() string
}

type CurrencyService interface {
	ConvertFrom(context.Context, string, decimal.Decimal, time.Time) (decimal.Decimal, error)
	ConvertTo(context.Context, string, decimal.Decimal, time.Time) (decimal.Decimal, error)
	GetRate(ctx context.Context, code string, date time.Time) (*dto.Currency, error)
}

type Command struct {
	Finished    bool
	NextCommand messages.Command
	Retry       bool
	CallBack    func(ctx context.Context, message *dto.Message) messages.CommandError
}

var DigitInput = regexp.MustCompile(`(^\d+$)|((^\d+)[\s|\.|,](\d{1,2}$))`)

func (c Command) DoRetry() bool {
	return c.Retry
}

func (c Command) Execute(ctx context.Context, message *dto.Message) messages.CommandError {
	return c.CallBack(ctx, message)
}

func (c Command) SetNext(comm Command) Command {
	c.NextCommand = comm
	return c
}

func (c Command) Next() (messages.Command, bool) {
	return c.NextCommand, c.NextCommand != nil
}

func ParseDigitInput(input string) (decimal.Decimal, error) {
	var parsed string
	matches := DigitInput.FindStringSubmatch(input)

	if matches == nil {
		return decimal.Decimal{}, types.Error{Msg: "No matches"}
	}
	//первая группа захвата - значение без дробей
	if len(matches[1]) > 0 {
		parsed = matches[1]
	}

	//третья группа целая часть, четвертая - дробная
	if len(matches[3]) > 0 && len(matches[4]) > 0 {
		parsed = fmt.Sprintf("%s.%s", matches[3], matches[4])
	}

	res, err := decimal.NewFromString(parsed)

	if err != nil {
		return decimal.Decimal{}, err
	}

	return res, err
}
