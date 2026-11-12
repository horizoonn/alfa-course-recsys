package domain

import (
	"errors"
	"fmt"
	"strings"
)

type MaterialMetadata struct {
	Category               *Label
	AdditionalLabels       []Label
	Description            *string
	AdditionalDescription  *string
	CardDescription        *string
	AlternativeTitle       *string
	AlternativeDescription *string
	NavigationTitle        *string
	Headings               []string
	Fragments              []ContentFragment
	Course                 *CourseMetadata
	File                   *MaterialFile
	Term                   *TermMetadata
	Webinar                *WebinarMetadata
}

func (m MaterialMetadata) Validate() error {
	if m.Category != nil {
		if err := validateLabel([]Label{*m.Category}); err != nil {
			return fmt.Errorf("category: %w", err)
		}
	}
	if err := validateLabel(m.AdditionalLabels); err != nil {
		return fmt.Errorf("additional labels: %w", err)
	}
	if m.Description != nil && strings.TrimSpace(*m.Description) == "" {
		return errors.New("description must be nonempty or nil")
	}
	if m.AdditionalDescription != nil && strings.TrimSpace(*m.AdditionalDescription) == "" {
		return errors.New("additional description must be nonempty or nil")
	}
	if m.CardDescription != nil && strings.TrimSpace(*m.CardDescription) == "" {
		return errors.New("card description must be nonempty or nil")
	}
	if m.AlternativeTitle != nil && strings.TrimSpace(*m.AlternativeTitle) == "" {
		return errors.New("alternative title must be nonempty or nil")
	}
	if m.AlternativeDescription != nil && strings.TrimSpace(*m.AlternativeDescription) == "" {
		return errors.New("alternative description must be nonempty or nil")
	}
	if m.NavigationTitle != nil && strings.TrimSpace(*m.NavigationTitle) == "" {
		return errors.New("navigation title must be nonempty or nil")
	}
	if err := validateContentFragments(m.Fragments); err != nil {
		return fmt.Errorf("fragments: %w", err)
	}
	if m.Course != nil {
		if err := m.Course.Validate(); err != nil {
			return fmt.Errorf("course: %w", err)
		}
	}
	if m.File != nil {
		if err := m.File.Validate(); err != nil {
			return fmt.Errorf("file: %w", err)
		}
	}
	if m.Term != nil {
		if err := m.Term.Validate(); err != nil {
			return fmt.Errorf("term: %w", err)
		}
	}
	if m.Webinar != nil {
		if err := m.Webinar.Validate(); err != nil {
			return fmt.Errorf("webinar: %w", err)
		}
	}
	return nil
}
