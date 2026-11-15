package converter

import (
	"fmt"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/model"
)

func WebinarToMaterial(data model.WebinarData) (domain.Material, error) {
	source := data.Webinar
	fragments, err := detailFragments(source.ID, data.Details)
	if err != nil {
		return domain.Material{}, fmt.Errorf("webinar %s: details: %w", source.ID, err)
	}
	description, err := optionalHTMLToText(source.Description)
	if err != nil {
		return domain.Material{}, fmt.Errorf("webinar %s: description: %w", source.ID, err)
	}
	additional, err := optionalHTMLToText(source.AdditionalDescription)
	if err != nil {
		return domain.Material{}, fmt.Errorf("webinar %s: additional description: %w", source.ID, err)
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

	details := &domain.WebinarMetadata{RegistrationRequired: source.RegistrationRequired}
	if source.StartsAt != nil {
		at := source.StartsAt.UTC()
		details.StartsAt = &at
	}
	if source.CompletedAt != nil {
		at := source.CompletedAt.UTC()
		details.CompletedAt = &at
	}
	metadata.Webinar = details

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

	slug := ""
	if source.Slug != nil {
		slug = *source.Slug
	}

	material := domain.Material{
		Key: domain.MaterialKey{
			Kind: domain.MaterialKindWebinar,
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
		Visibility:       webinarVisibility(source),
		Access:           webinarAccess(source),
		URL:              materialURL("webinars", slug),
		SourceModifiedAt: source.ModifiedAt.UTC(),
	}
	if source.SettingsPublishedAt != nil {
		at := source.SettingsPublishedAt.UTC()
		material.PublishedAt = &at
	}
	if err := material.Validate(); err != nil {
		return domain.Material{}, fmt.Errorf("webinar %s: material: %w", source.ID, err)
	}
	return material, nil
}

func webinarVisibility(source model.Webinar) domain.Visibility {
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

	if source.PasswordRequired || source.ModerationRequired {
		return domain.VisibilityUnknown
	}

	return domain.VisibilityListed
}

func webinarAccess(source model.Webinar) domain.AccessModel {
	access := accessModel(source.AccessTypeID, source.SettingsAccessTypeID)
	if access != domain.AccessModelFree {
		return domain.AccessModelUnknown
	}
	return access
}
