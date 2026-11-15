package model

import (
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
)

type Webinar struct {
	ID                     uuid.UUID
	StatusID               uuid.UUID
	AccessTypeID           uuid.UUID
	Title                  string
	Description            *string
	AdditionalDescription  *string
	AlternativeTitle       *string
	AlternativeDescription *string
	NavigationTitle        *string
	DurationText           *string
	Slug                   *string
	IsDeleted              bool
	IsClosed               bool
	ModifiedAt             time.Time
	Rating                 *float64
	StartsAt               *time.Time
	CompletedAt            *time.Time
	RegistrationRequired   bool
	PasswordRequired       bool
	ModerationRequired     bool
	ContentSettingsID      *uuid.UUID
	SettingsStatusID       *uuid.UUID
	SettingsAccessTypeID   *uuid.UUID
	SettingsIsDeleted      *bool
	SettingsPublishedAt    *time.Time
}

func (webinar *Webinar) Scan(row pgx.Row) error {
	return row.Scan(
		&webinar.ID,
		&webinar.StatusID,
		&webinar.AccessTypeID,
		&webinar.Title,
		&webinar.Description,
		&webinar.AdditionalDescription,
		&webinar.AlternativeTitle,
		&webinar.AlternativeDescription,
		&webinar.NavigationTitle,
		&webinar.DurationText,
		&webinar.Slug,
		&webinar.IsDeleted,
		&webinar.IsClosed,
		&webinar.ModifiedAt,
		&webinar.Rating,
		&webinar.StartsAt,
		&webinar.CompletedAt,
		&webinar.RegistrationRequired,
		&webinar.PasswordRequired,
		&webinar.ModerationRequired,
		&webinar.ContentSettingsID,
		&webinar.SettingsStatusID,
		&webinar.SettingsAccessTypeID,
		&webinar.SettingsIsDeleted,
		&webinar.SettingsPublishedAt,
	)
}

type WebinarData struct {
	Webinar    Webinar
	Topics     []ContentTopic
	Tags       []ContentTag
	Industries []ContentIndustry
	Details    []ContentDetail
}
