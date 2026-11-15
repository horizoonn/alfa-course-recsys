package model

import (
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
)

type Document struct {
	ID                     uuid.UUID
	StatusID               uuid.UUID
	CategoryID             uuid.UUID
	Title                  string
	DescriptionHTML        *string
	AlternativeTitle       *string
	AlternativeDescription *string
	NavigationTitle        *string
	FileExtension          *string
	FileSizeBytes          *int64
	Slug                   string
	IsDeleted              bool
	PublishedAt            *time.Time
	ModifiedAt             time.Time

	CategoryTitle     string
	CategoryIsDeleted bool

	ContentSettingsID    *uuid.UUID
	SettingsStatusID     *uuid.UUID
	SettingsAccessTypeID *uuid.UUID
	SettingsIsDeleted    *bool
}

func (document *Document) Scan(row pgx.Row) error {
	return row.Scan(
		&document.ID,
		&document.StatusID,
		&document.CategoryID,
		&document.Title,
		&document.DescriptionHTML,
		&document.AlternativeTitle,
		&document.AlternativeDescription,
		&document.NavigationTitle,
		&document.FileExtension,
		&document.FileSizeBytes,
		&document.Slug,
		&document.IsDeleted,
		&document.PublishedAt,
		&document.ModifiedAt,
		&document.CategoryTitle,
		&document.CategoryIsDeleted,
		&document.ContentSettingsID,
		&document.SettingsStatusID,
		&document.SettingsAccessTypeID,
		&document.SettingsIsDeleted,
	)
}

type DocumentData struct {
	Blocks     []ContentBlock
	Document   Document
	Topics     []ContentTopic
	Tags       []ContentTag
	Industries []ContentIndustry
}
