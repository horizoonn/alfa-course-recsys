package converter

import (
	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/model"
)

func topicLabels(topics []model.ContentTopic) []domain.Label {
	result := make([]domain.Label, 0, len(topics))
	for _, topic := range topics {
		if topic.IsDeleted {
			continue
		}
		result = append(result, domain.Label{
			ID:    topic.TopicID,
			Title: normalizeText(topic.Title),
		})
	}
	return result
}

func tagLabels(tags []model.ContentTag) []domain.Label {
	result := make([]domain.Label, 0, len(tags))
	for _, tag := range tags {
		if tag.IsDeleted {
			continue
		}
		result = append(result, domain.Label{
			ID:    tag.TagID,
			Title: normalizeText(tag.Title),
		})
	}
	return result
}

func additionalLabels(industries []model.ContentIndustry) []domain.Label {
	result := make([]domain.Label, 0, len(industries))
	for _, industry := range industries {
		if industry.IsDeleted {
			continue
		}
		result = append(result, domain.Label{
			ID:    industry.IndustryID,
			Title: normalizeText(industry.Title),
		})
	}
	return result
}
