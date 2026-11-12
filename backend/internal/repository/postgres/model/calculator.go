package model

import (
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
)

type Calculator struct {
	ID            uuid.UUID
	Title         string
	Description   *string
	IsDeleted     bool
	ModifiedAt    time.Time
	RatingAverage *float64
	RatingCount   *int64
}

func (calculator *Calculator) Scan(row pgx.Row) error {
	return row.Scan(
		&calculator.ID,
		&calculator.Title,
		&calculator.Description,
		&calculator.IsDeleted,
		&calculator.ModifiedAt,
		&calculator.RatingAverage,
		&calculator.RatingCount,
	)
}

type CalculatorPage struct {
	ID                     uuid.UUID
	StatusID               uuid.UUID
	Title                  string
	DescriptionHTML        *string
	AlternativeTitle       *string
	AlternativeDescription *string
	NavigationTitle        *string
	Slug                   string
	IsDeleted              bool
	PublishedAt            *time.Time
	ModifiedAt             time.Time
}

func (page *CalculatorPage) Scan(row pgx.Row) error {
	return row.Scan(
		&page.ID,
		&page.StatusID,
		&page.Title,
		&page.DescriptionHTML,
		&page.AlternativeTitle,
		&page.AlternativeDescription,
		&page.NavigationTitle,
		&page.Slug,
		&page.IsDeleted,
		&page.PublishedAt,
		&page.ModifiedAt,
	)
}

type CalculatorData struct {
	Calculator Calculator
	Page       CalculatorPage
	Blocks     []ContentBlock
}
