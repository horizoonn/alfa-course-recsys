package domain

import (
	"errors"
	"strings"
)

type MaterialFile struct {
	Extension *string
	SizeBytes *int64
}

func (f MaterialFile) Validate() error {
	if f.Extension != nil && strings.TrimSpace(*f.Extension) == "" {
		return errors.New("extension must be nonempty or nil")
	}
	if f.SizeBytes != nil && *f.SizeBytes < 0 {
		return errors.New("file size must not be negative")
	}
	return nil
}
