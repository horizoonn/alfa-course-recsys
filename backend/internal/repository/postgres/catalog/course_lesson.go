package catalog

import (
	"context"
	"fmt"
	"uuid"

	"github.com/jackc/pgx/v5"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/model"
)

var lessonRatingEntityTypeID = uuid.MustParse("8b440b1a-00c3-423a-9bae-3376475eb203")

func (r *Repository) loadCourseLessons(
	ctx context.Context,
	courseIDs []uuid.UUID,
) ([]model.CourseLesson, error) {
	if len(courseIDs) == 0 {
		return []model.CourseLesson{}, nil
	}
	const query = `
		SELECT
			binding."CourseId",
			lesson."Id",
			binding."ParentGroupId",
			binding."Order",
			binding."IsTrial",
			binding."IsDeleted",
			binding."IsClosed",
			binding."ClosedUntil",
			lesson."Title",
			lesson."Description",
			lesson."TitleNavigation",
			lesson."Duration",
			lesson."IsGroup",
			lesson."IsDeleted",
			lesson."Modified",
			lesson_type."Id",
			lesson_type."Title",
			lesson_type."IsDeleted",
			ratings.average,
			ratings.count
		FROM public."CourseLessonBindings" AS binding
		JOIN public."Lessons" AS lesson ON lesson."Id" = binding."LessonId"
		JOIN public."LessonTypes" AS lesson_type ON lesson_type."Id" = lesson."TypeId"
		LEFT JOIN (
			SELECT "EntityId", AVG("Value")::double precision AS average, COUNT(*) AS count
			FROM public."UserRatings"
			WHERE "EntityTypeId" = $2::uuid AND NOT "IsDeleted"
			GROUP BY "EntityId"
		) AS ratings ON ratings."EntityId" = lesson."Id"
		WHERE binding."CourseId" = ANY($1::uuid[])
		ORDER BY binding."CourseId", binding."Order" NULLS LAST, lesson."Id"
	`
	rows, err := r.executor(ctx).Query(ctx, query, courseIDs, lessonRatingEntityTypeID)
	if err != nil {
		return nil, fmt.Errorf("query course lessons: %w", err)
	}
	lessons, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.CourseLesson, error) {
		var lesson model.CourseLesson
		err := lesson.Scan(row)
		return lesson, err
	})
	if err != nil {
		return nil, fmt.Errorf("scan course lessons: %w", err)
	}
	return lessons, nil
}
