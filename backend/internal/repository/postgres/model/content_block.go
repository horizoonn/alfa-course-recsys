package model

import (
	"uuid"

	"github.com/jackc/pgx/v5"
)

type ContentBlock struct {
	ID         uuid.UUID
	EntityID   uuid.UUID
	Type       int32
	HTML       *string
	ParamsJSON *string
	Position   *int32
}

func (block *ContentBlock) Scan(row pgx.Row) error {
	return row.Scan(
		&block.ID,
		&block.EntityID,
		&block.Type,
		&block.HTML,
		&block.ParamsJSON,
		&block.Position,
	)
}
