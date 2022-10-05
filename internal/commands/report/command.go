package report

import (
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/model/messages"
)

type Command struct {
	spendingStorage commands.SpendingStorage
	tgClient        commands.MessageSender
	storage         commands.Storage
}

func New(spendStorage commands.SpendingStorage, tgClient commands.MessageSender, storage commands.Storage) *Command {
	return &Command{
		spendingStorage: spendStorage,
		tgClient:        tgClient,
		storage:         storage,
	}
}

func (c *Command) Execute(message messages.Message) bool {
	intercept := true
	_, ok := c.storage.Get(message.UserID)
	if !ok {
		c.storage.Add(message.UserID, c)
		var keyboard = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("год"),
				tgbotapi.NewKeyboardButton("месяц"),
				tgbotapi.NewKeyboardButton("неделя"),
			),
		)

		err := c.tgClient.SendMessage("Выбери период", message.UserID, keyboard)
		if err != nil {
			log.Printf("%s", err)
		}
		return intercept
	}

	period, periodName, err := getPeriod(message.Text)
	if err != nil {
		err = c.tgClient.SendMessage("Выбери период из предложенных на клавиатуре", message.UserID, nil)
		log.Printf("%s", err)
		return true
	}
	report := c.spendingStorage.GetReportByCategory(message.UserID, period)

	removeMarkup := tgbotapi.NewRemoveKeyboard(true)
	err = c.tgClient.SendMessage(formatSpending(report, periodName), message.UserID, removeMarkup)

	if err != nil {
		log.Printf("%s", err)
	}
	c.storage.Delete(message.UserID)
	return false
}

func formatSpending(report map[string]int64, period string) string {
	res := "У тебя нет трат за " + period
	if len(report) == 0 {
		return res
	}

	res = "Траты за " + period
	for i, v := range report {
		res += fmt.Sprintf("\n%s : %.2f", i, float64(v)/100)
	}
	return res
}

func getPeriod(key string) (time.Time, string, error) {
	period := time.Now()
	var name string
	var err error

	println(key)
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
