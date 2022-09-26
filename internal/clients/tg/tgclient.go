package tg

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/model/messages"
	"log"
)

type Client struct {
	client *tgbotapi.BotAPI
}

type TokenGetter interface {
	Token() string
}

func New(getter TokenGetter) (*Client, error) {
	client, err := tgbotapi.NewBotAPI(getter.Token())

	if err != nil {
		return nil, errors.Wrap(err, "NewBotAPI")
	}

	return &Client{client: client}, nil
}

func (c Client) SendMessage(text string, userID int64) error {
	_, err := c.client.Send(tgbotapi.NewMessage(userID, text))

	if err != nil {
		return errors.Wrap(err, "client.send")
	}

	return nil
}

func (c *Client) ListenUpdates(msgModel *messages.Model) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := c.client.GetUpdatesChan(u)

	log.Println("listening for messages")

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			err := msgModel.IncomingMessage(messages.Message{
				Text:   update.Message.Text,
				UserID: update.Message.From.ID,
			})
			if err != nil {
				log.Println("error processing message:", err)
			}
		}
	}
}
