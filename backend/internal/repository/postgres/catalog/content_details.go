package catalog

import (
	"context"
	"fmt"
	"uuid"

	"github.com/jackc/pgx/v5"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/model"
)

func (r *Repository) loadPodcastDetails(
	ctx context.Context,
	ids []uuid.UUID,
) ([]model.ContentDetail, error) {
	const query = `
		SELECT "Id", "PodcastId", "Title", "Description", "Order", "IsMain"
		FROM public."PodcastDetails"
		WHERE "PodcastId" = ANY($1::uuid[])
		ORDER BY "PodcastId", "Order" NULLS LAST, "Id"
	`
	return r.loadContentDetails(ctx, ids, query, "podcast")
}

func (r *Repository) loadWebinarDetails(
	ctx context.Context,
	ids []uuid.UUID,
) ([]model.ContentDetail, error) {
	const query = `
		SELECT "Id", "WebinarId", "Title", "Description", "Order", "IsMain"
		FROM public."WebinarInfoBindings"
		WHERE "WebinarId" = ANY($1::uuid[])
		ORDER BY "WebinarId", "Order" NULLS LAST, "Id"
	`
	return r.loadContentDetails(ctx, ids, query, "webinar")
}

func (r *Repository) loadContentDetails(
	ctx context.Context,
	ids []uuid.UUID,
	query, source string,
) ([]model.ContentDetail, error) {
	if len(ids) == 0 {
		return []model.ContentDetail{}, nil
	}
	rows, err := r.executor(ctx).Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("query %s details: %w", source, err)
	}
	details, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.ContentDetail, error) {
		var detail model.ContentDetail
		err := detail.Scan(row)
		return detail, err
	})
	if err != nil {
		return nil, fmt.Errorf("scan %s details: %w", source, err)
	}
	return details, nil
}
