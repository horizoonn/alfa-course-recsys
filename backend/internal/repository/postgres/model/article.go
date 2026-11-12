package model

import (
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
)

type Article struct {
	ID                     uuid.UUID
	StatusID               uuid.UUID
	AccessTypeID           uuid.UUID
	SubjectID              uuid.UUID
	Title                  string
	Description            *string
	CardDescription        *string
	AlternativeTitle       *string
	AlternativeDescription *string
	NavigationTitle        *string
	HeadingsJSON           *string
	DurationText           *string
	Slug                   string
	IsDeleted              bool
	IsPublic               bool
	IsClosed               bool
	PublishedAt            *time.Time
	ModifiedAt             time.Time

	SubjectTitle     string
	SubjectIsPublic  bool
	SubjectIsDeleted bool

	ContentSettingsID    *uuid.UUID
	SettingsStatusID     *uuid.UUID
	SettingsAccessTypeID *uuid.UUID
	SettingsIsDeleted    *bool
}

func (article *Article) Scan(row pgx.Row) error {
	return row.Scan(
		&article.ID,
		&article.StatusID,
		&article.AccessTypeID,
		&article.SubjectID,
		&article.Title,
		&article.Description,
		&article.CardDescription,
		&article.AlternativeTitle,
		&article.AlternativeDescription,
		&article.NavigationTitle,
		&article.HeadingsJSON,
		&article.DurationText,
		&article.Slug,
		&article.IsDeleted,
		&article.IsPublic,
		&article.IsClosed,
		&article.PublishedAt,
		&article.ModifiedAt,
		&article.SubjectTitle,
		&article.SubjectIsPublic,
		&article.SubjectIsDeleted,
		&article.ContentSettingsID,
		&article.SettingsStatusID,
		&article.SettingsAccessTypeID,
		&article.SettingsIsDeleted,
	)
}

type ArticleData struct {
	Blocks     []ContentBlock
	Article    Article
	Topics     []ContentTopic
	Tags       []ContentTag
	Industries []ContentIndustry
}
