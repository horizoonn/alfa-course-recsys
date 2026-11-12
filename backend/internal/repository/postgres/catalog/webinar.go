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

var webinarEntityTypeID = uuid.MustParse("ece4acf3-5dd3-47b9-9f08-99c0d6203cc7")

func (r *Repository) ListWebinars(ctx context.Context) ([]domain.Material, error) {
	data, err := r.loadWebinarData(ctx)
	if err != nil {
		return nil, fmt.Errorf("load webinars: %w", err)
	}
	materials := make([]domain.Material, 0, len(data))
	for _, item := range data {
		material, err := converter.WebinarToMaterial(item)
		if err != nil {
			return nil, fmt.Errorf("convert webinars: %w", err)
		}
		materials = append(materials, material)
	}
	return materials, nil
}

func (r *Repository) loadWebinarRows(ctx context.Context) ([]model.Webinar, error) {
	const query = `
		SELECT
			webinar."Id",
			webinar."StatusId",
			webinar."TypeId",
			webinar."Title",
			webinar."Description",
			webinar."DescriptionExt",
			webinar."TitleTag",
			webinar."DescriptionTag",
			webinar."TitleNavigation",
			webinar."Duration",
			webinar."Url",
			webinar."IsDeleted",
			webinar."IsClosed",
			webinar."Modified",
			webinar."Rating",
			webinar."StartsAt",
			webinar."CompletedAt",
			webinar."IsRegistrationRequired",
			webinar."IsPasswordRequired",
			webinar."IsModerationRequired",
			settings."Id",
			settings."StatusId",
			settings."TypeId",
			settings."IsDeleted",
			settings."Published"
		FROM public."Webinars" AS webinar
		LEFT JOIN public."ContentSettings" AS settings
			ON settings."EntityId" = webinar."Id" AND settings."EntityTypeId" = $1::uuid
		ORDER BY webinar."Id"
	`
	rows, err := r.executor(ctx).Query(ctx, query, webinarEntityTypeID)
	if err != nil {
		return nil, fmt.Errorf("query webinars: %w", err)
	}
	result, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.Webinar, error) {
		var item model.Webinar
		err := item.Scan(row)
		return item, err
	})
	if err != nil {
		return nil, fmt.Errorf("scan webinars: %w", err)
	}
	return result, nil
}
