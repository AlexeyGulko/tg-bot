package month_budget

import (
	"context"
	"fmt"
	"time"

	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/model/messages"
)

func Menu(
	tgClient commands.MessageSender,
	userStorage commands.UserStorage,
	config commands.Config,
	curSvc commands.CurrencyService,
) *commands.Command {
	return &commands.Command{
		CallBack: func(ctx context.Context, message *dto.Message) messages.CommandError {
			message.Command = "budget_menu"
			user, err := userStorage.GetOrCreate(ctx, dto.User{TgID: message.UserID, Currency: config.DefaultCurrency()})
			if err != nil {
				return commands.NewError(err, false)
			}
			budget := user.MonthBudget
			if user.Currency != config.DefaultCurrency() {
				budget, err = curSvc.ConvertTo(ctx, user.Currency, user.MonthBudget, time.Now())

				if err != nil {
					return commands.NewError(err, true)
				}
			}

			err = tgClient.SendMessage(
				ctx,
				fmt.Sprintf("Текущий бюдже: %s %s\nустанови бюджет", budget.RoundBank(2), user.Currency),
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

func Input(
	tgClient commands.MessageSender,
	userStorage commands.UserStorage,
	config commands.Config,
	curSvc commands.CurrencyService,
) commands.Command {
	return commands.Command{
		CallBack: func(ctx context.Context, message *dto.Message) messages.CommandError {
			message.Command = "budget_input"
			user, err := userStorage.Get(ctx, message.UserID)
			if err != nil {
				return commands.NewError(err, false)
			}

			budget, err := commands.ParseDigitInput(message.Text)
			if err != nil {
				err := tgClient.SendMessage(ctx, "Неправильный формат бюджета попробуй ещё", message.UserID, nil)
				if err != nil {
					return commands.NewError(err, false)
				}
				return commands.NewError(err, true)
			}

			if user.Currency != config.DefaultCurrency() {
				budget, err = curSvc.ConvertFrom(ctx, user.Currency, budget, time.Now())

				if err != nil {
					return commands.NewError(err, true)
				}
			}

			err = tgClient.SendMessage(
				ctx,
				fmt.Sprintf("Бюджет на месяц установлен: %s %s", message.Text, user.Currency),
				message.UserID,
				nil,
			)

			if err != nil {
				return commands.NewError(err, false)
			}

			user.MonthBudget = budget
			err = userStorage.Update(ctx, user)

			if err != nil {
				return commands.NewError(err, false)
			}
			return nil
		},
	}
}
