package catalog

import (
	"context"
	"fmt"
	"uuid"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/converter"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/model"
)

func (r *Repository) loadCalculatorData(ctx context.Context) ([]model.CalculatorData, error) {
	pages, err := r.loadCalculatorPages(ctx)
	if err != nil {
		return nil, fmt.Errorf("load calculator pages: %w", err)
	}
	if len(pages) == 0 {
		return []model.CalculatorData{}, nil
	}
	calculators, err := r.loadCalculatorRows(ctx)
	if err != nil {
		return nil, fmt.Errorf("load calculator rows: %w", err)
	}
	byID := make(map[uuid.UUID]model.Calculator, len(calculators))
	for _, calculator := range calculators {
		byID[calculator.ID] = calculator
	}
	ids := make([]uuid.UUID, 0, len(pages))
	for _, page := range pages {
		ids = append(ids, page.ID)
	}
	blocks, err := r.loadCalculatorBlocks(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("load calculator blocks: %w", err)
	}
	byPage := make(map[uuid.UUID][]model.ContentBlock, len(pages))
	for _, block := range blocks {
		byPage[block.EntityID] = append(byPage[block.EntityID], block)
	}
	result := make([]model.CalculatorData, 0, len(pages))
	for _, page := range pages {
		reference, err := converter.CalculatorReferenceID(page.ID, byPage[page.ID])
		if err != nil {
			return nil, err
		}
		calculator, ok := byID[reference]
		if !ok {
			return nil, fmt.Errorf("page %s: referenced calculator %s not found", page.ID, reference)
		}
		result = append(result, model.CalculatorData{
			Calculator: calculator,
			Page:       page,
			Blocks:     byPage[page.ID],
		})
	}
	return result, nil
}

func (r *Repository) loadCalculatorBlocks(
	ctx context.Context,
	ids []uuid.UUID,
) ([]model.ContentBlock, error) {
	const query = `
		SELECT "Id", "CalculatorPageId", "Type", "HtmlContent", "Params", "Order"
		FROM public."CalculatorPageEmbedItems"
		WHERE "CalculatorPageId" = ANY($1::uuid[])
		ORDER BY "CalculatorPageId", "Order" NULLS LAST, "Id"
	`
	return r.loadContentBlocks(ctx, ids, query, "calculator")
}
