package catalog

import (
	"context"
	"fmt"
	"uuid"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/model"
)

func (r *Repository) loadArticleData(ctx context.Context) ([]model.ArticleData, error) {
	articles, err := r.loadArticleRows(ctx)
	if err != nil {
		return nil, fmt.Errorf("load article rows: %w", err)
	}
	if len(articles) == 0 {
		return []model.ArticleData{}, nil
	}

	ids := make([]uuid.UUID, 0, len(articles))
	settingsIDs := make([]uuid.UUID, 0, len(articles))
	for _, article := range articles {
		ids = append(ids, article.ID)
		if article.ContentSettingsID != nil {
			settingsIDs = append(settingsIDs, *article.ContentSettingsID)
		}
	}

	relations, err := r.loadContentRelations(ctx, settingsIDs)
	if err != nil {
		return nil, fmt.Errorf("load article relations: %w", err)
	}

	blocks, err := r.loadArticleBlocks(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("load article blocks: %w", err)
	}
	blocksByEntity := make(map[uuid.UUID][]model.ContentBlock, len(articles))
	for _, block := range blocks {
		blocksByEntity[block.EntityID] = append(blocksByEntity[block.EntityID], block)
	}

	result := make([]model.ArticleData, 0, len(articles))
	for _, article := range articles {
		data := model.ArticleData{
			Article: article,
			Blocks:  blocksByEntity[article.ID],
		}
		if article.ContentSettingsID != nil {
			settingsID := *article.ContentSettingsID
			data.Topics = relations.topicsBySettings[settingsID]
			data.Tags = relations.tagsBySettings[settingsID]
			data.Industries = relations.industriesBySettings[settingsID]
		}
		result = append(result, data)
	}

	return result, nil
}
