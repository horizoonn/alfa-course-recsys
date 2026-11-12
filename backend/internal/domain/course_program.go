package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"uuid"
)

type CourseProgramItem struct {
	ID               uuid.UUID
	ParentGroupID    *uuid.UUID
	Position         *int32
	Title            string
	NavigationTitle  *string
	Description      *string
	SearchText       string
	IsGroup          bool
	Format           *Label
	Duration         MaterialDuration
	Rating           *RatingSummary
	Visibility       Visibility
	IsTrial          bool
	ClosedUntil      *time.Time
	SourceModifiedAt time.Time
}

func validateCourseProgram(items []CourseProgramItem) error {
	seen := make(map[uuid.UUID]CourseProgramItem, len(items))
	for i, item := range items {
		if item.ID == uuid.Nil() {
			return fmt.Errorf("item %d: ID must not be nil UUID", i)
		}
		if _, exists := seen[item.ID]; exists {
			return fmt.Errorf("item %d: duplicate ID %s", i, item.ID)
		}
		if item.ParentGroupID != nil {
			parent, exists := seen[*item.ParentGroupID]
			if !exists || !parent.IsGroup {
				return fmt.Errorf("item %d: parent must be an earlier group in this course", i)
			}
			if parent.Visibility == VisibilityHidden && item.Visibility != VisibilityHidden {
				return fmt.Errorf("item %d: child of a hidden group must be hidden", i)
			}
			if parent.Visibility == VisibilityUnknown && item.Visibility == VisibilityListed {
				return fmt.Errorf("item %d: child of an unknown group must not be listed", i)
			}
		}
		if err := item.Validate(); err != nil {
			return fmt.Errorf("item %d: %w", i, err)
		}
		seen[item.ID] = item
	}
	return nil
}

func (item CourseProgramItem) Validate() error {
	if item.ID == uuid.Nil() {
		return errors.New("ID must not be nil UUID")
	}
	if strings.TrimSpace(item.Title) == "" || strings.TrimSpace(item.SearchText) == "" {
		return errors.New("title and search text must not be empty")
	}
	if item.NavigationTitle != nil && strings.TrimSpace(*item.NavigationTitle) == "" {
		return errors.New("navigation title must be nonempty or nil")
	}
	if item.Description != nil && strings.TrimSpace(*item.Description) == "" {
		return errors.New("description must be nonempty or nil")
	}
	if item.IsGroup && item.Rating != nil {
		return errors.New("group must not have a rating")
	}
	if item.IsGroup && item.Format != nil {
		return errors.New("group must not have a lesson format")
	}
	if !item.IsGroup && item.Format == nil {
		return errors.New("lesson format must be present")
	}
	if item.Format != nil {
		if err := validateLabel([]Label{*item.Format}); err != nil {
			return fmt.Errorf("format: %w", err)
		}
	}
	if err := item.Duration.Validate(); err != nil {
		return fmt.Errorf("duration: %w", err)
	}
	if item.Rating != nil {
		if err := item.Rating.Validate(); err != nil {
			return fmt.Errorf("rating: %w", err)
		}
	}
	switch item.Visibility {
	case VisibilityListed, VisibilityHidden, VisibilityUnknown:
	default:
		return fmt.Errorf("unsupported visibility %q", item.Visibility)
	}
	if item.ClosedUntil != nil {
		if err := validateTimestamp(*item.ClosedUntil); err != nil {
			return fmt.Errorf("closed until: %w", err)
		}
	}
	return validateTimestamp(item.SourceModifiedAt)
}
