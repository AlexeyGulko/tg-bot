package http

import (
	"context"

	"github.com/google/uuid"
	api "gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/gen/proto/go"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/logger"
)

type Server struct {
	api.UnimplementedReporterServer
	tg      MessageSender
	userSrv UserStorage
}

type UserStorage interface {
	GetOrCreate(context.Context, dto.User) (*dto.User, error)
}

type MessageSender interface {
	SendMessage(ctx context.Context, text string, userId int64, markup interface{}) error
}

func NewServer(tg MessageSender, storage UserStorage) *Server {
	return &Server{tg: tg, userSrv: storage}
}

func (s Server) SendReport(ctx context.Context, in *api.ReportRequest) (*api.ReportResponse, error) {
	logger.Info(in.GetReport())

	id, err := uuid.FromBytes(in.UserId)
	if err != nil {
		return nil, err
	}

	searchUser := dto.User{ID: id}

	user, err := s.userSrv.GetOrCreate(ctx, searchUser)
	if err != nil {
		return nil, err
	}

	if err := s.tg.SendMessage(ctx, in.GetReport(), user.TgID, nil); err != nil {
		return nil, err
	}

	return &api.ReportResponse{}, nil
}
