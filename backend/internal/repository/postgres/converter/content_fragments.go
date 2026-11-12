package converter

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"uuid"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/model"
)

const (
	contentTextBlock    int32 = 0
	contentNoteBlock    int32 = 2
	contentQuoteBlock   int32 = 3
	contentListBlock    int32 = 4
	contentDetailsBlock int32 = 13
)

func contentFragments(
	entityID uuid.UUID,
	blocks []model.ContentBlock,
	allowedTypes ...int32,
) ([]domain.ContentFragment, error) {
	result := make([]domain.ContentFragment, 0, len(blocks))
	for _, block := range blocks {
		if block.EntityID != entityID {
			return nil, fmt.Errorf("block %s belongs to another material", block.ID)
		}
		allowed := false
		for _, kind := range allowedTypes {
			if kind == block.Type {
				allowed = true
				break
			}
		}
		if !allowed {
			continue
		}
		text, err := contentBlockText(block)
		if err != nil {
			return nil, fmt.Errorf("block %s: %w", block.ID, err)
		}
		if text != "" {
			result = append(result, domain.ContentFragment{
				ID:       block.ID,
				Position: block.Position,
				Text:     text,
			})
		}
	}
	return result, nil
}

func decodeBlockParams(value *string, target any) error {
	if value == nil {
		return nil
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal([]byte(*value), &object); err != nil {
		return err
	}
	if object == nil {
		return errors.New("expected a parameters object")
	}
	return json.Unmarshal([]byte(*value), target)
}

func contentBlockText(block model.ContentBlock) (string, error) {
	if block.Type == contentTextBlock {
		text, err := optionalHTMLToText(block.HTML)
		if text == nil || err != nil {
			return "", err
		}
		return *text, nil
	}
	parts := []string{}
	add := func(value *string) error {
		text, err := optionalHTMLToText(value)
		if err != nil {
			return err
		}
		if text != nil {
			parts = append(parts, *text)
		}
		return nil
	}
	switch block.Type {
	case contentNoteBlock:
		var params struct{ Title, Description *string }
		if err := decodeBlockParams(block.ParamsJSON, &params); err != nil {
			return "", err
		}
		if err := add(params.Title); err != nil {
			return "", err
		}
		if err := add(params.Description); err != nil {
			return "", err
		}
	case contentQuoteBlock:
		var params struct{ Text, Alt *string }
		if err := decodeBlockParams(block.ParamsJSON, &params); err != nil {
			return "", err
		}
		if err := add(params.Text); err != nil {
			return "", err
		}
		if err := add(params.Alt); err != nil {
			return "", err
		}
	case contentListBlock:
		var params struct {
			Title *string
			Items []*string
		}
		if err := decodeBlockParams(block.ParamsJSON, &params); err != nil {
			return "", err
		}
		if err := add(params.Title); err != nil {
			return "", err
		}
		for i, item := range params.Items {
			if item == nil {
				return "", fmt.Errorf("list item %d: expected a string", i)
			}
			if err := add(item); err != nil {
				return "", err
			}
		}
	case contentDetailsBlock:
		var params struct{ Title, Text *string }
		if err := decodeBlockParams(block.ParamsJSON, &params); err != nil {
			return "", err
		}
		if err := add(params.Title); err != nil {
			return "", err
		}
		if err := add(params.Text); err != nil {
			return "", err
		}
	default:
		return "", nil
	}
	return strings.Join(parts, "\n\n"), nil
}

func appendSearchFragments(
	parts []string,
	fragments []domain.ContentFragment,
	visibility domain.Visibility,
	access domain.AccessModel,
) []string {
	if visibility != domain.VisibilityListed || access != domain.AccessModelFree {
		return parts
	}
	for _, fragment := range fragments {
		parts = append(parts, fragment.Text)
	}
	return parts
}
