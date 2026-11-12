package domain

import (
	"errors"
	"fmt"
	"strings"
	"uuid"
)

type CourseMetadata struct {
	Target  *string
	Info    []CourseInfo
	Program []CourseProgramItem
}

type CourseInfo struct {
	ID          uuid.UUID
	Type        Label
	Title       string
	Description *string
	Position    *int32
	IsMain      bool
	TariffID    *uuid.UUID
}

func (m CourseMetadata) Validate() error {
	if m.Target != nil && strings.TrimSpace(*m.Target) == "" {
		return errors.New("target must be nonempty or nil")
	}
	seen := make(map[uuid.UUID]struct{}, len(m.Info))
	for i, info := range m.Info {
		if err := info.Validate(); err != nil {
			return fmt.Errorf("info item %d: %w", i, err)
		}
		if _, exists := seen[info.ID]; exists {
			return fmt.Errorf("info item %d: duplicate ID %s", i, info.ID)
		}
		seen[info.ID] = struct{}{}
	}
	if err := validateCourseProgram(m.Program); err != nil {
		return fmt.Errorf("program: %w", err)
	}
	return nil
}

func (info CourseInfo) Validate() error {
	if info.ID == uuid.Nil() {
		return errors.New("ID must not be nil UUID")
	}
	if err := validateLabel([]Label{info.Type}); err != nil {
		return fmt.Errorf("type: %w", err)
	}
	if strings.TrimSpace(info.Title) == "" {
		return errors.New("title must not be empty")
	}
	if info.Description != nil && strings.TrimSpace(*info.Description) == "" {
		return errors.New("description must be nonempty or nil")
	}
	if info.TariffID != nil && *info.TariffID == uuid.Nil() {
		return errors.New("tariff ID must not be nil UUID")
	}
	return nil
}
