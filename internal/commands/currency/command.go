package currency

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/helpers"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/model/messages"
)

func Menu(tgClient commands.MessageSender, cfg commands.Config, userStorage commands.UserStorage) *commands.Command {
	currencies := cfg.Currencies()

	keyboard := helpers.CreateMarkupMenu(currencies, 2)

	return &commands.Command{
		CallBack: func(ctx context.Context, message *dto.Message) messages.CommandError {
			message.Command = "currency_menu"
			user, err := userStorage.GetOrCreate(ctx, dto.User{TgID: message.UserID, Currency: cfg.DefaultCurrency()})
			if err != nil {
				return commands.NewError(err, false)
			}
			err = tgClient.SendMessage(ctx, fmt.Sprintf("Текущая валюта: %s\nВыбери валюту", user.Currency), message.UserID, keyboard)
			if err != nil {
				return commands.NewError(err, false)
			}

			return nil
		},
	}
}

func Input(tgClient commands.MessageSender, config commands.Config, userStorage commands.UserStorage) commands.Command {
	return commands.Command{
		CallBack: func(ctx context.Context, message *dto.Message) messages.CommandError {
			message.Command = "currency_input"
			currencies := config.Currencies()

			currMap := make(map[string]interface{}, len(currencies))
			for _, v := range currencies {
				currMap[v] = struct{}{}
			}

			_, has := currMap[message.Text]

			if !has {
				keyboard := helpers.CreateMarkupMenu(currencies, 2)
				err := tgClient.SendMessage(ctx, "Выбери валюту", message.UserID, keyboard)
				return commands.NewError(err, true)
			}

			removeMarkup := tgbotapi.NewRemoveKeyboard(true)

			err := tgClient.SendMessage(
				ctx,
				fmt.Sprintf("Валюта установлен: %s", message.Text),
				message.UserID,
				removeMarkup,
			)

			if err != nil {
				return commands.NewError(err, false)
			}

			user, err := userStorage.Get(ctx, message.UserID)
			if err != nil {
				return commands.NewError(err, false)
			}
			user.Currency = message.Text
			err = userStorage.Update(ctx, user)

			if err != nil {
				return commands.NewError(err, false)
			}

			return nil
		},
	}
}
