package httpapi

import (
	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/transport/http/generated"
)

func metadata() openapi.MetaResponse {
	options := domain.ProfileMetadata()
	convert := func(values []domain.Option) []openapi.MetaOption {
		result := make([]openapi.MetaOption, 0, len(values))
		for _, v := range values {
			result = append(result, openapi.MetaOption{ID: v.ID, Label: v.Label})
		}
		return result
	}
	return openapi.MetaResponse{
		Tasks:      convert(options.Tasks),
		Stages:     convert(options.Stages),
		Industries: convert(options.Industries),
		Experience: convert(options.Experience),
		HelpModes:  convert(options.HelpModes),
	}
}

func demoProfiles() openapi.DemoProfilesResponse {
	result := openapi.DemoProfilesResponse{}
	for _, p := range domain.DemoProfiles() {
		result = append(result, openapi.DemoProfile{
			ID:          p.ID,
			Title:       p.Title,
			Description: p.Description,
			Profile:     profileDTO(p.Profile),
		})
	}
	return result
}

func profileDTO(p domain.Profile) openapi.Profile {
	result := openapi.Profile{}
	if p.TaskID != "" {
		result.TaskID = openapi.NewOptNilString(string(p.TaskID))
	}
	if p.Stage != "" {
		result.Stage = openapi.NewOptNilString(string(p.Stage))
	}
	if p.Industry != "" {
		result.Industry = openapi.NewOptNilString(string(p.Industry))
	}
	if p.Experience != "" {
		result.Experience = openapi.NewOptNilString(string(p.Experience))
	}
	if p.HelpMode != "" {
		result.HelpMode = openapi.NewOptNilProfileHelpMode(openapi.ProfileHelpMode(p.HelpMode))
	}
	return result
}
