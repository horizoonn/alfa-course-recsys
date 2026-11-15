package recommendation

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"slices"
	"strings"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
)

const minimumTextScore = 0.1
const textAlgorithm = "tfidf-snowball-v1"

type TextScores map[domain.TaskID]map[domain.MaterialKey]float64

type SearchCorpus struct {
	ID        string                   `json:"catalog_id"`
	Documents []Document               `json:"documents"`
	Queries   map[domain.TaskID]string `json:"queries"`
}

type textArtifact struct {
	SchemaVersion int                             `json:"schema_version"`
	Algorithm     string                          `json:"algorithm"`
	CatalogID     string                          `json:"catalog_id"`
	Keys          []string                        `json:"keys"`
	Scores        map[domain.TaskID][]json.Number `json:"scores"`
	Producer      map[string]string               `json:"producer"`
}

func NewSearchCorpus(materials []domain.Material) SearchCorpus {
	corpus := SearchCorpus{Documents: Documents(materials), Queries: map[domain.TaskID]string{}}
	slices.SortFunc(corpus.Documents, func(a, b Document) int { return strings.Compare(a.Key, b.Key) })
	for i := range corpus.Documents {
		slices.SortFunc(corpus.Documents[i].Fields, func(a, b TextField) int {
			if n := strings.Compare(a.Name, b.Name); n != 0 {
				return n
			}
			return strings.Compare(a.Text, b.Text)
		})
	}
	for _, task := range domain.ProfileMetadata().Tasks {
		id := domain.TaskID(task.ID)
		corpus.Queries[id] = TaskQuery(id)
	}
	body, _ := json.Marshal(corpus)
	corpus.ID = fmt.Sprintf("%x", sha256.Sum256(body))
	return corpus
}

func LoadTextScores(path string, corpus SearchCorpus) (TextScores, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open text scores (run text preparation): %w", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(io.LimitReader(file, 64<<20))
	decoder.DisallowUnknownFields()
	var artifact textArtifact
	if err := decoder.Decode(&artifact); err != nil {
		return nil, fmt.Errorf("decode text scores: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, errors.New("trailing data in text scores")
	}
	if artifact.SchemaVersion != 1 || artifact.Algorithm != textAlgorithm {
		return nil, errors.New("unsupported text scores version; run text preparation")
	}
	if artifact.CatalogID != corpus.ID {
		return nil, errors.New("text scores do not match catalog or tasks; run text preparation")
	}
	if len(artifact.Keys) != len(corpus.Documents) || len(artifact.Scores) != len(corpus.Queries) {
		return nil, errors.New("incomplete text scores")
	}
	expected := make(map[string]bool, len(corpus.Documents))
	for _, doc := range corpus.Documents {
		expected[doc.Key] = true
	}
	keys := make([]domain.MaterialKey, len(artifact.Keys))
	for i, raw := range artifact.Keys {
		key, err := domain.ParseMaterialKey(raw)
		if err != nil || !expected[raw] {
			return nil, fmt.Errorf("invalid or duplicate text score key %q", raw)
		}
		delete(expected, raw)
		keys[i] = key
	}
	scores := make(TextScores, len(corpus.Queries))
	for task := range corpus.Queries {
		values, ok := artifact.Scores[task]
		if !ok || len(values) != len(keys) {
			return nil, fmt.Errorf("incomplete text scores for task %s", task)
		}
		scores[task] = make(map[domain.MaterialKey]float64, len(keys))
		for i, number := range values {
			value, err := number.Float64()
			if err != nil || math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 1 {
				return nil, fmt.Errorf("invalid text score for task %s, key %s", task, keys[i])
			}
			scores[task][keys[i]] = value
		}
	}
	return scores, nil
}
