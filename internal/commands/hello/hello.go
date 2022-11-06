package hello

import (
	"context"

	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/model/messages"
)

func Hello(tgClient commands.MessageSender, userStorage commands.UserStorage, config commands.Config) *commands.Command {
	return &commands.Command{
		CallBack: func(ctx context.Context, message *dto.Message) messages.CommandError {
			message.Command = "start"
			err := tgClient.SendMessage(ctx, "Привет! \n я подсчитываю твои расходы", message.UserID, nil)
			if err != nil {
				return commands.NewError(err, false)
			}
			_, err = userStorage.GetOrCreate(ctx, dto.User{TgID: message.UserID, Currency: config.DefaultCurrency()})
			if err != nil {
				return commands.NewError(err, false)
			}
			err = Help(tgClient).Execute(ctx, message)
			if err != nil {
				return commands.NewError(err, false)
			}
			return nil
		},
	}
}

func Help(tgClient commands.MessageSender) *commands.Command {
	return &commands.Command{
		CallBack: func(ctx context.Context, message *dto.Message) messages.CommandError {
			if len(message.Command) > 0 {
				message.Command = message.Command + "_help"
			} else {
				message.Command = "help"
			}

			err := tgClient.SendMessage(
				ctx,
				"Список комнд: \n /spend - добавить расход"+
					"\n/report - вывести сумму расходов за период "+
					"\n/currency изменить валюту"+
					"\n/budget установить бюджет на месяц",
				message.UserID,
				nil,
			)
			if err != nil {
				return commands.NewError(err, false)
			}
			return nil
		},
	}
}

func NotFoundCommand(tgClient commands.MessageSender) *commands.Command {
	return &commands.Command{
		CallBack: func(ctx context.Context, message *dto.Message) messages.CommandError {
			message.Command = "notfound"
			err := tgClient.SendMessage(ctx, "не знаю эту команду", message.UserID, nil)
			if err != nil {
				return commands.NewError(err, false)
			}
			err = Help(tgClient).Execute(ctx, message)
			if err != nil {
				return commands.NewError(err, false)
			}
			return nil
		},
	}
}

func StopCommand(tgClient commands.MessageSender) *commands.Command {
	return &commands.Command{
		CallBack: func(ctx context.Context, message *dto.Message) messages.CommandError {
			message.Command = "stop"
			err := tgClient.SendMessage(ctx, "Операция отменена", message.UserID, nil)
			if err != nil {
				return commands.NewError(err, false)
			}
			return nil
		},
	}
}
