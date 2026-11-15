package converter

import (
	"fmt"
	"strings"
	"uuid"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/model"
)

func detailFragments(
	entityID uuid.UUID,
	details []model.ContentDetail,
) ([]domain.ContentFragment, error) {
	fragments := make([]domain.ContentFragment, 0, len(details))
	for _, detail := range details {
		if detail.EntityID != entityID {
			return nil, fmt.Errorf("detail %s belongs to another material", detail.ID)
		}
		title, err := htmlToText(detail.Title)
		if err != nil {
			return nil, fmt.Errorf("detail %s: title: %w", detail.ID, err)
		}
		description, err := optionalHTMLToText(detail.Description)
		if err != nil {
			return nil, fmt.Errorf("detail %s: description: %w", detail.ID, err)
		}
		parts := []string{}
		if title != "" {
			parts = append(parts, title)
		}
		if description != nil {
			parts = append(parts, *description)
		}
		if len(parts) == 0 {
			continue
		}
		main := detail.IsMain
		fragments = append(fragments, domain.ContentFragment{
			ID:       detail.ID,
			Position: detail.Position,
			Text:     strings.Join(parts, "\n\n"),
			IsMain:   &main,
		})
	}
	return fragments, nil
}
