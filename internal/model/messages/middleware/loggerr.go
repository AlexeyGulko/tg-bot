package middleware

import (
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/clients/tg"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/logger"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type Logger struct {
	next tg.Message
}

func NewLogger(next tg.Message) *Logger {
	return &Logger{next: next}
}

func (m *Logger) IncomingMessage(ctx context.Context, msg *dto.Message) error {
	logger.Info("input from tg bot", zap.String("input", msg.Text))
	err := m.next.IncomingMessage(ctx, msg)
	if err != nil {
		return err
	}
	return nil
}
