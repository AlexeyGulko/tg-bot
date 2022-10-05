package commands

import (
	"log"
	"time"

	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/model/messages"
)

type Storage interface {
	Add(int64, messages.Command)
	Get(int64) (messages.Command, bool)
	Delete(int64)
}

type SpendingStorage interface {
	Add(int64, dto.Spending)
	Get(int64) ([]dto.Spending, bool)
	GetReportByCategory(int64, time.Time) map[string]int64
}

type MessageSender interface {
	SendMessage(text string, userId int64, markup interface{}) error
}

type Command struct {
	CallBack func(messages.Message) bool
}

func (c Command) Execute(msg messages.Message) bool {
	return c.CallBack(msg)
}

func Hello(tgClient MessageSender) Command {
	return Command{
		CallBack: func(message messages.Message) bool {
			err := tgClient.SendMessage("Привет! \n я подсчитываю твои расходы", message.UserID, nil)
			if err != nil {
				log.Printf("%s", err)
			}

			Help(tgClient).Execute(message)
			return false
		},
	}
}

func Help(tgClient MessageSender) Command {
	return Command{
		CallBack: func(message messages.Message) bool {
			err := tgClient.SendMessage(
				"Список комнд: \n /spend - добавить расход \n "+
					"вывести сумму расходов за период - /report",
				message.UserID,
				nil,
			)
			if err != nil {
				log.Printf("%s", err)
			}
			return false
		},
	}
}

func NotFoundCommand(tgClient MessageSender) Command {
	return Command{
		CallBack: func(message messages.Message) bool {
			err := tgClient.SendMessage("не знаю эту команду", message.UserID, nil)
			if err != nil {
				log.Printf("%s", err)
			}
			return false
		},
	}
}
