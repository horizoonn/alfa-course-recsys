package catalog

import (
	"context"
	"fmt"
	"uuid"

	"github.com/jackc/pgx/v5"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/converter"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/model"
)

var documentEntityTypeID = uuid.MustParse("da1ad865-8605-4ab1-9f6c-4489a26fbc35")

func (r *Repository) ListDocuments(ctx context.Context) ([]domain.Material, error) {
	data, err := r.loadDocumentData(ctx)
	if err != nil {
		return nil, fmt.Errorf("load documents: %w", err)
	}

	materials := make([]domain.Material, 0, len(data))
	for _, item := range data {
		material, err := converter.DocumentToMaterial(item)
		if err != nil {
			return nil, fmt.Errorf("convert documents: %w", err)
		}
		materials = append(materials, material)
	}
	return materials, nil
}

func (r *Repository) loadDocumentRows(ctx context.Context) ([]model.Document, error) {
	const query = `
		SELECT
			document."Id",
			document."StatusId",
			document."TypeId",
			document."Title",
			document."Description",
			document."TitleTag",
			document."DescriptionTag",
			document."TitleNavigation",
			document."Extension",
			document."Size",
			document."Url",
			document."IsDeleted",
			document."Published",
			document."Modified",
			category."Title",
			category."IsDeleted",
			settings."Id",
			settings."StatusId",
			settings."TypeId",
			settings."IsDeleted"
		FROM public."Documents" AS document
		JOIN public."DocumentTypes" AS category ON category."Id" = document."TypeId"
		LEFT JOIN public."ContentSettings" AS settings
			ON settings."EntityId" = document."Id"
			AND settings."EntityTypeId" = $1::uuid
		ORDER BY document."Id"
	`

	rows, err := r.executor(ctx).Query(ctx, query, documentEntityTypeID)
	if err != nil {
		return nil, fmt.Errorf("query documents: %w", err)
	}

	documents, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.Document, error) {
		var document model.Document
		err := document.Scan(row)
		return document, err
	})
	if err != nil {
		return nil, fmt.Errorf("scan documents: %w", err)
	}

	return documents, nil
}
