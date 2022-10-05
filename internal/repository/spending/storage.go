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
	println(UserID, model.Amount, model.Category)
	s.store[UserID] = append(s.store[UserID], model)
}

func (s *Storage) Get(UserID int64) ([]dto.Spending, bool) {
	if val, ok := s.store[UserID]; ok {
		return val, true
	}
	return nil, false
}

func (s *Storage) GetReportByCategory(UserID int64, date time.Time) map[string]int64 {
	spending, _ := s.Get(UserID)

	res := make(map[string]int64)
	for _, v := range spending {
		if v.Date.After(date) {
			res[v.Category] += v.Amount
		}
	}

	return res
}
