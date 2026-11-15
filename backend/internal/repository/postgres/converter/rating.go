package converter

import "github.com/horizoonn/alfa-course-recsys/backend/internal/domain"

func ratingSummary(average *float64, count *int64) *domain.RatingSummary {
	if average == nil {
		return nil
	}
	return &domain.RatingSummary{
		Average: *average,
		Count:   count,
	}
}
