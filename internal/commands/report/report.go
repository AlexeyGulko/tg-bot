package report

import (
	"context"
	"time"

	"github.com/Shopify/sarama"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	api "gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/gen/proto/go"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/helpers"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/model/messages"
	"google.golang.org/protobuf/proto"
)

const (
	year  = "год"
	month = "месяц"
	week  = "неделя"
)

var keyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(year),
		tgbotapi.NewKeyboardButton(month),
		tgbotapi.NewKeyboardButton(week),
	),
)

type CommandSequence struct {
	tgClient        commands.MessageSender
	spendStorage    commands.SpendingStorage
	config          commands.Config
	userStorage     commands.UserStorage
	currencyService commands.CurrencyService
	queueCh         chan<- *sarama.ProducerMessage
}

func New(
	tgClient commands.MessageSender,
	spendStorage commands.SpendingStorage,
	config commands.Config,
	userStorage commands.UserStorage,
	currencyService commands.CurrencyService,
	queueCh chan<- *sarama.ProducerMessage,
) commands.Command {
	comm := &CommandSequence{
		tgClient:        tgClient,
		spendStorage:    spendStorage,
		config:          config,
		userStorage:     userStorage,
		currencyService: currencyService,
		queueCh:         queueCh,
	}

	return commands.Command{CallBack: comm.Menu}.SetNext(commands.Command{CallBack: comm.Report})
}

func (c *CommandSequence) Menu(ctx context.Context, message *dto.Message) messages.CommandError {
	message.Command = "report_menu"
	err := c.tgClient.SendMessage(ctx, "Выбери период", message.UserID, keyboard)
	if err != nil {
		return commands.NewError(err, false)
	}

	return nil
}

func (c *CommandSequence) Report(ctx context.Context, message *dto.Message) messages.CommandError {
	message.Command = "report_input"
	start, periodName, err := getPeriod(message.Text)
	if err != nil {
		err = c.tgClient.SendMessage(ctx, "Выбери период из предложенных на клавиатуре", message.UserID, keyboard)
		return commands.NewError(err, true)
	}
	user, err := c.userStorage.GetOrCreate(ctx, dto.User{TgID: message.UserID, Currency: c.config.DefaultCurrency()})
	if err != nil {
		return commands.NewError(err, false)
	}

	loc, err := time.LoadLocation("GMT")
	if err != nil {
		return commands.NewError(err, false)
	}

	binaryId, err := user.ID.MarshalBinary()
	if err != nil {
		return commands.NewError(err, false)
	}

	protoMsg := api.GenerateReportRequest{
		UserId:   binaryId,
		Currency: user.Currency,
		Start:    helpers.StartOfDay(start, loc).Unix(),
		End:      helpers.StartOfDay(time.Now(), loc).Unix(),
		Period:   periodName,
	}

	protoMsgBytes, err := proto.Marshal(&protoMsg)
	if err != nil {
		return commands.NewError(err, false)
	}

	msg := sarama.ProducerMessage{
		Topic: "reporter",
		Key:   sarama.StringEncoder("spending_report"),
		Value: sarama.ByteEncoder(protoMsgBytes),
	}
	removeMarkup := tgbotapi.NewRemoveKeyboard(true)
	err = c.tgClient.SendMessage(ctx, "отчет генерируется...", message.UserID, removeMarkup)
	c.queueCh <- &msg

	if err != nil {
		return commands.NewError(err, false)
	}
	return nil
}

func getPeriod(key string) (time.Time, string, error) {
	period := time.Now()
	var name string
	var err error

	switch key {
	case "год":
		period = period.AddDate(-1, 0, 0)
		name = "год"
	case "месяц":
		period = period.AddDate(0, -1, 0)
		name = "месяц"
	case "неделя":
		period = period.AddDate(0, 0, -7)
		name = "неделю"
	default:
		err = errors.New("wrong period key")
	}

	return period, name, err
}
