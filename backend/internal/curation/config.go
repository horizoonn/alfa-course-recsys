package curation

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/service/catalog"
)

type Config struct {
	Reviewed    []catalog.ReviewedRule
	Starter     []domain.MaterialKey
	DemoHistory map[string]domain.History
}

type fileConfig struct {
	Version  string `json:"version"`
	Reviewed []struct {
		Key       string `json:"key"`
		TitleHash string `json:"title_hash"`
		TaskFits  []struct {
			TaskID domain.TaskID    `json:"task_id"`
			Match  domain.TaskMatch `json:"match"`
		} `json:"task_fits"`
	} `json:"reviewed"`
	Starter     []string `json:"starter"`
	DemoHistory map[string]struct {
		Completed []string `json:"completed"`
		Positive  []string `json:"positive"`
	} `json:"demo_history"`
}

func Load(path string) (Config, error) {
	cfg := Config{
		Reviewed:    []catalog.ReviewedRule{},
		Starter:     []domain.MaterialKey{},
		DemoHistory: map[string]domain.History{},
	}
	if path == "" {
		return cfg, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open catalog curation file: %w", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return Config{}, err
	}
	if info.Size() > 1<<20 {
		return Config{}, errors.New("catalog curation file is too large")
	}
	decoder := json.NewDecoder(io.LimitReader(file, 1<<20))
	decoder.DisallowUnknownFields()
	var data fileConfig
	if err := decoder.Decode(&data); err != nil {
		return Config{}, errors.New("invalid catalog curation JSON")
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return Config{}, errors.New("unexpected data after catalog curation JSON")
	}
	if data.Version != "v1" {
		return Config{}, errors.New("unsupported catalog curation version")
	}
	seen := map[domain.MaterialKey]bool{}
	for _, row := range data.Reviewed {
		key, err := domain.ParseMaterialKey(row.Key)
		if err != nil {
			return Config{}, err
		}
		if seen[key] {
			return Config{}, errors.New("duplicate reviewed material")
		}
		seen[key] = true
		hash, err := hex.DecodeString(row.TitleHash)
		if err != nil || len(hash) != 32 {
			return Config{}, errors.New("invalid reviewed title checksum")
		}
		if len(row.TaskFits) == 0 {
			return Config{}, errors.New("reviewed material requires task fits")
		}
		fits := []domain.TaskFit{}
		for _, fit := range row.TaskFits {
			if err := (domain.Profile{TaskID: fit.TaskID}).Validate(); err != nil {
				return Config{}, err
			}
			fits = append(fits, domain.TaskFit{
				TaskID: fit.TaskID,
				Match:  fit.Match,
			})
		}
		if err := (domain.MaterialFeatures{TaskFits: fits}).Validate(); err != nil {
			return Config{}, err
		}
		cfg.Reviewed = append(cfg.Reviewed, catalog.ReviewedRule{
			Key:       key,
			TitleHash: row.TitleHash,
			Fits:      fits,
		})
	}
	cfg.Starter, err = parseKeys(data.Starter)
	if err != nil {
		return Config{}, err
	}
	for _, key := range cfg.Starter {
		if !seen[key] {
			return Config{}, errors.New("starter material requires a reviewed rule")
		}
	}
	for id, history := range data.DemoHistory {
		if id != "P10" {
			return Config{}, errors.New("history is only supported for demo P10")
		}
		completed, err := parseKeys(history.Completed)
		if err != nil {
			return Config{}, err
		}
		positive, err := parseKeys(history.Positive)
		if err != nil {
			return Config{}, err
		}
		for _, key := range append(completed, positive...) {
			if !seen[key] {
				return Config{}, errors.New("history material requires a reviewed rule")
			}
		}
		cfg.DemoHistory[id] = domain.History{
			Completed: completed,
			Positive:  positive,
		}
	}
	return cfg, nil
}

func parseKeys(values []string) ([]domain.MaterialKey, error) {
	result := []domain.MaterialKey{}
	seen := map[domain.MaterialKey]bool{}
	for _, value := range values {
		key, err := domain.ParseMaterialKey(value)
		if err != nil {
			return nil, err
		}
		if seen[key] {
			return nil, errors.New("duplicate material key in curation list")
		}
		seen[key] = true
		result = append(result, key)
	}
	return result, nil
}

func (c Config) Active(
	evidence map[domain.MaterialKey][]domain.FeatureEvidence,
) ([]domain.MaterialKey, map[string]domain.History) {
	trusted := func(key domain.MaterialKey) bool {
		for _, e := range evidence[key] {
			if e.Method == "manual" {
				return true
			}
		}
		return false
	}
	starter := []domain.MaterialKey{}
	for _, key := range c.Starter {
		if trusted(key) {
			starter = append(starter, key)
		}
	}
	histories := map[string]domain.History{}
	for id, history := range c.DemoHistory {
		valid := len(history.Completed)+len(history.Positive) > 0
		for _, key := range append(append([]domain.MaterialKey{}, history.Completed...), history.Positive...) {
			if !trusted(key) {
				valid = false
			}
		}
		if valid {
			histories[id] = history
		}
	}
	return starter, histories
}
