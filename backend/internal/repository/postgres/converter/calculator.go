package converter

import (
	"fmt"
	"uuid"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/model"
)

const calculatorReferenceBlock int32 = 14

func CalculatorReferenceID(
	pageID uuid.UUID,
	blocks []model.ContentBlock,
) (uuid.UUID, error) {
	var reference uuid.UUID
	found := false
	for _, block := range blocks {
		if block.EntityID != pageID {
			return uuid.Nil(), fmt.Errorf("page %s: block %s belongs to another material", pageID, block.ID)
		}
		if block.Type != calculatorReferenceBlock {
			continue
		}
		if found {
			return uuid.Nil(), fmt.Errorf("page %s: duplicate calculator reference block %s", pageID, block.ID)
		}
		var params struct {
			Calculator *struct {
				ID *string
			}
		}
		if err := decodeBlockParams(block.ParamsJSON, &params); err != nil {
			return uuid.Nil(), fmt.Errorf("page %s: reference block %s: %w", pageID, block.ID, err)
		}
		if params.Calculator == nil || params.Calculator.ID == nil {
			return uuid.Nil(), fmt.Errorf("page %s: reference block %s: calculator ID is missing", pageID, block.ID)
		}
		id, err := uuid.Parse(*params.Calculator.ID)
		if err != nil {
			return uuid.Nil(), fmt.Errorf("page %s: reference block %s: calculator ID: %w", pageID, block.ID, err)
		}
		if id == uuid.Nil() {
			return uuid.Nil(), fmt.Errorf("page %s: reference block %s: calculator ID must not be nil UUID", pageID, block.ID)
		}
		reference = id
		found = true
	}
	if !found {
		return uuid.Nil(), fmt.Errorf("page %s: calculator reference is missing", pageID)
	}
	return reference, nil
}

func CalculatorToMaterial(data model.CalculatorData) (domain.Material, error) {
	calculator, page := data.Calculator, data.Page
	if page.ID == uuid.Nil() {
		return domain.Material{}, fmt.Errorf("calculator %s: page ID must not be nil UUID", calculator.ID)
	}
	reference, err := CalculatorReferenceID(page.ID, data.Blocks)
	if err != nil {
		return domain.Material{}, fmt.Errorf("calculator %s: %w", calculator.ID, err)
	}
	if reference != calculator.ID {
		return domain.Material{}, fmt.Errorf("calculator %s: page %s references calculator %s", calculator.ID, page.ID, reference)
	}
	fragments, err := contentFragments(page.ID, data.Blocks, contentTextBlock, contentNoteBlock)
	if err != nil {
		return domain.Material{}, fmt.Errorf("calculator %s: fragments: %w", calculator.ID, err)
	}
	description, err := optionalHTMLToText(page.DescriptionHTML)
	if err != nil {
		return domain.Material{}, fmt.Errorf("calculator %s: page description: %w", calculator.ID, err)
	}
	additional := normalizeOptionalText(calculator.Description)
	if description == nil {
		description, additional = additional, nil
	}
	metadata := domain.MaterialMetadata{
		Description:            description,
		AdditionalDescription:  additional,
		AlternativeTitle:       normalizeOptionalText(page.AlternativeTitle),
		AlternativeDescription: normalizeOptionalText(page.AlternativeDescription),
		NavigationTitle:        normalizeOptionalText(page.NavigationTitle),
		AdditionalLabels:       []domain.Label{},
		Fragments:              fragments,
	}
	title := normalizeText(calculator.Title)
	parts := []string{title, normalizeText(page.Title)}
	for _, value := range []*string{metadata.Description, metadata.AdditionalDescription} {
		if value != nil {
			parts = append(parts, *value)
		}
	}
	visibility := calculatorVisibility(calculator, page)
	access := domain.AccessModelFree
	parts = appendSearchFragments(parts, fragments, visibility, access)
	modified := page.ModifiedAt
	if calculator.ModifiedAt.After(modified) {
		modified = calculator.ModifiedAt
	}
	material := domain.Material{
		Key: domain.MaterialKey{
			Kind: domain.MaterialKindCalculator,
			ID:   calculator.ID,
		},
		Title:            title,
		Summary:          shortSummary(description),
		SearchText:       joinNormalizedText(parts...),
		Topics:           []domain.Label{},
		Tags:             []domain.Label{},
		Metadata:         metadata,
		Rating:           ratingSummary(calculator.RatingAverage, calculator.RatingCount),
		Visibility:       visibility,
		Access:           access,
		URL:              materialURL("calculators", page.Slug),
		SourceModifiedAt: modified.UTC(),
	}
	if page.PublishedAt != nil {
		published := page.PublishedAt.UTC()
		material.PublishedAt = &published
	}
	if err := material.Validate(); err != nil {
		return domain.Material{}, fmt.Errorf("calculator %s: material: %w", calculator.ID, err)
	}
	return material, nil
}

func calculatorVisibility(
	calculator model.Calculator,
	page model.CalculatorPage,
) domain.Visibility {
	if calculator.IsDeleted || page.IsDeleted || page.StatusID == draftStatusID {
		return domain.VisibilityHidden
	}
	if page.StatusID != publishedStatusID {
		return domain.VisibilityUnknown
	}
	return domain.VisibilityListed
}
