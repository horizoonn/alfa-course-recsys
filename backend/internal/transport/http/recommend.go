package httpapi

import (
	"context"
	"strings"

	"golang.org/x/net/html"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/service/recommendation"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/transport/http/generated"
)

func (h *Handler) Recommend(ctx context.Context, body *openapi.RecommendationsRequest) (openapi.RecommendRes, error) {
	if !h.ready() || h.recommender == nil {
		result := errorBody(ctx, openapi.ApiErrorCodeDATAUNAVAILABLE, "Подбор временно недоступен.")
		return &openapi.DataUnavailableHeaders{
			CacheControl: openapi.NewOptDataUnavailableCacheControl("no-store"),
			XRequestID:   openapi.NewOptRequestId(result.RequestID),
			Response:     result,
		}, nil
	}
	request := recommendation.Request{}
	if body.DemoProfileID.Set && (body.Profile.Set || body.Query.Set || body.HistoryProfileID.Set) {
		return nil, &domain.InputError{Fields: []domain.FieldError{{Field: "body", Message: "Демо-профиль нельзя совмещать со своим профилем или запросом."}}}
	}
	request.HistoryProfileID, _ = body.HistoryProfileID.Get()
	request.Query, _ = body.Query.Get()
	request.DemoProfileID, _ = body.DemoProfileID.Get()
	value, _ := body.Profile.Get()
	task, _ := value.TaskID.Get()
	stage, _ := value.Stage.Get()
	industry, _ := value.Industry.Get()
	experience, _ := value.Experience.Get()
	mode, _ := value.HelpMode.Get()
	request.Profile = domain.Profile{
		TaskID: domain.TaskID(task), Stage: domain.BusinessStage(stage),
		Industry: domain.Industry(industry), Experience: domain.ExperienceLevel(experience),
		HelpMode: domain.HelpMode(mode),
	}
	limit := body.Limit
	if limit.Set {
		request.Limit = int(limit.Value)
	}
	result, err := h.recommender.Recommend(ctx, request)
	if err != nil {
		return nil, err
	}
	h.logger.Info("recommendations ranked", "request_id", requestID(ctx), "algorithm", result.AlgorithmVersion, "items", len(result.Items))
	response := openapi.RecommendationsResponse{
		RequestID:   requestID(ctx),
		Mode:        openapi.RecommendationMode(result.Mode),
		Items:       []openapi.MaterialCard{},
		Limitations: []openapi.Limitation{},
	}
	for _, selected := range result.Items {
		response.Items = append(response.Items, materialCard(selected))
	}
	for _, limitation := range result.Limitations {
		response.Limitations = append(response.Limitations, openapi.Limitation{
			Code:    openapi.LimitationCode(limitation.Code),
			Message: limitation.Message,
		})
	}
	return &openapi.RecommendationsResponseHeaders{
		CacheControl: openapi.NewOptRecommendOKCacheControl("no-store"),
		XRequestID:   openapi.NewOptRequestId(response.RequestID),
		Response:     response,
	}, nil
}

func safeSummary(value *string) openapi.NilString {
	if value == nil {
		return openapi.NilString{Null: true}
	}
	document, err := html.Parse(strings.NewReader(*value))
	if err != nil {
		return openapi.NilString{Null: true}
	}
	var builder strings.Builder
	var write func(*html.Node)
	write = func(node *html.Node) {
		if node.Type == html.ElementNode && (node.Data == "head" || node.Data == "script" ||
			node.Data == "style" || node.Data == "template") {
			return
		}
		if node.Type == html.TextNode {
			builder.WriteString(node.Data)
		}
		block := node.Type == html.ElementNode && (node.Data == "p" || node.Data == "div" ||
			node.Data == "li" || node.Data == "br" || node.Data == "h1" || node.Data == "h2")
		if block {
			builder.WriteByte(' ')
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			write(child)
		}
		if block {
			builder.WriteByte(' ')
		}
	}
	write(document)
	text := strings.Join(strings.Fields(builder.String()), " ")
	if text == "" {
		return openapi.NilString{Null: true}
	}
	runes := []rune(text)
	if len(runes) > 300 {
		end := 299
		for end > 0 && runes[end] != ' ' {
			end--
		}
		if end == 0 {
			end = 299
		}
		text = strings.TrimSpace(string(runes[:end])) + "…"
	}
	return openapi.NewNilString(text)
}

func nextAction(item recommendation.Item) openapi.NextAction {
	if item.URL == nil {
		return openapi.NextAction{
			Type:  "refine_request",
			Label: "Изменить подборку",
		}
	}
	if item.Access == domain.AccessModelPaid || item.Access == domain.AccessModelSubscription {
		return openapi.NextAction{
			Type:  "open_purchase",
			Label: "Посмотреть условия доступа",
		}
	}
	if item.Access == domain.AccessModelUnknown {
		return openapi.NextAction{
			Type:  "open_material",
			Label: "Проверить условия доступа",
		}
	}
	labels := map[domain.MaterialKind]string{
		domain.MaterialKindArticle:    "Читать статью",
		domain.MaterialKindCourse:     "Изучить курс",
		domain.MaterialKindDocument:   "Открыть документ",
		domain.MaterialKindCalculator: "Открыть инструмент",
		domain.MaterialKindTerm:       "Открыть определение",
		domain.MaterialKindPodcast:    "Слушать подкаст",
		domain.MaterialKindWebinar:    "Открыть страницу вебинара",
	}
	label, ok := labels[item.Key.Kind]
	if !ok {
		label = "Открыть материал"
	}
	return openapi.NextAction{
		Type:  "open_material",
		Label: label,
	}
}

func materialCard(selected recommendation.Item) openapi.MaterialCard {
	reasons := make([]openapi.RecommendationReason, 0, len(selected.Reasons))
	for _, reason := range selected.Reasons {
		reasons = append(reasons, openapi.RecommendationReason{
			Code: openapi.RecommendationReasonCode(reason.Code),
			Text: reason.Text,
		})
	}
	url := openapi.NilString{Null: true}
	if selected.URL != nil {
		url = openapi.NewNilString(*selected.URL)
	}
	return openapi.MaterialCard{
		MaterialKey: selected.Key.String(),
		Type:        openapi.MaterialKind(selected.Key.Kind),
		Title:       selected.Title,
		SafeSummary: safeSummary(selected.Summary),
		Access:      openapi.AccessModel(selected.Access),
		SourceURL:   url,
		Reasons:     reasons,
		NextAction:  nextAction(selected),
	}
}
