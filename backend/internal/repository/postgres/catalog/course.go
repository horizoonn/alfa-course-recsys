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

var courseEntityTypeID = uuid.MustParse("8b440b1a-00c3-423a-9bae-3376475eb203")

func (r *Repository) ListCourses(ctx context.Context) ([]domain.Material, error) {
	data, err := r.loadCourseData(ctx)
	if err != nil {
		return nil, fmt.Errorf("load courses: %w", err)
	}

	materials := make([]domain.Material, 0, len(data))
	for _, item := range data {
		material, err := converter.CourseToMaterial(item)
		if err != nil {
			return nil, fmt.Errorf("convert courses: %w", err)
		}
		materials = append(materials, material)
	}
	return materials, nil
}

func (r *Repository) loadCourseRows(ctx context.Context) ([]model.Course, error) {
	const query = `
		SELECT
			course."Id",
			course."StatusId",
			course."TypeId",
			course."Title",
			course."Description",
			course."DescriptionAdditional",
			course."TitleTag",
			course."DescriptionTag",
			course."TitleNavigation",
			course."Target",
			course."Duration",
			course."Url",
			course."IsDeleted",
			course."IsClosed",
			course."Published",
			course."Modified",
			course."Rating",
			settings."Id",
			settings."StatusId",
			settings."TypeId",
			settings."IsDeleted",
			EXISTS (
				SELECT 1
				FROM public."CourseSubscriptions" AS binding
				JOIN public."Subscription" AS subscription ON subscription."Id" = binding."SubscriptionId"
				WHERE binding."CourseId" = course."Id"
					AND NOT subscription."IsDeleted"
			),
			EXISTS (
				SELECT 1
				FROM public."CourseTariffBindings" AS tariff
				WHERE tariff."CourseId" = course."Id"
			)
		FROM public."Courses" AS course
		LEFT JOIN public."ContentSettings" AS settings
			ON settings."EntityId" = course."Id"
			AND settings."EntityTypeId" = $1::uuid
		ORDER BY course."Id"
	`

	rows, err := r.executor(ctx).Query(ctx, query, courseEntityTypeID)
	if err != nil {
		return nil, fmt.Errorf("query courses: %w", err)
	}

	courses, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.Course, error) {
		var course model.Course
		err := course.Scan(row)
		return course, err
	})
	if err != nil {
		return nil, fmt.Errorf("scan courses: %w", err)
	}

	return courses, nil
}

func (r *Repository) loadCourseInfo(
	ctx context.Context,
	courseIDs []uuid.UUID,
) ([]model.CourseInfo, error) {
	if len(courseIDs) == 0 {
		return []model.CourseInfo{}, nil
	}

	const query = `
		SELECT
			info."Id",
			info."CourseId",
			info."CourseInfoTypeId",
			info_type."Title",
			info_type."IsDeleted",
			info."Title",
			info."Description",
			info."Order",
			info."IsMain",
			info."CourseTariffId"
		FROM public."CourseInfoBindings" AS info
		JOIN public."CourseInfoTypes" AS info_type ON info_type."Id" = info."CourseInfoTypeId"
		WHERE info."CourseId" = ANY($1::uuid[])
		ORDER BY info."CourseId", info."CourseInfoTypeId", info."Order" NULLS LAST, info."Id"
	`

	rows, err := r.executor(ctx).Query(ctx, query, courseIDs)
	if err != nil {
		return nil, fmt.Errorf("query course info: %w", err)
	}

	info, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.CourseInfo, error) {
		var item model.CourseInfo
		err := item.Scan(row)
		return item, err
	})
	if err != nil {
		return nil, fmt.Errorf("scan course info: %w", err)
	}

	return info, nil
}
