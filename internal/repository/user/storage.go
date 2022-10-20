package user

import "gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"

type Storage struct {
	store map[int64]dto.User
}

func New() *Storage {
	return &Storage{store: make(map[int64]dto.User)}
}

func (s *Storage) Add(model dto.User) {
	s.store[model.ID] = model
}

func (s *Storage) Get(UserID int64) (dto.User, bool) {
	val, has := s.store[UserID]
	return val, has
}
