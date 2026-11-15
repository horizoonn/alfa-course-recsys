package catalog

import (
	"context"
	"fmt"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
)

func (s *Service) Load(ctx context.Context) ([]domain.Material, error) {
	materials := []domain.Material{}
	err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		seen := make(map[domain.MaterialKey]struct{})
		for _, source := range []struct {
			name string
			list func(context.Context) ([]domain.Material, error)
		}{
			{"articles", s.repository.ListArticles},
			{"courses", s.repository.ListCourses},
			{"documents", s.repository.ListDocuments},
			{"terms", s.repository.ListTerms},
			{"calculators", s.repository.ListCalculators},
			{"podcasts", s.repository.ListPodcasts},
			{"webinars", s.repository.ListWebinars},
		} {
			items, err := source.list(txCtx)
			if err != nil {
				return fmt.Errorf("list %s: %w", source.name, err)
			}
			for _, item := range items {
				if _, exists := seen[item.Key]; exists {
					return fmt.Errorf("duplicate material key %s", item.Key)
				}
				seen[item.Key] = struct{}{}
				materials = append(materials, item)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("load catalog: %w", err)
	}
	return materials, nil
}
