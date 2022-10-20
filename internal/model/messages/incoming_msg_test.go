package messages_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
	mock_messages "gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/mocks/messages"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/model/messages"
)

func Test_OnUnknownCommand_ShouldAnswerWithHelpMessage(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	comm := mock_messages.NewMockCommand(ctrl)
	storage := mock_messages.NewMockStorage(ctrl)
	model := messages.New(storage)
	model.SetDefaultCommand(comm)

	comm.EXPECT().Execute(
		ctx,
		dto.Message{
			Text:   "some text",
			UserID: 123,
		}).Return(nil)

	storage.EXPECT().Get(int64(123)).Return(comm, false)

	err := model.IncomingMessage(
		ctx,
		dto.Message{
			Text:   "some text",
			UserID: 123,
		})

	assert.NoError(t, err)
}
