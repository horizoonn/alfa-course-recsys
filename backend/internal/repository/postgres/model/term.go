package model

import (
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
)

type Term struct {
	ID                     uuid.UUID
	StatusID               uuid.UUID
	Title                  string
	Description            *string
	AlternativeTitle       *string
	AlternativeDescription *string
	NavigationTitle        *string
	Letter                 *string
	DurationText           *string
	Slug                   string
	IsDeleted              bool
	PublishedAt            *time.Time
	ModifiedAt             time.Time
}

func (term *Term) Scan(row pgx.Row) error {
	return row.Scan(
		&term.ID,
		&term.StatusID,
		&term.Title,
		&term.Description,
		&term.AlternativeTitle,
		&term.AlternativeDescription,
		&term.NavigationTitle,
		&term.Letter,
		&term.DurationText,
		&term.Slug,
		&term.IsDeleted,
		&term.PublishedAt,
		&term.ModifiedAt,
	)
}

type TermData struct {
	Term   Term
	Blocks []ContentBlock
}
