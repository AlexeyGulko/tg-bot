package spending

import (
	"time"

	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
)

type Storage struct {
	store map[int64][]dto.Spending
}

func NewStorage() *Storage {
	return &Storage{store: make(map[int64][]dto.Spending)}
}

func (s *Storage) Add(UserID int64, model dto.Spending) {
	s.store[UserID] = append(s.store[UserID], model)
}

func (s *Storage) Get(UserID int64) ([]dto.Spending, bool) {
	if val, ok := s.store[UserID]; ok {
		return val, true
	}
	return nil, false
}

func (s *Storage) GetReportByCategory(UserID int64, date time.Time) map[string][]dto.Spending {
	spending, _ := s.Get(UserID)

	res := make(map[string][]dto.Spending)
	for _, v := range spending {
		if v.Date.After(date) {
			res[v.Category] = append(res[v.Category], v)
		}
	}

	return res
}
