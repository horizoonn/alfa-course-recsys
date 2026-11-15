package httpapi

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/service/recommendation"
	openapi "github.com/horizoonn/alfa-course-recsys/backend/internal/transport/http/generated"
)

type Recommender interface {
	Recommend(context.Context, recommendation.Request) (recommendation.Result, error)
}

type Handler struct {
	recommender Recommender
	ready       func() bool
	logger      *slog.Logger
}

var _ openapi.Handler = (*Handler)(nil)

func NewServer(ready func() bool, recommender Recommender, logger *slog.Logger) (http.Handler, error) {
	handler := &Handler{
		ready:       ready,
		recommender: recommender,
		logger:      logger,
	}
	server, err := openapi.NewServer(handler, openapi.WithErrorHandler(handler.handleError))
	if err != nil {
		return nil, err
	}
	return handler.middleware(server), nil
}

func (h *Handler) GetHealth(ctx context.Context) (openapi.GetHealthRes, error) {
	if !h.ready() {
		body := errorBody(ctx, openapi.ApiErrorCodeDATAUNAVAILABLE, "Данные для подбора недоступны.")
		return &openapi.DataUnavailableHeaders{
			CacheControl: openapi.NewOptDataUnavailableCacheControl("no-store"),
			XRequestID:   openapi.NewOptRequestId(body.RequestID),
			Response:     body,
		}, nil
	}
	return &openapi.HealthResponseHeaders{
		CacheControl: openapi.NewOptGetHealthOKCacheControl("no-store"),
		XRequestID:   openapi.NewOptRequestId(requestID(ctx)),
		Response:     openapi.HealthResponse{Status: "ok"},
	}, nil
}

func (h *Handler) GetMeta(ctx context.Context) (openapi.GetMetaRes, error) {
	return &openapi.MetaResponseHeaders{
		CacheControl: openapi.NewOptGetMetaOKCacheControl("no-store"),
		XRequestID:   openapi.NewOptRequestId(requestID(ctx)),
		Response:     metadata(),
	}, nil
}

func (h *Handler) ListDemoProfiles(ctx context.Context) (openapi.ListDemoProfilesRes, error) {
	return &openapi.DemoProfilesResponseHeaders{
		CacheControl: openapi.NewOptListDemoProfilesOKCacheControl("no-store"),
		XRequestID:   openapi.NewOptRequestId(requestID(ctx)),
		Response:     demoProfiles(),
	}, nil
}
