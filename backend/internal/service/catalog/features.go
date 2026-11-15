package catalog

import (
	"cmp"
	"crypto/sha256"
	"fmt"
	"slices"
	"strconv"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
)

const featureVersion = "v1"

type featureBuilder struct {
	features domain.MaterialFeatures
	evidence []domain.FeatureEvidence
}

func PrepareFeatures(
	materials []domain.Material,
	reviewed []ReviewedRule,
) ([]domain.Material, map[domain.MaterialKey][]domain.FeatureEvidence, error) {
	return prepareFeatures(materials, labelRules, reviewed)
}

func prepareFeatures(
	materials []domain.Material,
	labels []labelRule,
	reviewed []ReviewedRule,
) ([]domain.Material, map[domain.MaterialKey][]domain.FeatureEvidence, error) {
	result := make([]domain.Material, len(materials))
	evidence := make(map[domain.MaterialKey][]domain.FeatureEvidence)
	seen := make(map[domain.MaterialKey]struct{}, len(materials))
	for i, material := range materials {
		if _, exists := seen[material.Key]; exists {
			return nil, nil, fmt.Errorf("prepare features: duplicate material key %s", material.Key)
		}
		seen[material.Key] = struct{}{}
		builder := featureBuilder{
			features: domain.MaterialFeatures{
				TaskFits:         []domain.TaskFit{},
				BusinessStages:   []domain.BusinessStage{},
				Industries:       []domain.Industry{},
				ExperienceLevels: []domain.ExperienceLevel{},
				HelpModes:        []domain.HelpMode{},
			},
			evidence: []domain.FeatureEvidence{},
		}
		for _, rule := range labels {
			for _, label := range sourceLabels(material, rule.source) {
				if label.ID != rule.id || (rule.kind != "" && rule.kind != material.Key.Kind) {
					continue
				}
				source := rule.source + "/" + label.ID.String()
				ruleID := rule.source + "/" + rule.value
				builder.add(rule.feature, rule.value, domain.TaskMatchBroader, source, ruleID, "rule")
			}
		}
		addFormat(&builder, material)
		for _, rule := range reviewed {
			if material.Key != rule.Key || fmt.Sprintf("%x", sha256.Sum256([]byte(material.Title))) != rule.TitleHash {
				continue
			}
			for _, fit := range rule.Fits {
				builder.add(
					domain.FeatureTaskFit,
					string(fit.TaskID),
					fit.Match,
					"Title",
					"reviewed/"+material.Key.String()+"/"+string(fit.TaskID),
					"manual",
				)
			}
		}
		builder.sort()
		if err := builder.features.Validate(); err != nil {
			return nil, nil, fmt.Errorf("prepare features for %s: %w", material.Key, err)
		}
		result[i] = material
		result[i].Features = builder.features
		if len(builder.evidence) > 0 {
			evidence[material.Key] = builder.evidence
		}
	}
	return result, evidence, nil
}

func sourceLabels(material domain.Material, source string) []domain.Label {
	switch source {
	case "Topics":
		return material.Topics
	case "Tags":
		return material.Tags
	case "Metadata.Category":
		if material.Metadata.Category != nil {
			return []domain.Label{*material.Metadata.Category}
		}
	case "Metadata.AdditionalLabels":
		return material.Metadata.AdditionalLabels
	}
	return nil
}

func addFormat(builder *featureBuilder, material domain.Material) {
	mode := domain.HelpMode("")
	source := "Kind"
	rule := "format/" + string(material.Key.Kind)
	switch material.Key.Kind {
	case domain.MaterialKindDocument, domain.MaterialKindCalculator:
		mode = domain.HelpModeTool
	case domain.MaterialKindCourse:
		mode = domain.HelpModeDeepLearning
	case domain.MaterialKindTerm:
		mode = domain.HelpModeQuickAnswer
	case domain.MaterialKindArticle:
		if minutes := material.Duration.Minutes; minutes != nil && *minutes > 0 && *minutes <= 10 {
			mode = domain.HelpModeQuickAnswer
			source = "Duration.Minutes/" + strconv.Itoa(*minutes)
			rule = "format/article-up-to-10-minutes"
		}
	}
	if mode != "" {
		builder.add(domain.FeatureHelpMode, string(mode), "", source, rule, "rule")
	}
}

func (b *featureBuilder) add(
	feature domain.FeatureName,
	value string,
	match domain.TaskMatch,
	source, rule, method string,
) {
	switch feature {
	case domain.FeatureTaskFit:
		index := slices.IndexFunc(b.features.TaskFits, func(fit domain.TaskFit) bool { return string(fit.TaskID) == value })
		if index < 0 {
			b.features.TaskFits = append(b.features.TaskFits, domain.TaskFit{
				TaskID: domain.TaskID(value),
				Match:  match,
			})
		} else {
			previous := b.features.TaskFits[index].Match
			if previous == domain.TaskMatchDirect && match == domain.TaskMatchBroader {
				return
			}
			if previous == domain.TaskMatchBroader && match == domain.TaskMatchDirect {
				b.features.TaskFits[index].Match = match
				b.evidence = slices.DeleteFunc(b.evidence, func(e domain.FeatureEvidence) bool { return e.Feature == feature && e.Value == value })
			}
		}
	case domain.FeatureIndustry:
		match = ""
		if !slices.Contains(b.features.Industries, domain.Industry(value)) {
			b.features.Industries = append(b.features.Industries, domain.Industry(value))
		}
	case domain.FeatureHelpMode:
		match = ""
		if !slices.Contains(b.features.HelpModes, domain.HelpMode(value)) {
			b.features.HelpModes = append(b.features.HelpModes, domain.HelpMode(value))
		}
	}
	evidence := domain.FeatureEvidence{
		Feature: feature,
		Value:   value,
		Match:   match,
		Source:  source,
		Rule:    rule,
		Method:  method,
		Version: featureVersion,
	}
	if !slices.Contains(b.evidence, evidence) {
		b.evidence = append(b.evidence, evidence)
	}
}

func (b *featureBuilder) sort() {
	slices.SortFunc(b.features.TaskFits, func(a, b domain.TaskFit) int { return cmp.Compare(a.TaskID, b.TaskID) })
	slices.Sort(b.features.Industries)
	slices.Sort(b.features.HelpModes)
	slices.SortFunc(b.evidence, func(a, b domain.FeatureEvidence) int {
		if c := cmp.Compare(a.Feature, b.Feature); c != 0 {
			return c
		}
		if c := cmp.Compare(a.Value, b.Value); c != 0 {
			return c
		}
		if c := cmp.Compare(a.Source, b.Source); c != 0 {
			return c
		}
		return cmp.Compare(a.Rule, b.Rule)
	})
}
