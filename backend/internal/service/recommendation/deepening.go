package recommendation

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
)

type RelatedMatch struct {
	Key   string  `json:"key"`
	Score float64 `json:"score"`
}

type RelatedResult struct {
	CatalogID    string                `json:"catalog_id"`
	Algorithm    string                `json:"algorithm"`
	SourceKey    string                `json:"source_key"`
	Anchors      []string              `json:"anchors"`
	Matches      []RelatedMatch        `json:"matches"`
	FreeOnly     bool                  `json:"free_only"`
	ExcludeKinds []domain.MaterialKind `json:"exclude_kinds"`
	AllowedKinds []domain.MaterialKind `json:"allowed_kinds"`
}

type RelatedSearch interface {
	Related(context.Context, domain.MaterialKey, string) (RelatedResult, error)
}

func (s *Service) Deepen(ctx context.Context, key domain.MaterialKey, request Request, exclude []domain.MaterialKey) (Result, error) {
	result := Result{Items: []Item{}, Mode: "topic", Limitations: []Limitation{}, AlgorithmVersion: "related-content-v1"}
	index, exists := s.byKey[key]
	if !exists || s.materials[index].Visibility != domain.VisibilityListed {
		return result, invalid("material_key", "Материал недоступен.")
	}
	if utf8.RuneCountInString(request.Query) > 1000 {
		return result, invalid("query", "Не больше 1000 символов.")
	}
	if request.HistoryProfileID != "" && !slices.ContainsFunc(s.historyProfiles, func(p domain.HistoryProfile) bool { return p.ID == request.HistoryProfileID }) {
		return result, invalid("history_profile_id", "Неизвестный сценарий истории.")
	}
	if request.DemoProfileID != "" && (request.HistoryProfileID != "" || !slices.ContainsFunc(domain.DemoProfiles(), func(p domain.DemoProfile) bool { return p.ID == request.DemoProfileID })) {
		return result, invalid("demo_profile_id", "Неизвестный или несовместимый демо-профиль.")
	}
	historyID := request.HistoryProfileID
	if historyID == "" {
		historyID = request.DemoProfileID
	}
	excluded, _, _ := s.history(historyID)
	excluded[key] = true
	for _, item := range exclude {
		excluded[item] = true
	}
	provider, ok := s.search.(RelatedSearch)
	if !ok {
		return result, ErrSearchUnavailable
	}
	found, err := provider.Related(ctx, key, request.Query)
	if err != nil {
		return result, err
	}
	source := s.materials[index]
	usedURLs := map[string]bool{}
	if source.URL != nil {
		usedURLs[*source.URL] = true
	}
	for _, match := range found.Matches {
		targetKey, err := domain.ParseMaterialKey(match.Key)
		if err != nil {
			return result, ErrSearchUnavailable
		}
		targetIndex, ok := s.byKey[targetKey]
		if !ok {
			return result, ErrSearchUnavailable
		}
		target := s.materials[targetIndex]
		if excluded[targetKey] || target.Visibility != domain.VisibilityListed || target.URL == nil || usedURLs[*target.URL] {
			continue
		}
		if found.FreeOnly && target.Access != domain.AccessModelFree {
			continue
		}
		if slices.Contains(found.ExcludeKinds, targetKey.Kind) || (len(found.AllowedKinds) > 0 && !slices.Contains(found.AllowedKinds, targetKey.Kind)) {
			continue
		}
		reason, ok := deepeningReason(source, target, found.Anchors)
		if !ok {
			continue
		}
		result.Items = append(result.Items, item(target, []Reason{reason}))
		usedURLs[*target.URL] = true
		if len(result.Items) == 2 {
			break
		}
	}
	return result, nil
}

func deepeningReason(source, target domain.Material, anchors []string) (Reason, bool) {
	if target.Metadata.Course != nil && source.Metadata.Course == nil {
		lessons := 0
		section := ""
		for _, lesson := range target.Metadata.Course.Program {
			if lesson.Visibility != domain.VisibilityListed || lesson.IsGroup {
				continue
			}
			lessons++
			if section == "" && sectionMatches(lesson.Title+" "+lesson.SearchText, anchors) {
				section = lesson.Title
			}
		}
		if lessons >= 2 && section != "" {
			return Reason{Code: "course_section", Text: fmt.Sprintf("В программе курса есть раздел «%s». Курс позволяет разобрать тему системно.", section)}, true
		}
	}
	if target.Key.Kind != domain.MaterialKindArticle {
		return Reason{}, false
	}
	if source.Key.Kind == domain.MaterialKindTerm && sectionMatches(target.Title, anchors) {
		return Reason{Code: "deepening_topic", Text: "Подробный разбор темы после краткого определения."}, true
	}
	sourceSize, targetSize := materialTextSize(source), materialTextSize(target)
	if targetSize < 250 || targetSize*10 < sourceSize*14 {
		return Reason{}, false
	}
	for _, heading := range target.Metadata.Headings {
		if sectionMatches(heading, anchors) && !slices.Contains(source.Metadata.Headings, heading) && !strings.EqualFold(heading, source.Title) {
			return Reason{Code: "deepening_topic", Text: fmt.Sprintf("Более подробный материал с разделом «%s» по этой теме.", heading)}, true
		}
	}
	return Reason{}, false
}

func sectionMatches(section string, anchors []string) bool {
	words := strings.FieldsFunc(strings.ToLower(strings.ReplaceAll(section, "ё", "е")), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
	count := 0
	for _, anchor := range anchors {
		if len([]rune(anchor)) < 2 {
			continue
		}
		for _, word := range words {
			matches := strings.HasPrefix(word, anchor)
			if len([]rune(anchor)) <= 3 {
				matches = word == anchor
			}
			if matches {
				count++
				break
			}
		}
	}
	return count >= min(2, len(anchors)) && count > 0
}

func materialTextSize(material domain.Material) int {
	size := 0
	for _, document := range Documents([]domain.Material{material}) {
		for _, field := range document.Fields {
			if field.Name == "body" {
				size += len(strings.Fields(field.Text))
			}
		}
	}
	return size
}
