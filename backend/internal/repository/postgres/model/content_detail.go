package model

import (
	"uuid"

	"github.com/jackc/pgx/v5"
)

type ContentDetail struct {
	ID          uuid.UUID
	EntityID    uuid.UUID
	Title       string
	Description *string
	Position    *int32
	IsMain      bool
}

func (detail *ContentDetail) Scan(row pgx.Row) error {
	return row.Scan(
		&detail.ID,
		&detail.EntityID,
		&detail.Title,
		&detail.Description,
		&detail.Position,
		&detail.IsMain,
	)
}
