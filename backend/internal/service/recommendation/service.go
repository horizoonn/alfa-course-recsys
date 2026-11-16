package recommendation

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
)

type Request struct {
	HistoryProfileID string
	Query            string
	Profile          domain.Profile
	DemoProfileID    string
	Limit            int
}

type Reason struct {
	Code string
	Text string
}

type Limitation struct {
	Code    string
	Message string
}

type Item struct {
	Key     domain.MaterialKey
	Title   string
	Summary *string
	Access  domain.AccessModel
	URL     *string
	Reasons []Reason
}
type Result struct {
	Items            []Item
	Mode             string
	Limitations      []Limitation
	AlgorithmVersion string
}

type Service struct {
	historyProfiles []domain.HistoryProfile
	search          TextSearch
	materials       []domain.Material
	byKey           map[domain.MaterialKey]int
	starter         []domain.MaterialKey
	histories       map[string]domain.History
	textScores      map[domain.TaskID]map[domain.MaterialKey]float64
}

type signals struct{ m, s, e, i, h int }
type candidate struct {
	index, t  int
	score     float64
	signals   signals
	textMatch bool
}

func NewService(
	materials []domain.Material,
	starter []domain.MaterialKey,
	histories map[string]domain.History,
	scores TextScores,
	search TextSearch,
	profiles []domain.HistoryProfile,
) (*Service, error) {
	service := &Service{
		materials:       materials,
		byKey:           map[domain.MaterialKey]int{},
		starter:         slices.Clone(starter),
		histories:       map[string]domain.History{},
		historyProfiles: slices.Clone(profiles),
		textScores:      scores,
	}
	for id, history := range histories {
		service.histories[id] = history
	}
	for _, profile := range profiles {
		service.histories[profile.ID] = profile.History
	}
	service.search = search
	for i, material := range materials {
		if _, ok := service.byKey[material.Key]; ok {
			return nil, errors.New("duplicate material key in recommendation catalog")
		}
		service.byKey[material.Key] = i
	}

	return service, nil
}

func (s *Service) Recommend(ctx context.Context, request Request) (Result, error) {
	return s.recommend(ctx, request, false)
}
func (s *Service) Baseline(ctx context.Context, request Request) (Result, error) {
	return s.recommend(ctx, request, true)
}

func (s *Service) recommend(ctx context.Context, request Request, baseline bool) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	request.Query = strings.TrimSpace(request.Query)
	if utf8.RuneCountInString(request.Query) > 1000 {
		return Result{}, invalid("query", "Не больше 1000 символов.")
	}
	profile := request.Profile
	if request.DemoProfileID != "" {
		if profile != (domain.Profile{}) || request.Query != "" || request.HistoryProfileID != "" {
			return Result{}, invalid("profile", "Профиль и demo_profile_id нельзя передавать вместе.")
		}
		found := false
		for _, demo := range domain.DemoProfiles() {
			if demo.ID == request.DemoProfileID {
				profile = demo.Profile
				found = true
				break
			}
		}
		if !found {
			return Result{}, invalid("demo_profile_id", "Неизвестный демо-профиль.")
		}
	}
	if request.HistoryProfileID != "" {
		if !slices.ContainsFunc(s.historyProfiles, func(profile domain.HistoryProfile) bool { return profile.ID == request.HistoryProfileID }) {
			return Result{}, invalid("history_profile_id", "Неизвестный сценарий истории.")
		}
	}
	if err := profile.Validate(); err != nil {
		return Result{}, err
	}
	limit := request.Limit
	if limit == 0 {
		limit = 3
	}
	if limit < 1 || limit > 3 {
		return Result{}, invalid("limit", "Ожидается целое число от 1 до 3.")
	}
	result := Result{
		Items:            []Item{},
		Mode:             "fallback",
		Limitations:      []Limitation{},
		AlgorithmVersion: "content-rules-v5",
	}
	if baseline {
		result.AlgorithmVersion = "B0"
	}
	historyID := request.DemoProfileID
	if request.HistoryProfileID != "" {
		historyID = request.HistoryProfileID
	}
	completed, historyTopics, historyAvailable := s.history(historyID)

	if request.DemoProfileID == "P10" && !historyAvailable {
		result.addLimitation("HISTORY_UNAVAILABLE", "Демо-история недоступна для текущего каталога.")
	}
	if request.Query != "" {
		queried, err := s.recommendQuery(ctx, request.Query, profile, completed, historyTopics, limit)
		if err != nil {
			return Result{}, err
		}
		for _, limitation := range result.Limitations {
			queried.addLimitation(limitation.Code, limitation.Message)
		}
		return queried, nil
	}
	if profile.TaskID == "" {
		items, contextUsed, err := s.general(ctx, profile, completed, historyTopics, baseline, limit)
		if err != nil {
			return Result{}, err
		}
		result.Items = items
		message := "Для заполненных полей нет подтверждённых совпадений. Показана общая подборка; выберите задачу."
		if profile == (domain.Profile{}) && contextUsed {
			result.Mode = "personalized"
			message = "Учтены темы из вашей истории. Укажите текущую задачу, чтобы уточнить подборку."
		} else if profile == (domain.Profile{}) {
			result.Mode = "starter"
			result.AlgorithmVersion = "starter-v2"
			message = "Это общая стартовая подборка. Выберите задачу, чтобы уточнить рекомендации."
		} else if contextUsed {
			message = "Задача не указана. Учтены доступные совпадения профиля; выберите задачу, чтобы уточнить рекомендации."
		} else if baseline {
			message = "Задача не указана. Показана общая подборка; выберите задачу."
		}
		result.addLimitation("PROFILE_INCOMPLETE", message)
	} else {
		ranked := []candidate{}
		for i, material := range s.materials {
			if err := ctx.Err(); err != nil {
				return Result{}, err
			}
			if material.Visibility != domain.VisibilityListed || completed[material.Key] {
				continue
			}
			t := taskFit(material, profile.TaskID)
			match := signals{}
			if !baseline {
				match = matchSignals(material, profile, historyTopics)
			}
			lexical := s.textScores[profile.TaskID][material.Key]
			if baseline {
				lexical = 0
			}
			if t == 0 && lexical < minimumTextScore {
				continue
			}
			relevance := relevanceScore(t, lexical)
			context := float64(2*match.m + match.s + match.e + match.i + match.h)
			c := candidate{
				index:   i,
				t:       t,
				score:   relevance * (1 + 0.1*context),
				signals: match, textMatch: lexical >= minimumTextScore,
			}
			if baseline {
				c.score = 0
			}

			ranked = append(ranked, c)
		}
		s.sortCandidates(ranked, baseline)

		direct, contextUsed, contentUsed := false, false, false
		seenCards := map[string]bool{}
		for _, candidate := range ranked {
			material := s.materials[candidate.index]
			if material.URL != nil {
				card := *material.URL + "\n" + strings.ToLower(analysisText(material.Title))
				if seenCards[card] {
					continue
				}
				seenCards[card] = true
			}
			reasons := reasonsFor(material, profile, candidate)
			contentUsed = contentUsed || candidate.textMatch
			result.Items = append(result.Items, item(material, reasons))
			if candidate.t == 2 {
				direct = true
			}
			if candidate.signals != (signals{}) {
				contextUsed = true
			}
			if candidate.t == 1 {
				result.addLimitation("BROADER_TOPIC", "Часть материалов относится к более общей теме вашей задачи.")
			}
			if len(result.Items) == limit {
				break
			}
		}
		if direct || contentUsed {
			result.Mode = "topic"
			if contextUsed {
				result.Mode = "personalized"
			}
		}
	}
	if len(result.Items) == 0 {
		result.Mode = "fallback"
		result.Limitations = slices.DeleteFunc(result.Limitations, func(l Limitation) bool {
			return l.Code == "PROFILE_INCOMPLETE"
		})
		result.addLimitation("NO_MATCHES", "Подходящих материалов не нашлось. Попробуйте изменить запрос.")
	}
	for _, selected := range result.Items {
		if selected.URL == nil {
			result.addLimitation("URL_UNVERIFIED", "Для части материалов нет подтверждённой ссылки. Можно изменить подборку.")
		}
	}
	return result, nil
}

func (s *Service) history(demo string) (map[domain.MaterialKey]bool, map[string]bool, bool) {
	completed := map[domain.MaterialKey]bool{}
	topics := map[string]bool{}
	history, ok := s.histories[demo]
	if !ok {
		return completed, topics, false
	}
	keys := append(slices.Clone(history.Completed), history.Positive...)
	if len(keys) == 0 && len(history.Dismissed) == 0 {
		return completed, topics, false
	}
	for _, key := range append(slices.Clone(keys), history.Dismissed...) {
		index, exists := s.byKey[key]
		if !exists || s.materials[index].Visibility != domain.VisibilityListed {
			return map[domain.MaterialKey]bool{}, map[string]bool{}, false
		}
	}
	for _, key := range append(slices.Clone(history.Completed), history.Dismissed...) {
		completed[key] = true
	}
	for _, key := range keys {
		for _, topic := range s.materials[s.byKey[key]].Topics {
			topics[topic.ID.String()] = true
		}
	}
	return completed, topics, true
}

func taskFit(material domain.Material, task domain.TaskID) int {
	for _, fit := range material.Features.TaskFits {
		if fit.TaskID == task {
			if fit.Match == domain.TaskMatchDirect {
				return 2
			}
			if fit.Match == domain.TaskMatchBroader {
				return 1
			}
		}
	}
	return 0
}

func matchSignals(material domain.Material, profile domain.Profile, historyTopics map[string]bool) signals {
	matched := signals{}
	if profile.HelpMode != "" && slices.Contains(material.Features.HelpModes, profile.HelpMode) {
		matched.m = 1
	}
	if profile.Stage != "" && slices.Contains(material.Features.BusinessStages, profile.Stage) {
		matched.s = 1
	}
	if profile.Experience != "" && slices.Contains(material.Features.ExperienceLevels, profile.Experience) {
		matched.e = 1
	}
	if profile.Industry != "" && slices.Contains(material.Features.Industries, profile.Industry) {
		matched.i = 1
	}
	for _, topic := range material.Topics {
		if historyTopics[topic.ID.String()] {
			matched.h = 1
			break
		}
	}
	return matched
}

func reasonsFor(material domain.Material, profile domain.Profile, candidate candidate) []Reason {
	reasons := []Reason{}
	switch candidate.t {
	case 2:
		reasons = append(reasons, Reason{"task_match", "Соответствует выбранной задаче."})
	case 1:
		reasons = append(reasons, Reason{"broader_topic", "Материал по более общей теме вашей задачи."})
	default:
		if !candidate.textMatch {
			reasons = append(reasons, Reason{"general_selection", "Общая подборка по доступным признакам."})
		}
	}
	if candidate.textMatch {
		reasons = append(reasons, Reason{"content_match", "Содержание связано с выбранной задачей."})
	}
	matched := candidate.signals
	if matched.m > 0 {
		text := map[domain.HelpMode]string{
			domain.HelpModeQuickAnswer:  "Формат — быстрый ответ.",
			domain.HelpModeTool:         "Формат — практический инструмент.",
			domain.HelpModeDeepLearning: "Формат — подробное обучение."}[profile.HelpMode]
		if profile.HelpMode == domain.HelpModeQuickAnswer && material.Key.Kind == domain.MaterialKindArticle &&
			material.Duration.Minutes != nil {
			text = fmt.Sprintf("Короткий материал: %d мин.", *material.Duration.Minutes)
		}
		reasons = append(reasons, Reason{"help_mode_match", text})
	}
	if matched.s > 0 {
		reasons = append(reasons, Reason{"stage_match", "Подходит для выбранной стадии бизнеса."})
	}
	if matched.e > 0 {
		reasons = append(reasons, Reason{"experience_match", "Подходит для выбранного уровня опыта."})
	}
	if matched.i > 0 {
		reasons = append(reasons, Reason{"industry_match", "Соответствует выбранной сфере бизнеса."})
	}
	if matched.h > 0 {
		reasons = append(reasons, Reason{"history_match", "Связано с темами материалов из истории обучения и положительных действий."})
	}
	return reasons
}

func item(material domain.Material, reasons []Reason) Item {
	return Item{
		Key:     material.Key,
		Title:   material.Title,
		Summary: material.Summary,
		Access:  material.Access,
		URL:     material.URL,
		Reasons: reasons,
	}
}
func (r *Result) addLimitation(code, message string) {
	if !slices.ContainsFunc(r.Limitations, func(l Limitation) bool { return l.Code == code }) {
		r.Limitations = append(r.Limitations, Limitation{code, message})
	}
}
func invalid(field, message string) error {
	return &domain.InputError{Fields: []domain.FieldError{{Field: field, Message: message}}}
}
func compareKeys(a, b domain.MaterialKey) int {
	if a.String() < b.String() {
		return -1
	}
	if a.String() > b.String() {
		return 1
	}
	return 0
}

func (s *Service) sortCandidates(ranked []candidate, baseline bool) {
	slices.SortFunc(ranked, func(a, b candidate) int {
		if baseline && a.t != b.t {
			return b.t - a.t
		}
		if !baseline && (a.t == 2) != (b.t == 2) {
			if a.t == 2 {
				return -1
			}
			return 1
		}
		if a.score > b.score {
			return -1
		}
		if a.score < b.score {
			return 1
		}

		return compareKeys(s.materials[a.index].Key, s.materials[b.index].Key)
	})
}

func relevanceScore(taskMatch int, textScore float64) float64 {
	switch taskMatch {
	case 2:
		return 0.75 + 0.25*textScore
	case 1:
		return 0.25 + 0.5*textScore
	default:
		return textScore
	}
}

func (s *Service) HistoryProfiles() []domain.HistoryProfile { return slices.Clone(s.historyProfiles) }
