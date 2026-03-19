package prompt

import (
	"fmt"
	"time"
)

type Store interface {
	Create(template Template) error
	List() ([]Template, error)
}

type Service struct {
	store Store
	now   func() time.Time
	idGen func() string
}

func NewService(store Store) *Service {
	return &Service{
		store: store,
		now:   time.Now().UTC,
		idGen: func() string {
			return fmt.Sprintf("prompt-%d", time.Now().UTC().UnixNano())
		},
	}
}

func (s *Service) Create(input CreateInput) (Template, error) {
	now := s.now()
	record := Template{
		ID:        s.idGen(),
		Name:      input.Name,
		Content:   input.Content,
		Tags:      append([]string(nil), input.Tags...),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.store.Create(record); err != nil {
		return Template{}, fmt.Errorf("create prompt template: %w", err)
	}
	return record, nil
}

func (s *Service) List() ([]Template, error) {
	templates, err := s.store.List()
	if err != nil {
		return nil, fmt.Errorf("list prompt templates: %w", err)
	}
	return templates, nil
}
