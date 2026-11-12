package catalog

import (
	"context"
	"fmt"
	"uuid"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/model"
)

func (r *Repository) loadPodcastData(ctx context.Context) ([]model.PodcastData, error) {
	items, err := r.loadPodcastRows(ctx)
	if err != nil {
		return nil, fmt.Errorf("load podcast rows: %w", err)
	}
	if len(items) == 0 {
		return []model.PodcastData{}, nil
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
		return nil, fmt.Errorf("load podcast relations: %w", err)
	}
	details, err := r.loadPodcastDetails(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("load podcast details: %w", err)
	}
	byEntity := make(map[uuid.UUID][]model.ContentDetail, len(items))
	for _, detail := range details {
		byEntity[detail.EntityID] = append(byEntity[detail.EntityID], detail)
	}
	result := make([]model.PodcastData, 0, len(items))
	for _, item := range items {
		data := model.PodcastData{
			Podcast: item,
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
