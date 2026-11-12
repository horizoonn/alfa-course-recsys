package converter

import (
	"fmt"
	"strings"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/model"
)

func DocumentToMaterial(data model.DocumentData) (domain.Material, error) {
	document := data.Document
	fragments, err := contentFragments(document.ID, data.Blocks, contentTextBlock)
	if err != nil {
		return domain.Material{}, fmt.Errorf("document %s: fragments: %w", document.ID, err)
	}
	description, err := optionalHTMLToText(document.DescriptionHTML)
	if err != nil {
		return domain.Material{}, fmt.Errorf("document %s: description: %w", document.ID, err)
	}
	metadata := domain.MaterialMetadata{
		Fragments: fragments,
		Category: &domain.Label{
			ID:    document.CategoryID,
			Title: normalizeText(document.CategoryTitle),
		},
		Description:            description,
		AlternativeTitle:       normalizeOptionalText(document.AlternativeTitle),
		AlternativeDescription: normalizeOptionalText(document.AlternativeDescription),
		NavigationTitle:        normalizeOptionalText(document.NavigationTitle),
		AdditionalLabels:       additionalLabels(data.Industries),
		File:                   documentFile(document),
	}
	title := normalizeText(document.Title)
	parts := []string{title}
	if description != nil {
		parts = append(parts, *description)
	}
	access := domain.AccessModelUnknown
	if document.SettingsAccessTypeID != nil {
		access = accessTypeModel(*document.SettingsAccessTypeID)
	}
	visibility := documentVisibility(document)
	parts = appendSearchFragments(parts, fragments, visibility, access)
	material := domain.Material{
		Key: domain.MaterialKey{
			Kind: domain.MaterialKindDocument,
			ID:   document.ID,
		},
		Title:            title,
		Summary:          shortSummary(description),
		SearchText:       joinNormalizedText(parts...),
		Topics:           topicLabels(data.Topics),
		Tags:             tagLabels(data.Tags),
		Metadata:         metadata,
		Visibility:       visibility,
		Access:           access,
		URL:              materialURL("documents", document.Slug),
		SourceModifiedAt: document.ModifiedAt.UTC(),
	}
	if document.PublishedAt != nil {
		publishedAt := document.PublishedAt.UTC()
		material.PublishedAt = &publishedAt
	}
	if err := material.Validate(); err != nil {
		return domain.Material{}, fmt.Errorf("document %s: material: %w", document.ID, err)
	}
	return material, nil
}

func documentFile(document model.Document) *domain.MaterialFile {
	extension := normalizeOptionalText(document.FileExtension)
	if extension != nil {
		value := strings.ToLower(*extension)
		extension = &value
	}
	if extension == nil && document.FileSizeBytes == nil {
		return nil
	}
	return &domain.MaterialFile{
		Extension: extension,
		SizeBytes: document.FileSizeBytes,
	}
}

func documentVisibility(document model.Document) domain.Visibility {
	if document.IsDeleted || document.CategoryIsDeleted ||
		(document.SettingsIsDeleted != nil && *document.SettingsIsDeleted) {
		return domain.VisibilityHidden
	}
	if document.StatusID == draftStatusID ||
		(document.SettingsStatusID != nil && *document.SettingsStatusID == draftStatusID) {
		return domain.VisibilityHidden
	}
	if document.ContentSettingsID == nil || document.SettingsStatusID == nil ||
		document.SettingsIsDeleted == nil {
		return domain.VisibilityUnknown
	}
	if document.StatusID != publishedStatusID || *document.SettingsStatusID != publishedStatusID {
		return domain.VisibilityUnknown
	}
	return domain.VisibilityListed
}
