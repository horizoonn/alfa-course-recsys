package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/config"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/curation"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
	repository "github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/catalog"
	catalog "github.com/horizoonn/alfa-course-recsys/backend/internal/service/catalog"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/service/recommendation"
	"github.com/horizoonn/alfa-course-recsys/backend/platform/postgres"
)

type demoCase struct {
	ID              string   `json:"id"`
	Query           string   `json:"query"`
	Candidates      []string `json:"candidates"`
	Baseline        []string `json:"baseline"`
	Recommendations []string `json:"recommendations"`
	Cold            bool     `json:"cold"`
}

func main() {
	output := flag.String("output", "../data/ranking_catalog.json", "Output file")
	queries := flag.String("queries", "", "Query scenarios for comparing TF-IDF, E5 and RRF")
	documentsOnly := flag.Bool("documents-only", false, "Export search input without loading prepared scores")
	flag.Parse()
	if err := run(*output, *documentsOnly, *queries); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(output string, documentsOnly bool, queries string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), cfg.StartupTimeout)
	defer cancel()
	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	tx := postgres.NewTxManager(pool)
	materials, err := catalog.NewService(repository.NewRepository(tx.Executor), tx).Load(ctx)
	pool.Close()
	if err != nil {
		return err
	}
	curated, err := curation.Load(cfg.CurationPath)
	if err != nil {
		return err
	}
	materials, evidence, err := catalog.PrepareFeatures(materials, curated.Reviewed)
	if err != nil {
		return err
	}
	corpus := recommendation.NewSearchCorpus(materials)
	if queries != "" {
		return compareQueries(cfg, materials, corpus, queries, output)
	}
	documents := corpus.Documents
	starter, histories := curated.Active(evidence)
	var service *recommendation.Service
	if !documentsOnly {
		scores, err := recommendation.LoadTextScores(cfg.TextScoresPath, corpus)
		if err != nil {
			return err
		}
		service, err = recommendation.NewService(materials, starter, histories, scores, nil, nil)
		if err != nil {
			return err
		}
	}
	cases := []demoCase{}
	for _, demo := range domain.DemoProfiles() {
		if documentsOnly {
			break
		}
		request := recommendation.Request{DemoProfileID: demo.ID, Limit: 3}
		baseline, err := service.Baseline(ctx, request)
		if err != nil {
			return err
		}
		local, err := service.Recommend(ctx, request)
		if err != nil {
			return err
		}
		entry := demoCase{
			ID:              demo.ID,
			Query:           recommendation.TaskQuery(demo.Profile.TaskID),
			Candidates:      []string{},
			Baseline:        []string{},
			Recommendations: []string{},
			Cold:            demo.Profile == (domain.Profile{}),
		}
		completed := map[string]bool{}
		for _, key := range histories[demo.ID].Completed {
			completed[key.String()] = true
		}
		for _, doc := range documents {
			if !completed[doc.Key] {
				entry.Candidates = append(entry.Candidates, doc.Key)
			}
		}

		for _, item := range baseline.Items {
			entry.Baseline = append(entry.Baseline, item.Key.String())
		}
		for _, item := range local.Items {
			entry.Recommendations = append(entry.Recommendations, item.Key.String())
		}
		cases = append(cases, entry)
	}
	payload := struct {
		Catalog recommendation.SearchCorpus `json:"catalog"`
		Cases   []demoCase                  `json:"cases"`
	}{Catalog: corpus, Cases: cases}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(output), 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(output), ".ranking-export-*")
	if err != nil {
		return err
	}
	path := file.Name()
	defer os.Remove(path)
	if err := file.Chmod(0600); err != nil {
		file.Close()
		return err
	}
	if _, err := file.Write(body); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Rename(path, output); err != nil {
		return err
	}
	fmt.Printf("Exported %d eligible materials and %d profiles\n", len(documents), len(cases))
	return nil
}
