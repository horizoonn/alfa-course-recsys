package domain

import (
	"fmt"
	"strings"
	"uuid"
)

type ContentRole string

const (
	ContentRoleBody      ContentRole = "body"
	ContentRoleCaption   ContentRole = "caption"
	ContentRolePromotion ContentRole = "promotion"
	ContentRoleReference ContentRole = "reference"
)

type ContentFragment struct {
	Role     ContentRole
	ID       uuid.UUID
	Position *int32
	Text     string
	IsMain   *bool
}

func validateContentFragments(fragments []ContentFragment) error {
	seen := make(map[uuid.UUID]struct{}, len(fragments))
	for i, fragment := range fragments {
		if fragment.ID == uuid.Nil() {
			return fmt.Errorf("fragment %d: ID must not be nil UUID", i)
		}
		if strings.TrimSpace(fragment.Text) == "" {
			return fmt.Errorf("fragment %d: text must not be empty", i)
		}
		if _, exists := seen[fragment.ID]; exists {
			return fmt.Errorf("fragment %d: duplicate ID %s", i, fragment.ID)
		}
		switch fragment.Role {
		case "", ContentRoleBody, ContentRoleCaption, ContentRolePromotion, ContentRoleReference:
		default:
			return fmt.Errorf("fragment %d: unsupported role %q", i, fragment.Role)
		}
		seen[fragment.ID] = struct{}{}
	}
	return nil
}
