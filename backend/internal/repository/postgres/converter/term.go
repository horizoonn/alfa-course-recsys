package converter

import (
	"fmt"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/model"
)

func TermToMaterial(data model.TermData) (domain.Material, error) {
	term := data.Term
	fragments, err := contentFragments(
		term.ID,
		data.Blocks,
		contentTextBlock,
		contentNoteBlock,
		contentListBlock,
		contentDetailsBlock,
	)
	if err != nil {
		return domain.Material{}, fmt.Errorf("term %s: fragments: %w", term.ID, err)
	}
	description := normalizeOptionalText(term.Description)
	summary := description
	if summary == nil && len(fragments) > 0 {
		summary = &fragments[0].Text
	}
	title := normalizeText(term.Title)
	parts := make([]string, 0, len(fragments)+2)
	parts = append(parts, title)
	if description != nil {
		parts = append(parts, *description)
	}
	for _, fragment := range fragments {
		parts = append(parts, fragment.Text)
	}
	material := domain.Material{
		Key: domain.MaterialKey{
			Kind: domain.MaterialKindTerm,
			ID:   term.ID,
		},
		Title:      title,
		Summary:    shortSummary(summary),
		SearchText: joinNormalizedText(parts...),
		Topics:     []domain.Label{},
		Tags:       []domain.Label{},
		Metadata: domain.MaterialMetadata{
			Description:            description,
			Fragments:              fragments,
			AlternativeTitle:       normalizeOptionalText(term.AlternativeTitle),
			AlternativeDescription: normalizeOptionalText(term.AlternativeDescription),
			NavigationTitle:        normalizeOptionalText(term.NavigationTitle),
			Term: &domain.TermMetadata{
				Letter: normalizeOptionalText(term.Letter),
			},
		},
		Duration:         materialDuration(term.DurationText),
		Visibility:       termVisibility(term),
		Access:           domain.AccessModelFree,
		URL:              materialURL("vocabulary", term.Slug),
		SourceModifiedAt: term.ModifiedAt.UTC(),
	}
	if term.PublishedAt != nil {
		published := term.PublishedAt.UTC()
		material.PublishedAt = &published
	}
	if err := material.Validate(); err != nil {
		return domain.Material{}, fmt.Errorf("term %s: material: %w", term.ID, err)
	}
	return material, nil
}

func termVisibility(term model.Term) domain.Visibility {
	if term.IsDeleted || term.StatusID == draftStatusID {
		return domain.VisibilityHidden
	}
	if term.StatusID != publishedStatusID {
		return domain.VisibilityUnknown
	}
	return domain.VisibilityListed
}
