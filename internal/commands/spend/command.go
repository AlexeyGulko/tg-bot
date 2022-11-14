package spend

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/helpers"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/model/messages"
	"golang.org/x/net/context"
)

type CommandSequence struct {
	tgClient        commands.MessageSender
	spendStorage    commands.SpendingStorage
	config          commands.Config
	userStorage     commands.UserStorage
	currencyService commands.CurrencyService
	tempSpendings   map[int64]dto.Spending
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
		tempSpendings:   make(map[int64]dto.Spending),
	}

	return commands.Command{CallBack: comm.StartMessage}.SetNext(
		commands.Command{CallBack: comm.Category}.SetNext(
			commands.Command{CallBack: comm.Date}.SetNext(
				commands.Command{CallBack: comm.Sum},
			),
		),
	)
}

func (c *CommandSequence) StartMessage(ctx context.Context, message *dto.Message) messages.CommandError {
	message.Command = "spending_start"
	user, err := c.userStorage.GetOrCreate(ctx, dto.User{TgID: message.UserID, Currency: c.config.DefaultCurrency()})
	if err != nil {
		return commands.NewError(err, false)
	}
	err = c.tgClient.SendMessage(ctx, "Введи категорию", message.UserID, nil)
	c.tempSpendings[message.UserID] = dto.Spending{UserID: user.ID}
	if err != nil {
		return commands.NewError(err, false)
	}

	return nil
}

func (c *CommandSequence) Category(ctx context.Context, message *dto.Message) messages.CommandError {
	message.Command = "spending_input_category"
	spending := c.tempSpendings[message.UserID]
	spending.Category = message.Text
	c.tempSpendings[message.UserID] = spending
	rdate := helpers.RandomDate().Format("2 1 2006")

	err := c.tgClient.SendMessage(ctx, fmt.Sprintf("Введи дату в формате %s", rdate), message.UserID, nil)
	if err != nil {
		return commands.NewError(err, false)
	}

	return nil
}

func (c *CommandSequence) Sum(ctx context.Context, message *dto.Message) messages.CommandError {
	message.Command = "spending_input_amount"
	var err error
	amount, err := commands.ParseDigitInput(message.Text)
	if err != nil {
		err := c.tgClient.SendMessage(ctx, "Неправильный формат суммы попробуй ещё", message.UserID, nil)
		if err != nil {
			return commands.NewError(err, false)
		}
		return commands.NewError(err, true)
	}

	user, err := c.userStorage.Get(ctx, message.UserID)
	if err != nil {
		return commands.NewError(err, false)
	}
	spending := c.tempSpendings[message.UserID]

	year, month, _ := spending.Date.Date()
	firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, spending.Date.Location())
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	spent, err := c.spendStorage.GetSpendingAmount(ctx, user.ID, firstOfMonth, lastOfMonth)
	if err != sql.ErrNoRows && err != nil {
		return commands.NewError(err, false)
	}

	budget := user.MonthBudget

	var rate decimal.Decimal
	if user.Currency != c.config.DefaultCurrency() {
		r, err := c.currencyService.GetRate(ctx, user.Currency, spending.Date)
		if err != nil {
			return commands.NewError(err, false)
		}
		if r == nil {
			err = c.tgClient.SendMessage(ctx, "неверная дата", message.UserID, nil)
			if err != nil {
				return commands.NewError(err, false)
			}
			return commands.NewError(errors.New("no rates in source"), false)
		}
		rate = r.Rate
		amount = amount.Mul(rate)
	}

	if year == time.Now().Year() && month == time.Now().Month() {
		if err := c.checkBudget(ctx, spent, amount, budget, rate, user); err != nil {
			return commands.NewError(err, false)
		}
	}

	spending.Amount = amount

	err = c.spendStorage.Add(ctx, &spending)
	delete(c.tempSpendings, message.UserID)
	if err != nil {
		return commands.NewError(err, false)
	}

	finMsg :=
		"\nкатегория: " + spending.Category +
			"\nсумма: " + message.Text + " " + user.Currency +
			"\nдата: " + spending.Date.Format("2 1 2006") +
			"\nуспешно добавлена"

	err = c.tgClient.SendMessage(ctx, finMsg, message.UserID, nil)
	if err != nil {
		return commands.NewError(err, false)
	}

	return nil
}

func (c *CommandSequence) checkBudget(
	ctx context.Context,
	spent decimal.Decimal,
	amount decimal.Decimal,
	budget decimal.Decimal,
	rate decimal.Decimal,
	user *dto.User,
) error {
	if spent.Add(amount).GreaterThan(budget) && !user.MonthBudget.Equal(decimal.New(0, 0)) {
		if !rate.Equal(decimal.Decimal{}) {
			spent = spent.Div(rate)
			budget = budget.Div(rate)
		}

		err := c.tgClient.SendMessage(
			ctx,
			fmt.Sprintf(
				"В текущем месяце ты превысил бюджет %s %s\n"+
					"ты можешь потратиить %s %s или увеличить бюджет командой /budget",
				budget.RoundBank(2),
				user.Currency,
				budget.Sub(spent).RoundBank(2),
				user.Currency,
			),
			user.TgID,
			nil,
		)
		if err != nil {
			return err
		}
		return errors.New("budget overcome")
	}

	return nil
}

func (c *CommandSequence) Date(ctx context.Context, message *dto.Message) messages.CommandError {
	message.Command = "spending_input_date"
	user, err := c.userStorage.Get(ctx, message.UserID)
	if err != nil {
		return commands.NewError(err, false)
	}

	loc, err := time.LoadLocation("GMT")
	if err != nil {
		return commands.NewError(err, false)
	}
	t, err := time.ParseInLocation("2 1 2006", message.Text, loc)
	rdate := helpers.RandomDate().Format("2 1 2006")
	if err != nil {
		err := c.tgClient.SendMessage(ctx, fmt.Sprintf("Неправильный формат, попробуй в таком %s", rdate), message.UserID, nil)
		if err != nil {
			return commands.NewError(err, false)
		}

		return commands.NewError(err, true)
	}

	currency := user.Currency
	if currency == c.config.DefaultCurrency() {
		for _, item := range c.config.Currencies() {
			if item != c.config.DefaultCurrency() {
				currency = item
				break
			}
		}
	}

	r, err := c.currencyService.GetRate(ctx, currency, t)
	if err != nil {
		return commands.NewError(err, false)
	}
	if r == nil {
		err = c.tgClient.SendMessage(ctx, "дата за рамками времени и пространства", message.UserID, nil)
		if err != nil {
			return commands.NewError(err, false)
		}
		return commands.NewError(errors.New("no rates in source"), true)
	}

	spending := c.tempSpendings[message.UserID]
	spending.Date = t
	c.tempSpendings[message.UserID] = spending
	err = c.tgClient.SendMessage(ctx, "Введи сумму", message.UserID, nil)
	if err != nil {
		return commands.NewError(err, false)
	}

	return nil
}
