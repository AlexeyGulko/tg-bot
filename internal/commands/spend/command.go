package spend

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/helpers"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/model/messages"
)

type Command struct {
	step         int64
	spending     dto.Spending
	storage      commands.Storage
	spendStorage commands.SpendingStorage
	tgClient     commands.MessageSender
}

var sumRE = regexp.MustCompile(`(^\d+$)|((^\d+)[\.|,](\d{1,2}$))`)

func New(storage commands.Storage, spendStorage commands.SpendingStorage, tgClient commands.MessageSender) *Command {
	return &Command{storage: storage, spendStorage: spendStorage, tgClient: tgClient}
}

func (c *Command) Execute(message messages.Message) bool {
	intercept := true
	command, ok := c.storage.Get(message.UserID)
	if !ok {

		c.storage.Add(message.UserID, command)
		c.step = 1
		res, _ := c.Handle("")
		err := c.tgClient.SendMessage(res, message.UserID, nil)
		if err != nil {
			log.Printf("%s", err)
		}
		return intercept
	}

	res, done := c.Handle(message.Text)

	if done {
		model, _ := c.GetSpending()
		c.spendStorage.Add(message.UserID, model)
		c.storage.Delete(message.UserID)
		intercept = false
	}

	err := c.tgClient.SendMessage(res, message.UserID, nil)
	if err != nil {
		log.Printf("%s", err)
	}
	return intercept
}

func (c *Command) GetSpending() (dto.Spending, bool) {
	return c.spending, c.spending.IsEmpty()
}

func (c *Command) Handle(msg string) (string, bool) {
	//TODO порефакторить на декоратор
	if msg == "/stop" {
		c.spending = dto.Spending{}
		return "Добавление траты отменено", true
	}

	switch c.step {
	case 1:
		c.step++
		c.spending = dto.Spending{}
		return "Введи категорию", false
	case 2:
		c.spending.Category = msg
		c.step++
		return "Введи сумму", false
	case 3:
		var sum int
		var err error
		matches := sumRE.FindStringSubmatch(msg)

		if matches == nil {
			return "Неверная сумма", false
		}
		//первая группа захвата - значение без дробей
		if len(matches[1]) > 0 {
			sum, err = strconv.Atoi(matches[1])
			sum *= 100
		}

		//третья группа целая часть, четвертая - дробная
		if len(matches[3]) > 0 && len(matches[4]) > 0 {
			sum, err = strconv.Atoi(matches[3] + matches[4])
		}

		if err != nil {
			return "Неверная сумма", false
		}

		c.spending.Amount = int64(sum)
		c.step++
		rdate := helpers.RandomDate().Format("2 1 2006")
		return fmt.Sprintf("Введи день, месяц, год - например %s", rdate), false
	case 4:
		t, err := time.Parse("2 1 2006", msg)
		rdate := helpers.RandomDate().Format("2 1 2006")
		if err != nil {
			return fmt.Sprintf("Неправильный формат, попробуй в таком %s", rdate), false
		}

		c.spending.Date = t

		sum := fmt.Sprintf("%.2f", float64(c.spending.Amount)/100)

		finMsg :=
			"\nкатегория: " + c.spending.Category +
				"\nсумма: " + sum +
				"\nдата: " + c.spending.Date.Format("2 1 2006") +
				"\nуспешно добавлена"

		return finMsg, true
	default:
		return "", false
	}
}
