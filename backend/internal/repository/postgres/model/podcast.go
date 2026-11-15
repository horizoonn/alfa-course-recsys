package model

import (
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
)

type Podcast struct {
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
	Slug                   string
	IsDeleted              bool
	IsClosed               bool
	PublishedAt            *time.Time
	ModifiedAt             time.Time
	Rating                 *float64
	ContentSettingsID      *uuid.UUID
	SettingsStatusID       *uuid.UUID
	SettingsAccessTypeID   *uuid.UUID
	SettingsIsDeleted      *bool
	HasActiveSubscription  bool
}

func (podcast *Podcast) Scan(row pgx.Row) error {
	return row.Scan(
		&podcast.ID,
		&podcast.StatusID,
		&podcast.AccessTypeID,
		&podcast.Title,
		&podcast.Description,
		&podcast.AdditionalDescription,
		&podcast.AlternativeTitle,
		&podcast.AlternativeDescription,
		&podcast.NavigationTitle,
		&podcast.DurationText,
		&podcast.Slug,
		&podcast.IsDeleted,
		&podcast.IsClosed,
		&podcast.PublishedAt,
		&podcast.ModifiedAt,
		&podcast.Rating,
		&podcast.ContentSettingsID,
		&podcast.SettingsStatusID,
		&podcast.SettingsAccessTypeID,
		&podcast.SettingsIsDeleted,
		&podcast.HasActiveSubscription,
	)
}

type PodcastData struct {
	Podcast    Podcast
	Topics     []ContentTopic
	Tags       []ContentTag
	Industries []ContentIndustry
	Details    []ContentDetail
}
