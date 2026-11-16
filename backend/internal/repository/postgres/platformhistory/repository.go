package platformhistory

import (
	"context"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (r *Repository) Load(ctx context.Context) ([]domain.LearningRecord, []domain.RatingRecord, error) {
	learning := []domain.LearningRecord{}
	ratings := []domain.RatingRecord{}
	err := pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly}, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `WITH progress AS (
   SELECT ul."CourseUserId",COUNT(DISTINCT ul."LessonId")::int AS completed
   FROM public."CourseUserLessonBindings" ul
   JOIN public."CourseUserBindings" u ON u."Id"=ul."CourseUserId" AND NOT u."IsDeleted"
   JOIN public."CourseLessonBindings" l ON l."CourseId"=u."CourseId" AND l."LessonId"=ul."LessonId" AND NOT l."IsDeleted"
   WHERE ul."Completed" IS NOT NULL GROUP BY ul."CourseUserId"
  ) SELECT u."UserId",u."CourseId",u."IsCompleted",u."StudyingStartDate" IS NOT NULL,
   GREATEST(COALESCE(u."LessonsCompleted",0),COALESCE(p.completed,0))
   FROM public."CourseUserBindings" u LEFT JOIN progress p ON p."CourseUserId"=u."Id"
   WHERE NOT u."IsDeleted" AND u."UserId" IS NOT NULL
   ORDER BY u."UserId",u."CourseId"`)
		if err != nil {
			return err
		}
		for rows.Next() {
			var item domain.LearningRecord
			if err := rows.Scan(&item.UserID, &item.CourseID, &item.Completed, &item.Started, &item.LessonsCompleted); err != nil {
				rows.Close()
				return err
			}
			learning = append(learning, item)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
		rows, err = tx.Query(ctx, `SELECT DISTINCT ON ("UserId","EntityTypeId","EntityId")
   "UserId","EntityId","EntityTypeId","Value" FROM public."UserRatings"
   WHERE NOT "IsDeleted" AND "UserId" IS NOT NULL AND "Value" BETWEEN 1 AND 5
   ORDER BY "UserId","EntityTypeId","EntityId",COALESCE("Modified","Created") DESC,"Id" DESC`)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var item domain.RatingRecord
			if err := rows.Scan(&item.UserID, &item.EntityID, &item.EntityTypeID, &item.Value); err != nil {
				return err
			}
			ratings = append(ratings, item)
		}
		return rows.Err()
	})
	return learning, ratings, err
}
