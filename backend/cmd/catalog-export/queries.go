package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/client/textsearch"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/config"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/service/recommendation"
)

type queryScenario struct {
	ID      string       `json:"id"`
	Query   string       `json:"query"`
	Profile queryProfile `json:"profile"`
}

type queryProfile struct {
	TaskID     domain.TaskID          `json:"task_id,omitempty"`
	Stage      domain.BusinessStage   `json:"stage,omitempty"`
	Industry   domain.Industry        `json:"industry,omitempty"`
	Experience domain.ExperienceLevel `json:"experience,omitempty"`
	HelpMode   domain.HelpMode        `json:"help_mode,omitempty"`
}

func (p queryProfile) domain() domain.Profile {
	return domain.Profile{TaskID: p.TaskID, Stage: p.Stage, Industry: p.Industry, Experience: p.Experience, HelpMode: p.HelpMode}
}

type queryVariant struct {
	Mode            string                      `json:"mode"`
	Search          recommendation.SearchResult `json:"search"`
	Recommendations recommendation.Result       `json:"recommendations"`
	Top             []queryCard                 `json:"top"`
	ElapsedMS       float64                     `json:"elapsed_ms"`
}

type queryCard struct {
	Key   string  `json:"key"`
	Title string  `json:"title"`
	Score float64 `json:"score"`
}
type queryComparison struct {
	Scenario queryScenario  `json:"scenario"`
	Variants []queryVariant `json:"variants"`
}

type capturedSearch struct {
	client *textsearch.Client
	result recommendation.SearchResult
}

func (s *capturedSearch) Search(ctx context.Context, query string) (recommendation.SearchResult, error) {
	result, err := s.client.Search(ctx, query)
	s.result = result
	return result, err
}

func compareQueries(cfg config.Config, materials []domain.Material, corpus recommendation.SearchCorpus, input, output string) error {
	body, err := os.ReadFile(input)
	if err != nil {
		return err
	}
	var scenarios []queryScenario
	if err = json.Unmarshal(body, &scenarios); err != nil {
		return err
	}
	titles := map[string]string{}
	for _, material := range materials {
		titles[material.Key.String()] = material.Title
	}
	comparisons := []queryComparison{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	for _, scenario := range scenarios {
		comparison := queryComparison{Scenario: scenario, Variants: []queryVariant{}}
		for _, mode := range []string{"tfidf", "e5", "rrf"} {
			client, err := textsearch.New(cfg.SearchURL, mode, corpus.ID, cfg.SearchTimeout)
			if err != nil {
				return err
			}
			search := &capturedSearch{client: client}
			service, err := recommendation.NewService(materials, nil, nil, nil, search, nil)
			if err != nil {
				return err
			}
			started := time.Now()
			result, err := service.Recommend(ctx, recommendation.Request{Query: scenario.Query, Profile: scenario.Profile.domain()})
			if err != nil {
				return fmt.Errorf("%s/%s: %w", scenario.ID, mode, err)
			}
			variant := queryVariant{Mode: mode, Search: search.result, Recommendations: result, ElapsedMS: float64(time.Since(started).Microseconds()) / 1000, Top: []queryCard{}}
			for _, match := range search.result.Matches {
				variant.Top = append(variant.Top, queryCard{match.Key, titles[match.Key], match.Score})
				if len(variant.Top) == 10 {
					break
				}
			}
			comparison.Variants = append(comparison.Variants, variant)
		}
		comparisons = append(comparisons, comparison)
	}
	body, err = json.MarshalIndent(struct {
		CatalogID string            `json:"catalog_id"`
		Cases     []queryComparison `json:"cases"`
	}{corpus.ID, comparisons}, "", "  ")
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(output), 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(output), ".query-comparison-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err = file.Write(body); err != nil {
		file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	if err = os.Rename(file.Name(), output); err != nil {
		return err
	}
	fmt.Printf("Compared %d queries in three modes; saved %s\n", len(scenarios), output)
	return nil
}
