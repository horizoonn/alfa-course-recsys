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

var calculatorRatingEntityTypeIDs = []uuid.UUID{
	uuid.MustParse("7ac1c8c3-18b8-42ea-92ea-fcc10a6d496f"),
	uuid.MustParse("72dfaf58-4e43-4b92-a1cd-5f8b5de739b4"),
}

func (r *Repository) ListCalculators(ctx context.Context) ([]domain.Material, error) {
	data, err := r.loadCalculatorData(ctx)
	if err != nil {
		return nil, fmt.Errorf("load calculators: %w", err)
	}
	materials := make([]domain.Material, 0, len(data))
	for _, item := range data {
		material, err := converter.CalculatorToMaterial(item)
		if err != nil {
			return nil, fmt.Errorf("convert calculators: %w", err)
		}
		materials = append(materials, material)
	}
	return materials, nil
}

func (r *Repository) loadCalculatorRows(ctx context.Context) ([]model.Calculator, error) {
	const query = `
		SELECT
			calculator."Id",
			calculator."Title",
			calculator."Description",
			calculator."IsDeleted",
			calculator."Modified",
			ratings.average,
			ratings.count
		FROM public."Calculators" AS calculator
		LEFT JOIN (
			SELECT "EntityId", AVG("Value")::double precision AS average, COUNT(*) AS count
			FROM public."UserRatings"
			WHERE "EntityTypeId" = ANY($1::uuid[]) AND NOT "IsDeleted"
			GROUP BY "EntityId"
		) AS ratings ON ratings."EntityId" = calculator."Id"
		ORDER BY calculator."Id"
	`
	rows, err := r.executor(ctx).Query(ctx, query, calculatorRatingEntityTypeIDs)
	if err != nil {
		return nil, fmt.Errorf("query calculators: %w", err)
	}
	calculators, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.Calculator, error) {
		var calculator model.Calculator
		err := calculator.Scan(row)
		return calculator, err
	})
	if err != nil {
		return nil, fmt.Errorf("scan calculators: %w", err)
	}
	return calculators, nil
}

func (r *Repository) loadCalculatorPages(ctx context.Context) ([]model.CalculatorPage, error) {
	const query = `
		SELECT
			"Id",
			"StatusId",
			"Title",
			"Description",
			"TitleTag",
			"DescriptionTag",
			"TitleNavigation",
			"Url",
			"IsDeleted",
			"Published",
			"Modified"
		FROM public."CalculatorPages"
		ORDER BY "Id"
	`
	rows, err := r.executor(ctx).Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query calculator pages: %w", err)
	}
	pages, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.CalculatorPage, error) {
		var page model.CalculatorPage
		err := page.Scan(row)
		return page, err
	})
	if err != nil {
		return nil, fmt.Errorf("scan calculator pages: %w", err)
	}
	return pages, nil
}
