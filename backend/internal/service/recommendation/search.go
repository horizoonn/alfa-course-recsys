package recommendation

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
)

var ErrSearchUnavailable = errors.New("text search unavailable")

type TextSearch interface {
	Search(context.Context, string) (SearchResult, error)
}

type SearchMatch struct {
	Word      float64  `json:"word"`
	Character float64  `json:"character"`
	Goals     []string `json:"goals"`
	Key       string   `json:"key"`
	Score     float64  `json:"score"`
	Rank      int      `json:"rank"`
	Lexical   float64  `json:"lexical"`
	Semantic  *float64 `json:"semantic"`
}

type SearchResult struct {
	QueryView    json.RawMessage       `json:"query_view"`
	AllowedKinds []domain.MaterialKind `json:"allowed_kinds"`
	CatalogID    string                `json:"catalog_id"`
	IndexID      string                `json:"index_id"`
	Algorithm    string                `json:"algorithm"`
	Mode         string                `json:"mode"`
	State        string                `json:"state"`
	Message      string                `json:"message"`
	Matches      []SearchMatch         `json:"matches"`
	Limitations  []SearchLimitation    `json:"limitations"`
	FreeOnly     bool                  `json:"free_only"`
	ExcludeKinds []domain.MaterialKind `json:"exclude_kinds"`
	ElapsedMS    float64               `json:"elapsed_ms"`
}

type SearchLimitation struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (s *Service) recommendQuery(ctx context.Context, query string, profile domain.Profile, completed map[domain.MaterialKey]bool, historyTopics map[string]bool, limit int) (Result, error) {
	if s.search == nil {
		return Result{}, ErrSearchUnavailable
	}
	found, err := s.search.Search(ctx, query)
	if err != nil {
		return Result{}, err
	}
	result := Result{Items: []Item{}, Mode: "fallback", Limitations: []Limitation{}, AlgorithmVersion: found.Algorithm}
	for _, limitation := range found.Limitations {
		result.addLimitation(limitation.Code, limitation.Message)
	}
	if found.State == "unclear" {
		result.addLimitation("QUERY_UNCLEAR", found.Message)
		return result, nil
	}
	ranked := []candidate{}
	for _, match := range found.Matches {
		key, err := domain.ParseMaterialKey(match.Key)
		if err != nil {
			return Result{}, ErrSearchUnavailable
		}
		index, ok := s.byKey[key]
		if !ok {
			return Result{}, ErrSearchUnavailable
		}
		material := s.materials[index]
		if material.Visibility != domain.VisibilityListed || completed[key] {
			continue
		}
		if found.FreeOnly && material.Access != domain.AccessModelFree {
			continue
		}
		excluded := false
		for _, kind := range found.ExcludeKinds {
			excluded = excluded || key.Kind == kind
		}
		if len(found.AllowedKinds) > 0 {
			allowed := false
			for _, kind := range found.AllowedKinds {
				allowed = allowed || key.Kind == kind
			}
			excluded = excluded || !allowed
		}
		if excluded {
			continue
		}
		signals := matchSignals(material, profile, historyTopics)
		context := float64(2*signals.m + signals.s + signals.e + signals.i + signals.h)
		ranked = append(ranked, candidate{index: index, score: match.Score * (1 + 0.1*context), signals: signals, textMatch: true})
	}
	s.sortCandidates(ranked, false)
	seen := map[string]bool{}
	for _, candidate := range ranked {
		material := s.materials[candidate.index]
		card := material.Key.String()
		if material.URL != nil {
			card = *material.URL + "\n" + strings.ToLower(analysisText(material.Title))
		}
		if seen[card] {
			continue
		}
		seen[card] = true
		reasons := reasonsFor(material, profile, candidate)
		reasons[0].Text = "Текст материала совпадает с вашим запросом."
		if found.Mode != "tfidf" {
			reasons[0].Text = "Описание материала близко по смыслу вашему запросу."
		}
		result.Items = append(result.Items, item(material, reasons))
		if result.Mode != "personalized" {
			result.Mode = "topic"
		}
		if candidate.signals != (signals{}) {
			result.Mode = "personalized"
		}
		if material.URL == nil {
			result.addLimitation("URL_UNVERIFIED", "Для части материалов нет подтверждённой ссылки.")
		}
		if len(result.Items) == limit {
			break
		}
	}
	if len(result.Items) == 0 {
		result.addLimitation("NO_MATCHES", "Подходящих материалов не нашлось. Опишите конкретное действие или выберите тему.")
	}
	return result, nil
}
