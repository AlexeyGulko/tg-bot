package command

import (
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/model/messages"
)

type Storage struct {
	store map[int64]messages.Command
}

func (s *Storage) Add(UserID int64, command messages.Command) {
	s.store[UserID] = command
}

func (s *Storage) Get(UserID int64) (messages.Command, bool) {
	val, ok := s.store[UserID]
	return val, ok
}

func (s *Storage) Delete(UserID int64) {
	delete(s.store, UserID)
}

func NewStorage() *Storage {
	return &Storage{store: make(map[int64]messages.Command)}
}
