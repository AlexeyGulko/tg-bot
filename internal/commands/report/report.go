package report

import (
	"context"
	"fmt"
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

func (c *CommandSequence) Menu(ctx context.Context, message *dto.Message) messages.CommandError {
	message.Command = "report_menu"
	err := c.tgClient.SendMessage(ctx, "Выбери период", message.UserID, keyboard)
	if err != nil {
		return commands.NewError(err, false)
	}

	return nil
}

func (c *CommandSequence) Report(ctx context.Context, message *dto.Message) messages.CommandError {
	message.Command = "report_input"
	start, periodName, err := c.getPeriod(message.Text)
	if err != nil {
		err = c.tgClient.SendMessage(ctx, "Выбери период из предложенных на клавиатуре", message.UserID, keyboard)
		return commands.NewError(err, true)
	}
	user, err := c.userStorage.GetOrCreate(ctx, dto.User{TgID: message.UserID, Currency: c.config.DefaultCurrency()})
	if err != nil {
		return commands.NewError(err, false)
	}
	report, err := c.spendStorage.GetReportByCategory(ctx, user.ID, start, time.Now())

	if err != nil {
		return commands.NewError(err, false)
	}
	removeMarkup := tgbotapi.NewRemoveKeyboard(true)

	output, err := c.formatSpending(ctx, report, periodName)
	if err != nil {
		return commands.NewError(err, false)
	}

	err = c.tgClient.SendMessage(ctx, output, message.UserID, removeMarkup)
	if err != nil {
		return commands.NewError(err, false)
	}
	return nil
}

func (c *CommandSequence) formatSpending(
	ctx context.Context,
	report dto.SpendingReport,
	period string,
) (string, error) {
	res := "У тебя нет трат за " + period
	if len(report) == 0 {
		return res, nil
	}
	currency := c.config.DefaultCurrency()
	res = "Траты за " + period
	for i, spendings := range report {
		sum := decimal.Decimal{}
		for _, v := range spendings {
			amount := v.Amount
			if c.config.DefaultCurrency() != v.Currency {
				currency = v.Currency
				var rate = v.Rate
				if v.Rate.Equal(decimal.Decimal{}) {
					r, err := c.currencyService.GetRate(ctx, v.Currency, v.Date)
					rate = r.Rate
					if err != nil {
						return "", err
					}
				}

				amount = amount.Div(rate)
			}
			sum = sum.Add(amount)
		}

		res += fmt.Sprintf("\n%s : %s %s", i, sum.RoundBank(2).String(), currency)
	}
	return res, nil
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
