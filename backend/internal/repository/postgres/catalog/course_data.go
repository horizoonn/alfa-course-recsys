package catalog

import (
	"context"
	"fmt"
	"uuid"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/model"
)

func (r *Repository) loadCourseData(ctx context.Context) ([]model.CourseData, error) {
	courses, err := r.loadCourseRows(ctx)
	if err != nil {
		return nil, fmt.Errorf("load course rows: %w", err)
	}
	if len(courses) == 0 {
		return []model.CourseData{}, nil
	}

	courseIDs := make([]uuid.UUID, 0, len(courses))
	settingsIDs := make([]uuid.UUID, 0, len(courses))
	for _, course := range courses {
		courseIDs = append(courseIDs, course.ID)
		if course.ContentSettingsID != nil {
			settingsIDs = append(settingsIDs, *course.ContentSettingsID)
		}
	}

	relations, err := r.loadContentRelations(ctx, settingsIDs)
	if err != nil {
		return nil, fmt.Errorf("load course relations: %w", err)
	}
	info, err := r.loadCourseInfo(ctx, courseIDs)
	if err != nil {
		return nil, fmt.Errorf("load course info: %w", err)
	}

	lessons, err := r.loadCourseLessons(ctx, courseIDs)
	if err != nil {
		return nil, fmt.Errorf("load course lessons: %w", err)
	}
	lessonsByCourse := make(map[uuid.UUID][]model.CourseLesson, len(courses))
	for _, lesson := range lessons {
		lessonsByCourse[lesson.CourseID] = append(lessonsByCourse[lesson.CourseID], lesson)
	}

	infoByCourse := make(map[uuid.UUID][]model.CourseInfo, len(courses))
	for _, item := range info {
		infoByCourse[item.CourseID] = append(infoByCourse[item.CourseID], item)
	}

	result := make([]model.CourseData, 0, len(courses))
	for _, course := range courses {
		data := model.CourseData{
			Course:  course,
			Info:    infoByCourse[course.ID],
			Lessons: lessonsByCourse[course.ID],
		}
		if course.ContentSettingsID != nil {
			settingsID := *course.ContentSettingsID
			data.Topics = relations.topicsBySettings[settingsID]
			data.Tags = relations.tagsBySettings[settingsID]
			data.Industries = relations.industriesBySettings[settingsID]
		}
		result = append(result, data)
	}

	return result, nil
}
