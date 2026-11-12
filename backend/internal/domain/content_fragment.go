package domain

import (
	"fmt"
	"strings"
	"uuid"
)

type ContentFragment struct {
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
		seen[fragment.ID] = struct{}{}
	}
	return nil
}
