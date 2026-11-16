package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/transport/http/generated"
)

type requestIDKey struct{}

func requestID(ctx context.Context) openapi.RequestId {
	if id, ok := ctx.Value(requestIDKey{}).(openapi.RequestId); ok {
		return id
	}
	return openapi.RequestId(uuid.New())
}

func (h *Handler) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		id := openapi.RequestId(uuid.New())
		ctx := context.WithValue(r.Context(), requestIDKey{}, id)
		w.Header().Set("X-Request-Id", uuid.UUID(id).String())
		w.Header().Set("Cache-Control", "no-store")
		defer func() {
			if recovered := recover(); recovered != nil {
				h.logger.Error("HTTP handler panicked", "request_id", uuid.UUID(id).String(), "panic", recovered)
				h.writeError(w, internalError(ctx))
			}
		}()
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
		path := r.URL.Path
		switch path {
		case "/api/meta", "/api/demo-profiles", "/api/history-profiles", "/api/health", "/api/recommendations", "/api/deepening":
		default:
			path = "unmatched"
		}
		h.logger.Info("HTTP request", "request_id", uuid.UUID(id).String(), "method", r.Method, "path", path, "duration", time.Since(started))
	})
}
