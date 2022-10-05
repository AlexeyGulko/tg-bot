package messages_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/commands"
	mocks "gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/mocks/messages"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/model/messages"
)

func Test_OnUnknownCommand_ShouldAnswerWithHelpMessage(t *testing.T) {
	ctrl := gomock.NewController(t)

	sender := mocks.NewMockMessageSender(ctrl)
	command := commands.NotFoundCommand(sender)
	sender.EXPECT().SendMessage("не знаю эту команду", int64(123), nil)
	model := messages.New()
	model.SetDefaultCommand(command)
	err := model.IncomingMessage(messages.Message{
		Text:   "some text",
		UserID: 123,
	})

	assert.NoError(t, err)
}
