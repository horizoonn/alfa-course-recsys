package domain

import (
	"fmt"
	"strings"
)

type (
	TaskID          string
	BusinessStage   string
	Industry        string
	ExperienceLevel string
)

type TaskMatch string

const (
	TaskMatchDirect  TaskMatch = "direct"
	TaskMatchBroader TaskMatch = "broader"
)

type HelpMode string

const (
	HelpModeQuickAnswer  HelpMode = "quick_answer"
	HelpModeTool         HelpMode = "tool"
	HelpModeDeepLearning HelpMode = "deep_learning"
)

type TaskFit struct {
	TaskID TaskID
	Match  TaskMatch
}

type MaterialFeatures struct {
	TaskFits         []TaskFit
	BusinessStages   []BusinessStage
	Industries       []Industry
	ExperienceLevels []ExperienceLevel
	HelpModes        []HelpMode
}

func (f MaterialFeatures) Validate() error {
	tasks := make(map[TaskID]struct{}, len(f.TaskFits))
	for i, fit := range f.TaskFits {
		if strings.TrimSpace(string(fit.TaskID)) == "" {
			return fmt.Errorf("task fits[%d]: task ID must not be empty", i)
		}
		switch fit.Match {
		case TaskMatchDirect, TaskMatchBroader:
		default:
			return fmt.Errorf("task fits[%d]: unsupported match %q", i, fit.Match)
		}
		if _, exists := tasks[fit.TaskID]; exists {
			return fmt.Errorf("task fits[%d]: duplicate task ID %q", i, fit.TaskID)
		}
		tasks[fit.TaskID] = struct{}{}
	}
	if err := validateCodes(f.BusinessStages); err != nil {
		return fmt.Errorf("business stages: %w", err)
	}
	if err := validateCodes(f.Industries); err != nil {
		return fmt.Errorf("industries: %w", err)
	}
	if err := validateCodes(f.ExperienceLevels); err != nil {
		return fmt.Errorf("experience levels: %w", err)
	}
	if err := validateCodes(f.HelpModes); err != nil {
		return fmt.Errorf("help modes: %w", err)
	}
	for i, mode := range f.HelpModes {
		switch mode {
		case HelpModeQuickAnswer, HelpModeTool, HelpModeDeepLearning:
		default:
			return fmt.Errorf("help modes[%d]: unsupported mode %q", i, mode)
		}
	}
	return nil
}

func validateCodes[T ~string](codes []T) error {
	seen := make(map[T]struct{}, len(codes))
	for i, code := range codes {
		if strings.TrimSpace(string(code)) == "" {
			return fmt.Errorf("item %d: code must not be empty", i)
		}
		if _, exists := seen[code]; exists {
			return fmt.Errorf("item %d: duplicate code %q", i, code)
		}
		seen[code] = struct{}{}
	}
	return nil
}
