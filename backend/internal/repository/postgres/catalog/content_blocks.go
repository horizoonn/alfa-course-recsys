package catalog

import (
	"context"
	"fmt"
	"uuid"

	"github.com/jackc/pgx/v5"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/model"
)

func (r *Repository) loadArticleBlocks(
	ctx context.Context,
	ids []uuid.UUID,
) ([]model.ContentBlock, error) {
	const query = `
		SELECT "Id", "ArticleId", "Type", "HtmlContent", "Params", "Order"
		FROM public."ArticleEmbedItems"
		WHERE "ArticleId" = ANY($1::uuid[])
		ORDER BY "ArticleId", "Order" NULLS LAST, "Id"
	`
	return r.loadContentBlocks(ctx, ids, query, "article")
}

func (r *Repository) loadDocumentBlocks(
	ctx context.Context,
	ids []uuid.UUID,
) ([]model.ContentBlock, error) {
	const query = `
		SELECT "Id", "DocumentId", "Type", "HtmlContent", "Params", "Order"
		FROM public."DocumentEmbedItems"
		WHERE "DocumentId" = ANY($1::uuid[])
		ORDER BY "DocumentId", "Order" NULLS LAST, "Id"
	`
	return r.loadContentBlocks(ctx, ids, query, "document")
}

func (r *Repository) loadContentBlocks(
	ctx context.Context,
	ids []uuid.UUID,
	query, source string,
) ([]model.ContentBlock, error) {
	if len(ids) == 0 {
		return []model.ContentBlock{}, nil
	}
	rows, err := r.executor(ctx).Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("query %s blocks: %w", source, err)
	}
	blocks, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.ContentBlock, error) {
		var block model.ContentBlock
		err := block.Scan(row)
		return block, err
	})
	if err != nil {
		return nil, fmt.Errorf("scan %s blocks: %w", source, err)
	}
	return blocks, nil
}
