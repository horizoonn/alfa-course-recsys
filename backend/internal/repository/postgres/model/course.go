package model

import (
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
)

type Course struct {
	ID                     uuid.UUID
	StatusID               uuid.UUID
	AccessTypeID           uuid.UUID
	Title                  string
	Description            string
	AdditionalDescription  *string
	AlternativeTitle       *string
	AlternativeDescription *string
	NavigationTitle        *string
	TargetHTML             *string
	DurationText           *string
	Slug                   string
	IsDeleted              bool
	IsClosed               bool
	PublishedAt            *time.Time
	ModifiedAt             time.Time
	Rating                 *float64

	ContentSettingsID    *uuid.UUID
	SettingsStatusID     *uuid.UUID
	SettingsAccessTypeID *uuid.UUID
	SettingsIsDeleted    *bool

	HasActiveSubscription bool
	HasTariffs            bool
}

func (course *Course) Scan(row pgx.Row) error {
	return row.Scan(
		&course.ID,
		&course.StatusID,
		&course.AccessTypeID,
		&course.Title,
		&course.Description,
		&course.AdditionalDescription,
		&course.AlternativeTitle,
		&course.AlternativeDescription,
		&course.NavigationTitle,
		&course.TargetHTML,
		&course.DurationText,
		&course.Slug,
		&course.IsDeleted,
		&course.IsClosed,
		&course.PublishedAt,
		&course.ModifiedAt,
		&course.Rating,
		&course.ContentSettingsID,
		&course.SettingsStatusID,
		&course.SettingsAccessTypeID,
		&course.SettingsIsDeleted,
		&course.HasActiveSubscription,
		&course.HasTariffs,
	)
}

type CourseInfo struct {
	ID                uuid.UUID
	CourseID          uuid.UUID
	InfoTypeID        uuid.UUID
	InfoTypeTitle     string
	InfoTypeIsDeleted bool
	Title             string
	Description       *string
	Position          *int32
	IsMain            bool
	TariffID          *uuid.UUID
}

func (info *CourseInfo) Scan(row pgx.Row) error {
	return row.Scan(
		&info.ID,
		&info.CourseID,
		&info.InfoTypeID,
		&info.InfoTypeTitle,
		&info.InfoTypeIsDeleted,
		&info.Title,
		&info.Description,
		&info.Position,
		&info.IsMain,
		&info.TariffID,
	)
}

type CourseData struct {
	Course     Course
	Topics     []ContentTopic
	Tags       []ContentTag
	Industries []ContentIndustry
	Info       []CourseInfo
	Lessons    []CourseLesson
}
