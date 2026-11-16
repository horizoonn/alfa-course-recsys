package recommendation

import (
	"context"
	"strings"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
)

func (s *Service) general(
	ctx context.Context,
	profile domain.Profile,
	completed map[domain.MaterialKey]bool,
	historyTopics map[string]bool,
	baseline bool,
	limit int,
) ([]Item, bool, error) {
	items := []Item{}
	keys := map[domain.MaterialKey]bool{}
	cards := map[string]bool{}
	add := func(index int, reasons []Reason) bool {
		material := s.materials[index]
		if len(items) == limit || material.Visibility != domain.VisibilityListed ||
			completed[material.Key] || keys[material.Key] || material.URL == nil || strings.TrimSpace(*material.URL) == "" {
			return false
		}
		card := *material.URL + "\n" + strings.ToLower(analysisText(material.Title))
		if cards[card] {
			return false
		}
		keys[material.Key] = true
		cards[card] = true
		items = append(items, item(material, reasons))
		return true
	}
	contextUsed := false
	if !baseline && (profile != (domain.Profile{}) || len(historyTopics) > 0) {
		ranked := []candidate{}
		for index, material := range s.materials {
			if err := ctx.Err(); err != nil {
				return nil, false, err
			}
			matched := matchSignals(material, profile, historyTopics)
			if matched == (signals{}) {
				continue
			}
			textScore := 0.0
			for _, scores := range s.textScores {
				textScore = max(textScore, scores[material.Key])
			}
			ranked = append(ranked, candidate{
				index:   index,
				score:   float64(2*matched.m+matched.s+matched.e+matched.i+matched.h) + 0.1*textScore,
				signals: matched,
			})
		}
		s.sortCandidates(ranked, false)
		for _, candidate := range ranked {
			if add(candidate.index, reasonsFor(s.materials[candidate.index], profile, candidate)) {
				contextUsed = true
			}
			if len(items) == limit {
				return items, contextUsed, nil
			}
		}
	}
	for _, key := range s.starter {
		if err := ctx.Err(); err != nil {
			return nil, false, err
		}
		if index, exists := s.byKey[key]; exists {
			add(index, []Reason{{"general_selection", "Материал из проверенной стартовой подборки."}})
		}
		if len(items) == limit {
			return items, contextUsed, nil
		}
	}
	generalReasons := []Reason{{"general_selection", "Материал из общей подборки по запуску и ведению бизнеса."}}
	remaining := []candidate{}
	for _, task := range []domain.TaskID{"first_steps", "finance", "attract_clients"} {
		ranked := []candidate{}
		for index, material := range s.materials {
			if err := ctx.Err(); err != nil {
				return nil, false, err
			}
			if material.Visibility != domain.VisibilityListed || material.Access != domain.AccessModelFree ||
				material.URL == nil || completed[material.Key] || keys[material.Key] {
				continue
			}
			fit := taskFit(material, task)
			textScore := s.textScores[task][material.Key]
			if fit == 0 && textScore < minimumTextScore {
				continue
			}
			ranked = append(ranked, candidate{index: index, t: fit, score: relevanceScore(fit, textScore)})
		}
		s.sortCandidates(ranked, false)
		for _, candidate := range ranked {
			if add(candidate.index, generalReasons) {
				break
			}
		}
		if len(items) == limit {
			return items, contextUsed, nil
		}
		remaining = append(remaining, ranked...)
	}
	s.sortCandidates(remaining, false)
	for _, candidate := range remaining {
		if err := ctx.Err(); err != nil {
			return nil, false, err
		}
		add(candidate.index, generalReasons)
		if len(items) == limit {
			break
		}
	}
	return items, contextUsed, nil
}
