package spend

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/shopspring/decimal"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/helpers"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/model/messages"
	"golang.org/x/net/context"
)

var sumRE = regexp.MustCompile(`(^\d+$)|((^\d+)[\.|,](\d{1,2}$))`)

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

func (c *CommandSequence) StartMessage(ctx context.Context, message dto.Message) messages.CommandError {
	c.tempSpendings[message.UserID] = dto.Spending{}
	err := c.tgClient.SendMessage("Введи категорию", message.UserID, nil)
	if err != nil {
		log.Printf("%s", err)
	}

	return nil
}

func (c *CommandSequence) Category(ctx context.Context, message dto.Message) messages.CommandError {
	spending := c.tempSpendings[message.UserID]
	spending.Category = message.Text
	c.tempSpendings[message.UserID] = spending
	rdate := helpers.RandomDate().Format("2 1 2006")

	err := c.tgClient.SendMessage(fmt.Sprintf("Введи дату в формате %s", rdate), message.UserID, nil)
	if err != nil {
		log.Print(err.Error())
	}

	return nil
}

func (c *CommandSequence) Sum(ctx context.Context, message dto.Message) messages.CommandError {
	var sum string
	var err error
	matches := sumRE.FindStringSubmatch(message.Text)

	if matches == nil {
		err := c.tgClient.SendMessage("Неправильный формат суммы попробуй ещё", message.UserID, nil)
		if err != nil {
			log.Print(err.Error())
		}
		return commands.CommandError{Retry: true}
	}
	//первая группа захвата - значение без дробей
	if len(matches[1]) > 0 {
		sum = matches[1]
	}

	//третья группа целая часть, четвертая - дробная
	if len(matches[3]) > 0 && len(matches[4]) > 0 {
		sum = fmt.Sprintf("%s.%s", matches[3], matches[4])
	}

	if err != nil {
		err := c.tgClient.SendMessage("Неправильный формат суммы попробуй ещё", message.UserID, nil)
		if err != nil {
			log.Print(err.Error())
		}
		return commands.CommandError{Retry: true}
	}

	amount, err := decimal.NewFromString(sum)
	if err != nil {
		err := c.tgClient.SendMessage("Неправильный формат суммы попробуй ещё", message.UserID, nil)
		if err != nil {
			log.Print(err.Error())
		}
		return commands.CommandError{Retry: true}
	}

	user, _ := c.userStorage.Get(message.UserID)
	spending := c.tempSpendings[message.UserID]
	if user.Currency != c.config.DefaultCurrency() {
		amount, err = c.currencyService.ConvertFrom(ctx, user.Currency, amount, spending.Date)

		if err != nil {
			return commands.CommandError{Retry: true}
		}
	}

	spending.Amount = amount
	c.tempSpendings[message.UserID] = spending

	c.spendStorage.Add(message.UserID, spending)
	delete(c.tempSpendings, message.UserID)

	finMsg :=
		"\nкатегория: " + spending.Category +
			"\nсумма: " + sum + " " + user.Currency +
			"\nдата: " + spending.Date.Format("2 1 2006") +
			"\nуспешно добавлена"

	err = c.tgClient.SendMessage(finMsg, message.UserID, nil)
	if err != nil {
		log.Print(err.Error())
	}

	return nil
}

func (c *CommandSequence) Date(ctx context.Context, message dto.Message) messages.CommandError {
	t, err := time.ParseInLocation("2 1 2006", message.Text, time.Now().Location())
	rdate := helpers.RandomDate().Format("2 1 2006")
	if err != nil {
		err := c.tgClient.SendMessage(fmt.Sprintf("Неправильный формат, попробуй в таком %s", rdate), message.UserID, nil)
		if err != nil {
			log.Print(err)
		}

		return &commands.CommandError{Retry: true}
	}

	spending := c.tempSpendings[message.UserID]
	spending.Date = t
	c.tempSpendings[message.UserID] = spending
	err = c.tgClient.SendMessage("Введи сумму", message.UserID, nil)
	if err != nil {
		log.Print(err.Error())
	}

	return nil
}
