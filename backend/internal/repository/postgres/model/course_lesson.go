package model

import (
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
)

type CourseLesson struct {
	CourseID            uuid.UUID
	LessonID            uuid.UUID
	ParentGroupID       *uuid.UUID
	Position            *int32
	IsTrial             bool
	BindingIsDeleted    bool
	IsClosed            bool
	ClosedUntil         *time.Time
	Title               string
	Description         *string
	NavigationTitle     *string
	DurationText        *string
	IsGroup             bool
	LessonIsDeleted     bool
	ModifiedAt          time.Time
	LessonTypeID        uuid.UUID
	LessonTypeTitle     string
	LessonTypeIsDeleted bool
	RatingAverage       *float64
	RatingCount         *int64
}

func (lesson *CourseLesson) Scan(row pgx.Row) error {
	return row.Scan(
		&lesson.CourseID,
		&lesson.LessonID,
		&lesson.ParentGroupID,
		&lesson.Position,
		&lesson.IsTrial,
		&lesson.BindingIsDeleted,
		&lesson.IsClosed,
		&lesson.ClosedUntil,
		&lesson.Title,
		&lesson.Description,
		&lesson.NavigationTitle,
		&lesson.DurationText,
		&lesson.IsGroup,
		&lesson.LessonIsDeleted,
		&lesson.ModifiedAt,
		&lesson.LessonTypeID,
		&lesson.LessonTypeTitle,
		&lesson.LessonTypeIsDeleted,
		&lesson.RatingAverage,
		&lesson.RatingCount,
	)
}
