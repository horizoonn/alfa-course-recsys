package converter

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/model"
)

func ArticleToMaterial(data model.ArticleData) (domain.Material, error) {
	article := data.Article
	fragments, err := contentFragments(
		article.ID,
		data.Blocks,
		contentTextBlock,
		contentNoteBlock,
		contentQuoteBlock,
		contentListBlock,
		contentDetailsBlock,
	)
	if err != nil {
		return domain.Material{}, fmt.Errorf("article %s: fragments: %w", article.ID, err)
	}

	headings, err := articleHeadings(article.HeadingsJSON)
	if err != nil {
		return domain.Material{}, fmt.Errorf("article %s: headings: %w", article.ID, err)
	}
	for i := range headings {
		headings[i] = normalizeText(headings[i])
	}

	metadata := domain.MaterialMetadata{
		Fragments: fragments,
		Category: &domain.Label{
			ID:    article.SubjectID,
			Title: normalizeText(article.SubjectTitle),
		},
		AdditionalLabels:       additionalLabels(data.Industries),
		Description:            normalizeOptionalText(article.Description),
		CardDescription:        normalizeOptionalText(article.CardDescription),
		AlternativeTitle:       normalizeOptionalText(article.AlternativeTitle),
		AlternativeDescription: normalizeOptionalText(article.AlternativeDescription),
		NavigationTitle:        normalizeOptionalText(article.NavigationTitle),
		Headings:               headings,
	}

	summary := metadata.Description
	if summary == nil {
		summary = metadata.CardDescription
	}

	title := normalizeText(article.Title)
	parts := make([]string, 0, len(metadata.Headings)+3)
	parts = append(parts, title)
	if metadata.Description != nil {
		parts = append(parts, *metadata.Description)
	}
	if metadata.CardDescription != nil {
		parts = append(parts, *metadata.CardDescription)
	}
	parts = append(parts, metadata.Headings...)

	visibility := articleVisibility(article)
	access := accessModel(article.AccessTypeID, article.SettingsAccessTypeID)
	parts = appendSearchFragments(parts, fragments, visibility, access)
	material := domain.Material{
		Key: domain.MaterialKey{
			Kind: domain.MaterialKindArticle,
			ID:   article.ID,
		},
		Title:            title,
		Summary:          summary,
		SearchText:       joinNormalizedText(parts...),
		Topics:           topicLabels(data.Topics),
		Tags:             tagLabels(data.Tags),
		Metadata:         metadata,
		Duration:         materialDuration(article.DurationText),
		Visibility:       visibility,
		Access:           access,
		URL:              materialURL("articles", article.Slug),
		SourceModifiedAt: article.ModifiedAt.UTC(),
	}
	if article.PublishedAt != nil {
		publishedAt := article.PublishedAt.UTC()
		material.PublishedAt = &publishedAt
	}

	if err := material.Validate(); err != nil {
		return domain.Material{}, fmt.Errorf("article %s: material: %w", article.ID, err)
	}
	return material, nil
}

func articleHeadings(value *string) ([]string, error) {
	if value == nil {
		return nil, nil
	}

	var headings []*string
	if err := json.Unmarshal([]byte(*value), &headings); err != nil {
		return nil, err
	}
	if headings == nil {
		return nil, errors.New("expected an array of strings")
	}

	result := make([]string, 0, len(headings))
	for i, heading := range headings {
		if heading == nil {
			return nil, fmt.Errorf("item %d: expected a string", i)
		}
		result = append(result, *heading)
	}
	return result, nil
}

func articleVisibility(article model.Article) domain.Visibility {
	if article.IsDeleted || article.IsClosed || !article.IsPublic ||
		article.SubjectIsDeleted || !article.SubjectIsPublic ||
		(article.SettingsIsDeleted != nil && *article.SettingsIsDeleted) {
		return domain.VisibilityHidden
	}
	if article.StatusID == draftStatusID ||
		(article.SettingsStatusID != nil && *article.SettingsStatusID == draftStatusID) {
		return domain.VisibilityHidden
	}
	if article.ContentSettingsID == nil || article.SettingsStatusID == nil ||
		article.SettingsIsDeleted == nil {
		return domain.VisibilityUnknown
	}
	if article.StatusID != publishedStatusID || *article.SettingsStatusID != publishedStatusID {
		return domain.VisibilityUnknown
	}
	return domain.VisibilityListed
}
