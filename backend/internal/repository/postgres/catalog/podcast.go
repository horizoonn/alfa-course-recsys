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

var podcastEntityTypeID = uuid.MustParse("142c23e2-bf71-4a73-b12e-6b1b5e4d47cc")

func (r *Repository) ListPodcasts(ctx context.Context) ([]domain.Material, error) {
	data, err := r.loadPodcastData(ctx)
	if err != nil {
		return nil, fmt.Errorf("load podcasts: %w", err)
	}
	materials := make([]domain.Material, 0, len(data))
	for _, item := range data {
		material, err := converter.PodcastToMaterial(item)
		if err != nil {
			return nil, fmt.Errorf("convert podcasts: %w", err)
		}
		materials = append(materials, material)
	}
	return materials, nil
}

func (r *Repository) loadPodcastRows(ctx context.Context) ([]model.Podcast, error) {
	const query = `
		SELECT
			podcast."Id",
			podcast."StatusId",
			podcast."TypeId",
			podcast."Title",
			podcast."Description",
			podcast."DescriptionExt",
			podcast."TitleTag",
			podcast."DescriptionTag",
			podcast."TitleNavigation",
			podcast."Duration",
			podcast."Url",
			podcast."IsDeleted",
			podcast."IsClosed",
			podcast."Published",
			podcast."Modified",
			podcast."Rating",
			settings."Id",
			settings."StatusId",
			settings."TypeId",
			settings."IsDeleted",
			EXISTS (
				SELECT 1 FROM public."PodcastSubscriptions" binding
				JOIN public."Subscription" subscription ON subscription."Id" = binding."SubscriptionId"
				WHERE binding."PodcastId" = podcast."Id" AND NOT subscription."IsDeleted"
			)
		FROM public."Podcasts" AS podcast
		LEFT JOIN public."ContentSettings" AS settings
			ON settings."EntityId" = podcast."Id" AND settings."EntityTypeId" = $1::uuid
		ORDER BY podcast."Id"
	`
	rows, err := r.executor(ctx).Query(ctx, query, podcastEntityTypeID)
	if err != nil {
		return nil, fmt.Errorf("query podcasts: %w", err)
	}
	result, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.Podcast, error) {
		var item model.Podcast
		err := item.Scan(row)
		return item, err
	})
	if err != nil {
		return nil, fmt.Errorf("scan podcasts: %w", err)
	}
	return result, nil
}
