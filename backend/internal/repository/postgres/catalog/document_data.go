package catalog

import (
	"context"
	"fmt"
	"uuid"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/model"
)

func (r *Repository) loadDocumentData(ctx context.Context) ([]model.DocumentData, error) {
	documents, err := r.loadDocumentRows(ctx)
	if err != nil {
		return nil, fmt.Errorf("load document rows: %w", err)
	}
	if len(documents) == 0 {
		return []model.DocumentData{}, nil
	}

	ids := make([]uuid.UUID, 0, len(documents))
	settingsIDs := make([]uuid.UUID, 0, len(documents))
	for _, document := range documents {
		ids = append(ids, document.ID)
		if document.ContentSettingsID != nil {
			settingsIDs = append(settingsIDs, *document.ContentSettingsID)
		}
	}

	relations, err := r.loadContentRelations(ctx, settingsIDs)
	if err != nil {
		return nil, fmt.Errorf("load document relations: %w", err)
	}

	blocks, err := r.loadDocumentBlocks(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("load document blocks: %w", err)
	}
	blocksByEntity := make(map[uuid.UUID][]model.ContentBlock, len(documents))
	for _, block := range blocks {
		blocksByEntity[block.EntityID] = append(blocksByEntity[block.EntityID], block)
	}

	result := make([]model.DocumentData, 0, len(documents))
	for _, document := range documents {
		data := model.DocumentData{
			Document: document,
			Blocks:   blocksByEntity[document.ID],
		}
		if document.ContentSettingsID != nil {
			settingsID := *document.ContentSettingsID
			data.Topics = relations.topicsBySettings[settingsID]
			data.Tags = relations.tagsBySettings[settingsID]
			data.Industries = relations.industriesBySettings[settingsID]
		}
		result = append(result, data)
	}

	return result, nil
}
