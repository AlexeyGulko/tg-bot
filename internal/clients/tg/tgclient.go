package tg

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/logger"
	"go.uber.org/zap"
)

type Client struct {
	client *tgbotapi.BotAPI
}

type TokenGetter interface {
	Token() string
}

type Message interface {
	IncomingMessage(ctx context.Context, msg *dto.Message) error
}

func New(getter TokenGetter) (*Client, error) {
	client, err := tgbotapi.NewBotAPI(getter.Token())

	if err != nil {
		return nil, errors.Wrap(err, "NewBotAPI")
	}

	return &Client{client: client}, nil
}

func (c Client) SendMessage(ctx context.Context, text string, userID int64, markup interface{}) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "send message to tg")
	defer span.Finish()
	msg := tgbotapi.NewMessage(userID, text)
	msg.ReplyMarkup = markup
	_, err := c.client.Send(msg)

	if err != nil {
		return errors.Wrap(err, "client.send")
	}

	return nil
}

func (c *Client) ListenUpdates(ctx context.Context, msgModel Message) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := c.client.GetUpdatesChan(u)

	logger.Info("listening for messages")

	for {
		select {
		case <-ctx.Done():
			c.client.StopReceivingUpdates()
			logger.Info("fetch message stopped")
			return
		case update := <-updates:
			if update.Message != nil { // If we got a message
				err := msgModel.IncomingMessage(
					ctx,
					&dto.Message{
						Text:   update.Message.Text,
						UserID: update.Message.From.ID,
					},
				)
				if err != nil {
					logger.Error("error processing message:", zap.Error(err))
				}
			}
		}
	}
}
