package converter

import (
	"fmt"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/model"
)

func PodcastToMaterial(data model.PodcastData) (domain.Material, error) {
	source := data.Podcast
	fragments, err := detailFragments(source.ID, data.Details)
	if err != nil {
		return domain.Material{}, fmt.Errorf("podcast %s: details: %w", source.ID, err)
	}
	description, err := optionalHTMLToText(source.Description)
	if err != nil {
		return domain.Material{}, fmt.Errorf("podcast %s: description: %w", source.ID, err)
	}
	additional, err := optionalHTMLToText(source.AdditionalDescription)
	if err != nil {
		return domain.Material{}, fmt.Errorf("podcast %s: additional description: %w", source.ID, err)
	}
	metadata := domain.MaterialMetadata{
		Description:            description,
		AdditionalDescription:  additional,
		AlternativeTitle:       normalizeOptionalText(source.AlternativeTitle),
		AlternativeDescription: normalizeOptionalText(source.AlternativeDescription),
		NavigationTitle:        normalizeOptionalText(source.NavigationTitle),
		AdditionalLabels:       additionalLabels(data.Industries),
		Fragments:              fragments,
	}

	summary := description
	if summary == nil {
		summary = additional
	}
	materialTitle := normalizeText(source.Title)
	parts := []string{materialTitle}
	for _, value := range []*string{description, additional} {
		if value != nil {
			parts = append(parts, *value)
		}
	}
	for _, fragment := range fragments {
		parts = append(parts, fragment.Text)
	}

	material := domain.Material{
		Key: domain.MaterialKey{
			Kind: domain.MaterialKindPodcast,
			ID:   source.ID,
		},
		Title:            materialTitle,
		Summary:          shortSummary(summary),
		SearchText:       joinNormalizedText(parts...),
		Topics:           topicLabels(data.Topics),
		Tags:             tagLabels(data.Tags),
		Metadata:         metadata,
		Duration:         materialDuration(source.DurationText),
		Rating:           ratingSummary(source.Rating, nil),
		Visibility:       podcastVisibility(source),
		Access:           podcastAccess(source),
		URL:              materialURL("podcasts", source.Slug),
		SourceModifiedAt: source.ModifiedAt.UTC(),
	}
	if source.PublishedAt != nil {
		at := source.PublishedAt.UTC()
		material.PublishedAt = &at
	}
	if err := material.Validate(); err != nil {
		return domain.Material{}, fmt.Errorf("podcast %s: material: %w", source.ID, err)
	}
	return material, nil
}

func podcastVisibility(source model.Podcast) domain.Visibility {
	if source.IsDeleted || source.IsClosed || (source.SettingsIsDeleted != nil && *source.SettingsIsDeleted) {
		return domain.VisibilityHidden
	}
	if source.StatusID == draftStatusID || (source.SettingsStatusID != nil && *source.SettingsStatusID == draftStatusID) {
		return domain.VisibilityHidden
	}
	if source.ContentSettingsID == nil || source.SettingsStatusID == nil || source.SettingsIsDeleted == nil {
		return domain.VisibilityUnknown
	}
	if source.StatusID != publishedStatusID || *source.SettingsStatusID != publishedStatusID {
		return domain.VisibilityUnknown
	}

	return domain.VisibilityListed
}

func podcastAccess(source model.Podcast) domain.AccessModel {
	access := accessModel(source.AccessTypeID, source.SettingsAccessTypeID)
	if access == domain.AccessModelSubscription && !source.HasActiveSubscription {
		return domain.AccessModelUnknown
	}
	if access == domain.AccessModelPaid {
		return domain.AccessModelUnknown
	}
	return access
}
