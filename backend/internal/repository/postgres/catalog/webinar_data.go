package catalog

import (
	"context"
	"fmt"
	"uuid"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/model"
)

func (r *Repository) loadWebinarData(ctx context.Context) ([]model.WebinarData, error) {
	items, err := r.loadWebinarRows(ctx)
	if err != nil {
		return nil, fmt.Errorf("load webinar rows: %w", err)
	}
	if len(items) == 0 {
		return []model.WebinarData{}, nil
	}
	ids := make([]uuid.UUID, 0, len(items))
	settingsIDs := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
		if item.ContentSettingsID != nil {
			settingsIDs = append(settingsIDs, *item.ContentSettingsID)
		}
	}
	relations, err := r.loadContentRelations(ctx, settingsIDs)
	if err != nil {
		return nil, fmt.Errorf("load webinar relations: %w", err)
	}
	details, err := r.loadWebinarDetails(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("load webinar details: %w", err)
	}
	byEntity := make(map[uuid.UUID][]model.ContentDetail, len(items))
	for _, detail := range details {
		byEntity[detail.EntityID] = append(byEntity[detail.EntityID], detail)
	}
	result := make([]model.WebinarData, 0, len(items))
	for _, item := range items {
		data := model.WebinarData{
			Webinar: item,
			Details: byEntity[item.ID],
		}
		if item.ContentSettingsID != nil {
			id := *item.ContentSettingsID
			data.Topics = relations.topicsBySettings[id]
			data.Tags = relations.tagsBySettings[id]
			data.Industries = relations.industriesBySettings[id]
		}
		result = append(result, data)
	}
	return result, nil
}
