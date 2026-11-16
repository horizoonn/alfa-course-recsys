package httpapi

import (
	"context"

	openapi "github.com/horizoonn/alfa-course-recsys/backend/internal/transport/http/generated"
)

func (h *Handler) ListHistoryProfiles(ctx context.Context) (openapi.ListHistoryProfilesRes, error) {
	if !h.ready() || h.recommender == nil {
		body := errorBody(ctx, openapi.ApiErrorCodeDATAUNAVAILABLE, "Сценарии истории временно недоступны.")
		return &openapi.DataUnavailableHeaders{
			CacheControl: openapi.NewOptDataUnavailableCacheControl("no-store"),
			XRequestID:   openapi.NewOptRequestId(body.RequestID), Response: body,
		}, nil
	}
	profiles := openapi.HistoryProfilesResponse{}
	for _, profile := range h.recommender.HistoryProfiles() {
		profiles = append(profiles, openapi.HistoryProfile{ID: profile.ID, Title: profile.Title, Description: profile.Description,
			CompletedCourses: profile.CompletedCourses, LearningCourses: profile.LearningCourses, PositiveRatings: profile.PositiveRatings, NegativeRatings: profile.NegativeRatings})
	}
	return &openapi.HistoryProfilesResponseHeaders{CacheControl: openapi.NewOptListHistoryProfilesOKCacheControl("no-store"), XRequestID: openapi.NewOptRequestId(requestID(ctx)), Response: profiles}, nil
}
