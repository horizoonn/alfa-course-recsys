package catalog

import (
	"context"
	"fmt"
	"uuid"

	"github.com/jackc/pgx/v5"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/converter"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/model"
)

func (r *Repository) ListTerms(ctx context.Context) ([]domain.Material, error) {
	data, err := r.loadTermData(ctx)
	if err != nil {
		return nil, fmt.Errorf("load terms: %w", err)
	}
	materials := make([]domain.Material, 0, len(data))
	for _, term := range data {
		material, err := converter.TermToMaterial(term)
		if err != nil {
			return nil, fmt.Errorf("convert terms: %w", err)
		}
		materials = append(materials, material)
	}
	return materials, nil
}

func (r *Repository) loadTermRows(ctx context.Context) ([]model.Term, error) {
	const query = `
		SELECT
			term."Id",
			term."StatusId",
			term."Title",
			term."Description",
			term."TitleTag",
			term."DescriptionTag",
			term."TitleNavigation",
			term."Letter",
			term."Duration",
			term."Url",
			term."IsDeleted",
			term."Published",
			term."Modified"
		FROM public."Glossary" AS term
		ORDER BY term."Id"
	`
	rows, err := r.executor(ctx).Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query terms: %w", err)
	}
	terms, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.Term, error) {
		var term model.Term
		err := term.Scan(row)
		return term, err
	})
	if err != nil {
		return nil, fmt.Errorf("scan terms: %w", err)
	}
	return terms, nil
}

func (r *Repository) loadTermBlocks(
	ctx context.Context,
	termIDs []uuid.UUID,
) ([]model.ContentBlock, error) {
	if len(termIDs) == 0 {
		return []model.ContentBlock{}, nil
	}
	const query = `
		SELECT
			"Id",
			"GlossaryId",
			"Type",
			"HtmlContent",
			"Params",
			"Order"
		FROM public."GlossaryEmbedItems"
		WHERE "GlossaryId" = ANY($1::uuid[])
		ORDER BY "GlossaryId", "Order" NULLS LAST, "Id"
	`
	return r.loadContentBlocks(ctx, termIDs, query, "term")
}
