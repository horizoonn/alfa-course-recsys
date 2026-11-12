package converter

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"uuid"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/model"
)

func courseProgram(
	course model.Course,
	lessons []model.CourseLesson,
) ([]domain.CourseProgramItem, error) {
	byID := make(map[uuid.UUID]model.CourseLesson, len(lessons))
	for _, lesson := range lessons {
		if lesson.CourseID != course.ID {
			return nil, fmt.Errorf("lesson %s belongs to another course", lesson.LessonID)
		}
		if _, exists := byID[lesson.LessonID]; exists {
			return nil, fmt.Errorf("duplicate lesson %s", lesson.LessonID)
		}
		byID[lesson.LessonID] = lesson
	}
	roots := make([]model.CourseLesson, 0)
	children := make(map[uuid.UUID][]model.CourseLesson)
	for _, lesson := range lessons {
		if lesson.ParentGroupID == nil {
			roots = append(roots, lesson)
			continue
		}
		parent, exists := byID[*lesson.ParentGroupID]
		if !exists || !parent.IsGroup {
			return nil, fmt.Errorf("lesson %s has no parent group in this course", lesson.LessonID)
		}
		children[parent.LessonID] = append(children[parent.LessonID], lesson)
	}
	slices.SortFunc(roots, compareCourseLessons)
	for id := range children {
		slices.SortFunc(children[id], compareCourseLessons)
	}
	result := make([]domain.CourseProgramItem, 0, len(lessons))
	var visit func(model.CourseLesson, domain.Visibility)
	visit = func(lesson model.CourseLesson, parentVisibility domain.Visibility) {
		description := normalizeOptionalText(lesson.Description)
		title := normalizeText(lesson.Title)
		parts := []string{title}
		if description != nil {
			parts = append(parts, *description)
		}
		visibility := parentVisibility
		if lesson.BindingIsDeleted || lesson.LessonIsDeleted || lesson.IsClosed ||
			(!lesson.IsGroup && lesson.LessonTypeIsDeleted) {
			visibility = domain.VisibilityHidden
		} else if lesson.ClosedUntil != nil && visibility == domain.VisibilityListed {
			visibility = domain.VisibilityUnknown
		}
		item := domain.CourseProgramItem{
			ID:               lesson.LessonID,
			ParentGroupID:    lesson.ParentGroupID,
			Position:         lesson.Position,
			Title:            title,
			NavigationTitle:  normalizeOptionalText(lesson.NavigationTitle),
			Description:      description,
			SearchText:       joinNormalizedText(parts...),
			IsGroup:          lesson.IsGroup,
			Duration:         materialDuration(lesson.DurationText),
			Visibility:       visibility,
			IsTrial:          lesson.IsTrial,
			SourceModifiedAt: lesson.ModifiedAt.UTC(),
		}
		if !lesson.IsGroup {
			item.Rating = ratingSummary(lesson.RatingAverage, lesson.RatingCount)
			item.Format = &domain.Label{
				ID:    lesson.LessonTypeID,
				Title: normalizeText(lesson.LessonTypeTitle),
			}
		}
		if lesson.ClosedUntil != nil {
			closedUntil := lesson.ClosedUntil.UTC()
			item.ClosedUntil = &closedUntil
		}
		result = append(result, item)
		for _, child := range children[lesson.LessonID] {
			visit(child, visibility)
		}
	}
	for _, root := range roots {
		visit(root, courseVisibility(course))
	}
	if len(result) != len(lessons) {
		return nil, fmt.Errorf("course program contains a cycle")
	}
	return result, nil
}

func compareCourseLessons(a, b model.CourseLesson) int {
	if a.Position == nil && b.Position != nil {
		return 1
	}
	if a.Position != nil && b.Position == nil {
		return -1
	}
	if a.Position != nil && b.Position != nil {
		if order := cmp.Compare(*a.Position, *b.Position); order != 0 {
			return order
		}
	}
	return strings.Compare(a.LessonID.String(), b.LessonID.String())
}
