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

var articleEntityTypeID = uuid.MustParse("8ede4179-0faf-407d-a4fc-1d94f9a63089")

func (r *Repository) ListArticles(ctx context.Context) ([]domain.Material, error) {
	data, err := r.loadArticleData(ctx)
	if err != nil {
		return nil, fmt.Errorf("load articles: %w", err)
	}

	materials := make([]domain.Material, 0, len(data))
	for _, article := range data {
		material, err := converter.ArticleToMaterial(article)
		if err != nil {
			return nil, fmt.Errorf("convert articles: %w", err)
		}
		materials = append(materials, material)
	}
	return materials, nil
}

func (r *Repository) loadArticleRows(ctx context.Context) ([]model.Article, error) {
	const query = `
		SELECT
			a."Id",
			a."StatusId",
			a."TypeId",
			a."SubjectId",
			a."Title",
			a."Description",
			a."CardDescription",
			a."TitleTag",
			a."DescriptionTag",
			a."TitleNavigation",
			a."ContentHeadings",
			a."Duration",
			a."Url",
			a."IsDeleted",
			a."IsPublic",
			a."IsClosed",
			a."Published",
			a."Modified",
			subject."Title",
			subject."IsPublic",
			subject."IsDeleted",
			settings."Id",
			settings."StatusId",
			settings."TypeId",
			settings."IsDeleted"
		FROM public."Articles" AS a
		JOIN public."ArticleSubjects" AS subject ON subject."Id" = a."SubjectId"
		LEFT JOIN public."ContentSettings" AS settings
			ON settings."EntityId" = a."Id"
			AND settings."EntityTypeId" = $1::uuid
		ORDER BY a."Id"
	`

	rows, err := r.executor(ctx).Query(ctx, query, articleEntityTypeID)
	if err != nil {
		return nil, fmt.Errorf("query articles: %w", err)
	}

	articles, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.Article, error) {
		var article model.Article
		err := article.Scan(row)
		return article, err
	})
	if err != nil {
		return nil, fmt.Errorf("scan articles: %w", err)
	}

	return articles, nil
}
