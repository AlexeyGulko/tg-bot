package report

import (
	"context"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/model/messages"
)

const (
	year  = "год"
	month = "месяц"
	week  = "неделя"
)

var keyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(year),
		tgbotapi.NewKeyboardButton(month),
		tgbotapi.NewKeyboardButton(week),
	),
)

type CommandSequence struct {
	tgClient        commands.MessageSender
	spendStorage    commands.SpendingStorage
	config          commands.Config
	userStorage     commands.UserStorage
	currencyService commands.CurrencyService
}

func New(
	tgClient commands.MessageSender,
	spendStorage commands.SpendingStorage,
	config commands.Config,
	userStorage commands.UserStorage,
	currencyService commands.CurrencyService,
) commands.Command {
	comm := &CommandSequence{
		tgClient:        tgClient,
		spendStorage:    spendStorage,
		config:          config,
		userStorage:     userStorage,
		currencyService: currencyService,
	}

	return commands.Command{CallBack: comm.Menu}.SetNext(commands.Command{CallBack: comm.Report})
}

func (c *CommandSequence) Menu(ctx context.Context, message dto.Message) messages.CommandError {
	err := c.tgClient.SendMessage("Выбери период", message.UserID, keyboard)
	if err != nil {
		log.Printf("%s", err)
	}

	return nil
}

func (c *CommandSequence) Report(ctx context.Context, message dto.Message) messages.CommandError {
	period, periodName, err := c.getPeriod(message.Text)
	if err != nil {
		err = c.tgClient.SendMessage("Выбери период из предложенных на клавиатуре", message.UserID, keyboard)
		log.Printf("%s", err)
		return &commands.CommandError{Retry: true}
	}
	report := c.spendStorage.GetReportByCategory(message.UserID, period)

	removeMarkup := tgbotapi.NewRemoveKeyboard(true)
	err = c.tgClient.SendMessage(c.formatSpending(ctx, message.UserID, report, periodName), message.UserID, removeMarkup)

	if err != nil {
		log.Printf("%s", err)
	}
	return nil
}

func (c *CommandSequence) formatSpending(ctx context.Context, serID int64, report map[string][]dto.Spending, period string) string {
	res := "У тебя нет трат за " + period
	if len(report) == 0 {
		return res
	}

	user, _ := c.userStorage.Get(serID)
	res = "Траты за " + period
	for i, spendings := range report {
		sum := decimal.Decimal{}
		for _, v := range spendings {
			amount := v.Amount
			if c.config.DefaultCurrency() != user.Currency {
				amount, _ = c.currencyService.ConvertTo(ctx, user.Currency, v.Amount, v.Date)
			}
			sum = sum.Add(amount)
		}

		res += fmt.Sprintf("\n%s : %s %s", i, sum.RoundBank(2).String(), user.Currency)
	}
	return res
}

func (c *CommandSequence) getPeriod(key string) (time.Time, string, error) {
	period := time.Now()
	var name string
	var err error

	switch key {
	case "год":
		period = period.AddDate(-1, 0, 0)
		name = "год"
	case "месяц":
		period = period.AddDate(0, -1, 0)
		name = "месяц"
	case "неделя":
		period = period.AddDate(0, 0, -7)
		name = "неделю"
	default:
		err = errors.New("wrong period key")
	}

	return period, name, err
}
