package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/ogen-go/ogen/ogenerrors"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
	openapi "github.com/horizoonn/alfa-course-recsys/backend/internal/transport/http/generated"
)

func errorBody(ctx context.Context, code openapi.ApiErrorCode, message string) openapi.ErrorResponse {
	return openapi.ErrorResponse{RequestID: requestID(ctx), Error: openapi.ApiError{
		Code: code, Message: message, Fields: []openapi.FieldError{},
	}}
}

func internalError(ctx context.Context) *openapi.UnexpectedErrorStatusCodeWithHeaders {
	body := errorBody(ctx, openapi.ApiErrorCodeINTERNALERROR, "Не удалось выполнить запрос. Повторите попытку позже.")
	return &openapi.UnexpectedErrorStatusCodeWithHeaders{
		StatusCode:   http.StatusInternalServerError,
		CacheControl: openapi.NewOptUnexpectedErrorCacheControl("no-store"),
		XRequestID:   openapi.NewOptRequestId(body.RequestID),
		Response:     body,
	}
}

func (h *Handler) NewError(ctx context.Context, err error) *openapi.UnexpectedErrorStatusCodeWithHeaders {
	var invalid *domain.InputError
	if errors.As(err, &invalid) {
		response := internalError(ctx)
		response.StatusCode = http.StatusBadRequest
		response.Response.Error.Code = openapi.ApiErrorCodeINVALIDINPUT
		response.Response.Error.Message = "Проверьте параметры запроса."
		for _, field := range invalid.Fields {
			response.Response.Error.Fields = append(response.Response.Error.Fields, openapi.FieldError{
				Field:   field.Field,
				Message: field.Message,
			})
		}
		h.logger.Warn("invalid request", "request_id", uuid.UUID(response.Response.RequestID).String(), "code", "INVALID_INPUT")
		return response
	}
	response := internalError(ctx)
	h.logger.Error("HTTP request failed", "request_id", uuid.UUID(response.Response.RequestID).String(), "error", err)
	return response
}

func (h *Handler) handleError(ctx context.Context, w http.ResponseWriter, _ *http.Request, err error) {
	var decode *ogenerrors.DecodeRequestError
	if errors.As(err, &decode) || ogenerrors.ErrorCode(err) < 500 {
		err = &domain.InputError{
			Fields: []domain.FieldError{{
				Field:   "body",
				Message: "Некорректный JSON или параметры запроса.",
			}}}
	}
	h.writeError(w, h.NewError(ctx, err))
}

func (h *Handler) writeError(w http.ResponseWriter, response *openapi.UnexpectedErrorStatusCodeWithHeaders) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Request-Id", uuid.UUID(response.Response.RequestID).String())
	w.WriteHeader(response.StatusCode)
	if err := json.NewEncoder(w).Encode(response.Response); err != nil {
		h.logger.Error("write HTTP error", "error", err)
	}
}
