package currency

import (
	"context"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/helpers"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/model/messages"
)

func Menu(tgClient commands.MessageSender, config commands.Config, userStorage commands.UserStorage) *commands.Command {

	currencies := config.Currencies()

	keyboard := helpers.CreateMarkupMenu(currencies, 2)

	return &commands.Command{
		CallBack: func(ctx context.Context, message dto.Message) messages.CommandError {

			user, _ := userStorage.Get(message.UserID)
			err := tgClient.SendMessage(fmt.Sprintf("Текущая валюта: %s\nВыбери валюту", user.Currency), message.UserID, keyboard)
			if err != nil {
				log.Printf("%s", err)
			}

			return nil
		},
	}
}

func Input(tgClient commands.MessageSender, config commands.Config, userStorage commands.UserStorage) commands.Command {
	return commands.Command{
		CallBack: func(ctx context.Context, message dto.Message) messages.CommandError {
			currencies := config.Currencies()

			currMap := make(map[string]interface{}, len(currencies))
			for _, v := range currencies {
				currMap[v] = struct{}{}
			}

			_, has := currMap[message.Text]

			if !has {
				keyboard := helpers.CreateMarkupMenu(currencies, 2)
				err := tgClient.SendMessage("Выбери валюту", message.UserID, keyboard)
				log.Printf("%s", err)
				return &commands.CommandError{Retry: true}
			}

			removeMarkup := tgbotapi.NewRemoveKeyboard(true)

			err := tgClient.SendMessage(
				fmt.Sprintf("Валюта установлен: %s", message.Text),
				message.UserID,
				removeMarkup,
			)

			user, _ := userStorage.Get(message.UserID)
			user.Currency = message.Text
			userStorage.Add(user)

			if err != nil {
				log.Printf("%s", err)
			}
			return nil
		},
	}
}
