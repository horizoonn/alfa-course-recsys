package model

import (
	"uuid"

	"github.com/jackc/pgx/v5"
)

type ContentTopic struct {
	ContentSettingsID uuid.UUID
	TopicID           uuid.UUID
	Title             string
	IsDeleted         bool
	IsPublic          bool
	Position          int32
}

func (topic *ContentTopic) Scan(row pgx.Row) error {
	return row.Scan(
		&topic.ContentSettingsID,
		&topic.TopicID,
		&topic.Title,
		&topic.IsDeleted,
		&topic.IsPublic,
		&topic.Position,
	)
}

type ContentTag struct {
	ContentSettingsID uuid.UUID
	TagID             uuid.UUID
	Title             string
	IsDeleted         bool
	Position          int32
}

func (tag *ContentTag) Scan(row pgx.Row) error {
	return row.Scan(
		&tag.ContentSettingsID,
		&tag.TagID,
		&tag.Title,
		&tag.IsDeleted,
		&tag.Position,
	)
}

type ContentIndustry struct {
	ContentSettingsID uuid.UUID
	IndustryID        uuid.UUID
	Title             string
	IsDeleted         bool
	Position          int32
}

func (industry *ContentIndustry) Scan(row pgx.Row) error {
	return row.Scan(
		&industry.ContentSettingsID,
		&industry.IndustryID,
		&industry.Title,
		&industry.IsDeleted,
		&industry.Position,
	)
}
