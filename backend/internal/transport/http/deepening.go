package httpapi

import (
	"context"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/service/recommendation"
	openapi "github.com/horizoonn/alfa-course-recsys/backend/internal/transport/http/generated"
)

func (h *Handler) Deepen(ctx context.Context, body *openapi.DeepeningRequest) (openapi.DeepenRes, error) {
	if !h.ready() || h.recommender == nil {
		return nil, recommendation.ErrSearchUnavailable
	}
	key, err := domain.ParseMaterialKey(body.MaterialKey)
	if err != nil {
		return nil, &domain.InputError{Fields: []domain.FieldError{{Field: "material_key", Message: "Неизвестный материал."}}}
	}
	request := recommendation.Request{}
	request.Query, _ = body.Query.Get()
	request.HistoryProfileID, _ = body.HistoryProfileID.Get()
	request.DemoProfileID, _ = body.DemoProfileID.Get()
	excluded := []domain.MaterialKey{}
	for _, value := range body.ExcludeKeys {
		excludedKey, err := domain.ParseMaterialKey(value)
		if err != nil {
			return nil, &domain.InputError{Fields: []domain.FieldError{{Field: "exclude_keys", Message: "Некорректный ключ материала."}}}
		}
		excluded = append(excluded, excludedKey)
	}
	result, err := h.recommender.Deepen(ctx, key, request, excluded)
	if err != nil {
		return nil, err
	}
	response := openapi.DeepeningResponse{RequestID: requestID(ctx), Items: []openapi.MaterialCard{}}
	for _, selected := range result.Items {
		response.Items = append(response.Items, materialCard(selected))
	}
	return &openapi.DeepeningResponseHeaders{CacheControl: openapi.NewOptDeepenOKCacheControl("no-store"), XRequestID: openapi.NewOptRequestId(response.RequestID), Response: response}, nil
}
