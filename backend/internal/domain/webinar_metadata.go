package domain

import (
	"errors"
	"fmt"
	"time"
)

type WebinarMetadata struct {
	StartsAt             *time.Time
	CompletedAt          *time.Time
	RegistrationRequired bool
}

func (m WebinarMetadata) Validate() error {
	if m.StartsAt != nil {
		if err := validateTimestamp(*m.StartsAt); err != nil {
			return fmt.Errorf("starts at: %w", err)
		}
	}
	if m.CompletedAt != nil {
		if err := validateTimestamp(*m.CompletedAt); err != nil {
			return fmt.Errorf("completed at: %w", err)
		}
	}
	if m.StartsAt != nil && m.CompletedAt != nil && m.CompletedAt.Before(*m.StartsAt) {
		return errors.New("completion must not precede start")
	}
	return nil
}
