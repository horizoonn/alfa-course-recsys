package converter

import (
	"fmt"
	"uuid"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/model"
)

var (
	courseAudienceInfoTypeID = uuid.MustParse("fafb8d36-dbc9-48ab-9e04-fd0266cd1d18")
	courseOutcomesInfoTypeID = uuid.MustParse("fbf98d7f-0bc7-4195-be27-be3129378a61")
)

func CourseToMaterial(data model.CourseData) (domain.Material, error) {
	course := data.Course
	target, err := optionalHTMLToText(course.TargetHTML)
	if err != nil {
		return domain.Material{}, fmt.Errorf("course %s: target: %w", course.ID, err)
	}
	program, err := courseProgram(course, data.Lessons)
	if err != nil {
		return domain.Material{}, fmt.Errorf("course %s: program: %w", course.ID, err)
	}
	details := &domain.CourseMetadata{
		Target:  target,
		Info:    courseInfoItems(data.Info),
		Program: program,
	}
	metadata := domain.MaterialMetadata{
		Description:            normalizeOptionalText(&course.Description),
		AdditionalDescription:  normalizeOptionalText(course.AdditionalDescription),
		AlternativeTitle:       normalizeOptionalText(course.AlternativeTitle),
		AlternativeDescription: normalizeOptionalText(course.AlternativeDescription),
		NavigationTitle:        normalizeOptionalText(course.NavigationTitle),
		AdditionalLabels:       additionalLabels(data.Industries),
		Course:                 details,
	}
	summary := metadata.Description
	if summary == nil {
		summary = metadata.AdditionalDescription
	}
	title := normalizeText(course.Title)
	parts := make([]string, 0, 4+2*len(details.Info))
	parts = append(parts, title)
	for _, value := range []*string{
		metadata.Description,
		metadata.AdditionalDescription,
		details.Target,
	} {
		if value != nil {
			parts = append(parts, *value)
		}
	}
	for _, infoTypeID := range []uuid.UUID{courseAudienceInfoTypeID, courseOutcomesInfoTypeID} {
		for _, info := range details.Info {
			if info.Type.ID != infoTypeID || info.TariffID != nil {
				continue
			}
			parts = append(parts, info.Title)
			if info.Description != nil {
				parts = append(parts, *info.Description)
			}
		}
	}
	material := domain.Material{
		Key: domain.MaterialKey{
			Kind: domain.MaterialKindCourse,
			ID:   course.ID,
		},
		Title:            title,
		Summary:          summary,
		SearchText:       joinNormalizedText(parts...),
		Topics:           topicLabels(data.Topics),
		Tags:             tagLabels(data.Tags),
		Metadata:         metadata,
		Duration:         materialDuration(course.DurationText),
		Rating:           ratingSummary(course.Rating, nil),
		Visibility:       courseVisibility(course),
		Access:           courseAccess(course),
		URL:              materialURL("courses", course.Slug),
		SourceModifiedAt: course.ModifiedAt.UTC(),
	}
	if course.PublishedAt != nil {
		publishedAt := course.PublishedAt.UTC()
		material.PublishedAt = &publishedAt
	}
	if err := material.Validate(); err != nil {
		return domain.Material{}, fmt.Errorf("course %s: material: %w", course.ID, err)
	}
	return material, nil
}

func courseInfoItems(items []model.CourseInfo) []domain.CourseInfo {
	result := make([]domain.CourseInfo, 0, len(items))
	for _, info := range items {
		if info.InfoTypeIsDeleted {
			continue
		}
		result = append(result, domain.CourseInfo{
			ID: info.ID,
			Type: domain.Label{
				ID:    info.InfoTypeID,
				Title: normalizeText(info.InfoTypeTitle),
			},
			Title:       normalizeText(info.Title),
			Description: normalizeOptionalText(info.Description),
			Position:    info.Position,
			IsMain:      info.IsMain,
			TariffID:    info.TariffID,
		})
	}
	return result
}

func courseAccess(course model.Course) domain.AccessModel {
	access := accessModel(course.AccessTypeID, course.SettingsAccessTypeID)
	switch access {
	case domain.AccessModelSubscription:
		if !course.HasActiveSubscription {
			return domain.AccessModelUnknown
		}
	case domain.AccessModelPaid:
		if !course.HasTariffs {
			return domain.AccessModelUnknown
		}
	}
	return access
}

func courseVisibility(course model.Course) domain.Visibility {
	if course.IsDeleted || course.IsClosed ||
		(course.SettingsIsDeleted != nil && *course.SettingsIsDeleted) {
		return domain.VisibilityHidden
	}
	if course.StatusID == draftStatusID ||
		(course.SettingsStatusID != nil && *course.SettingsStatusID == draftStatusID) {
		return domain.VisibilityHidden
	}
	if course.ContentSettingsID == nil || course.SettingsStatusID == nil ||
		course.SettingsIsDeleted == nil {
		return domain.VisibilityUnknown
	}
	if course.StatusID != publishedStatusID || *course.SettingsStatusID != publishedStatusID {
		return domain.VisibilityUnknown
	}
	return domain.VisibilityListed
}
