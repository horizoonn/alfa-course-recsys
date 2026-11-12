package catalog

import (
	"context"
	"fmt"
	"uuid"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/model"
)

func (r *Repository) loadTermData(ctx context.Context) ([]model.TermData, error) {
	terms, err := r.loadTermRows(ctx)
	if err != nil {
		return nil, fmt.Errorf("load term rows: %w", err)
	}
	if len(terms) == 0 {
		return []model.TermData{}, nil
	}
	ids := make([]uuid.UUID, 0, len(terms))
	for _, term := range terms {
		ids = append(ids, term.ID)
	}
	blocks, err := r.loadTermBlocks(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("load term blocks: %w", err)
	}
	byTerm := make(map[uuid.UUID][]model.ContentBlock, len(terms))
	for _, block := range blocks {
		byTerm[block.EntityID] = append(byTerm[block.EntityID], block)
	}
	result := make([]model.TermData, 0, len(terms))
	for _, term := range terms {
		result = append(result, model.TermData{
			Term:   term,
			Blocks: byTerm[term.ID],
		})
	}
	return result, nil
}
