package domain

import (
	"errors"
	"strings"
)

type TermMetadata struct {
	Letter *string
}

func (m TermMetadata) Validate() error {
	if m.Letter != nil && strings.TrimSpace(*m.Letter) == "" {
		return errors.New("letter must be nonempty or nil")
	}
	return nil
}
