package hello

import (
	"context"
	"log"

	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/model/messages"
)

func Hello(tgClient commands.MessageSender, userStorage commands.UserStorage, config commands.Config) *commands.Command {
	return &commands.Command{
		CallBack: func(ctx context.Context, message dto.Message) messages.CommandError {
			err := tgClient.SendMessage("Привет! \n я подсчитываю твои расходы", message.UserID, nil)
			if err != nil {
				log.Printf("%s", err)
			}

			_, has := userStorage.Get(message.UserID)

			if !has {
				userStorage.Add(dto.User{ID: message.UserID, Currency: config.DefaultCurrency()})
			}
			return Help(tgClient).Execute(ctx, message)
		},
	}
}

func Help(tgClient commands.MessageSender) *commands.Command {
	return &commands.Command{
		CallBack: func(ctx context.Context, message dto.Message) messages.CommandError {
			err := tgClient.SendMessage(
				"Список комнд: \n /spend - добавить расход"+
					"\n/report - вывести сумму расходов за период "+
					"\n/currency изменить валюту",
				message.UserID,
				nil,
			)
			if err != nil {
				log.Printf("%s", err)
			}

			return nil
		},
	}
}

func NotFoundCommand(tgClient commands.MessageSender) *commands.Command {
	return &commands.Command{
		CallBack: func(ctx context.Context, message dto.Message) messages.CommandError {
			err := tgClient.SendMessage("не знаю эту команду", message.UserID, nil)
			if err != nil {
				log.Printf("%s", err)
			}

			return nil
		},
	}
}

func StopCommand(tgClient commands.MessageSender) *commands.Command {
	return &commands.Command{
		CallBack: func(ctx context.Context, message dto.Message) messages.CommandError {
			err := tgClient.SendMessage("Операция отменена", message.UserID, nil)
			if err != nil {
				log.Printf("%s", err)
			}

			return nil
		},
	}
}
