package converter

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"uuid"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/model"
)

func contentFragments(
	entityID uuid.UUID,
	blocks []model.ContentBlock,
) ([]domain.ContentFragment, error) {
	result := make([]domain.ContentFragment, 0, len(blocks))
	for _, block := range blocks {
		if block.EntityID != entityID {
			return nil, fmt.Errorf("block %s belongs to another material", block.ID)
		}
		text, err := contentBlockText(block)
		if err != nil {
			return nil, fmt.Errorf("block %s: %w", block.ID, err)
		}
		if text != "" {
			result = append(result, domain.ContentFragment{
				Role:     contentRole(block),
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
	parts := []string{}
	add := func(value string) error {
		text, err := htmlToText(value)
		if err != nil {
			return err
		}
		if text != "" {
			parts = append(parts, text)
		}
		return nil
	}
	if block.HTML != nil {
		if err := add(*block.HTML); err != nil {
			return "", err
		}
	}
	if block.ParamsJSON != nil {
		var object map[string]any
		if err := decodeBlockParams(block.ParamsJSON, &object); err != nil {
			return "", err
		}
		var visit func(any, bool) error
		visit = func(value any, readable bool) error {
			switch item := value.(type) {
			case string:
				if readable {
					return add(item)
				}
			case []any:
				for _, value := range item {
					if err := visit(value, readable); err != nil {
						return err
					}
				}
			case map[string]any:
				keys := make([]string, 0, len(item))
				for key := range item {
					keys = append(keys, key)
				}
				slices.Sort(keys)
				for _, key := range keys {
					text := false
					switch strings.ToLower(key) {
					case "title", "description", "text", "alt", "caption", "content", "question", "answer", "items", "label", "name", "filename":
						text = true
					}
					if err := visit(item[key], text); err != nil {
						return err
					}
				}
			}
			return nil
		}
		if err := visit(object, false); err != nil {
			return "", err
		}
	}
	return joinNormalizedText(parts...), nil
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

func contentRole(block model.ContentBlock) domain.ContentRole {
	var params map[string]any
	if block.ParamsJSON != nil {
		_ = json.Unmarshal([]byte(*block.ParamsJSON), &params)
	}
	var promoted func(any) bool
	promoted = func(value any) bool {
		switch item := value.(type) {
		case map[string]any:
			for key, child := range item {
				switch strings.ToLower(key) {
				case "advcode", "productcode", "metrictarget":
					if text, ok := child.(string); ok && strings.TrimSpace(text) != "" {
						return true
					}
				}
				if promoted(child) {
					return true
				}
			}
		case []any:
			for _, child := range item {
				if promoted(child) {
					return true
				}
			}
		}
		return false
	}
	if promoted(params) {
		return domain.ContentRolePromotion
	}
	switch block.Type {
	case 1, 5, 7:
		return domain.ContentRoleCaption
	case 8:
		return domain.ContentRoleReference
	default:
		return domain.ContentRoleBody
	}
}
