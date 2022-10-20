package helpers

import (
	"math/rand"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func RandomDate() time.Time {
	min := time.Date(1970, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(2070, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	delta := max - min

	sec := rand.Int63n(delta) + min
	return time.Unix(sec, 0)
}

func CreateMarkupMenu(data []string, perRow int64) tgbotapi.ReplyKeyboardMarkup {
	var keyboard [][]tgbotapi.KeyboardButton
	row := 0
	keyboard = append(keyboard, []tgbotapi.KeyboardButton{})
	for i, v := range data {
		keyboard[row] = append(keyboard[row], tgbotapi.NewKeyboardButton(v))

		if perRow == 0 {
			continue
		}

		if int64(i+1)%perRow == 0 {
			keyboard = append(keyboard, []tgbotapi.KeyboardButton{})
			row++
		}
	}

	return tgbotapi.NewReplyKeyboard(keyboard...)
}
