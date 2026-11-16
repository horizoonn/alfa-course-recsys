package textsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/service/recommendation"
)

type Client struct {
	url, mode, catalogID string
	http                 *http.Client
}

func New(address, mode, catalogID string, timeout time.Duration) (*Client, error) {
	parsed, err := url.Parse(address)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.User != nil {
		return nil, fmt.Errorf("TEXT_SEARCH_URL must be an absolute HTTP URL")
	}
	if mode != "tfidf" && mode != "e5" && mode != "rrf" {
		return nil, fmt.Errorf("TEXT_SEARCH_MODE must be tfidf, e5 or rrf")
	}
	if timeout <= 0 {
		return nil, fmt.Errorf("TEXT_SEARCH_TIMEOUT must be positive")
	}
	return &Client{url: strings.TrimRight(address, "/") + "/search", mode: mode, catalogID: catalogID, http: &http.Client{Timeout: timeout}}, nil
}

func (c *Client) Search(ctx context.Context, query string) (recommendation.SearchResult, error) {
	body, err := json.Marshal(struct {
		Query     string `json:"query"`
		CatalogID string `json:"catalog_id"`
		Mode      string `json:"mode"`
	}{query, c.catalogID, c.mode})
	if err != nil {
		return recommendation.SearchResult{}, recommendation.ErrSearchUnavailable
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return recommendation.SearchResult{}, recommendation.ErrSearchUnavailable
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := c.http.Do(request)
	if err != nil {
		return recommendation.SearchResult{}, recommendation.ErrSearchUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return recommendation.SearchResult{}, recommendation.ErrSearchUnavailable
	}
	var result recommendation.SearchResult
	decoder := json.NewDecoder(io.LimitReader(response.Body, 4<<20))
	if err := decoder.Decode(&result); err != nil {
		return result, recommendation.ErrSearchUnavailable
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF {
		return result, recommendation.ErrSearchUnavailable
	}
	if result.CatalogID != c.catalogID || result.IndexID == "" || (result.State != "ok" && result.State != "unclear") || (result.Mode != "tfidf" && result.Mode != "e5" && result.Mode != "rrf") || result.Algorithm != "query-search-v3/"+result.Mode {
		return result, recommendation.ErrSearchUnavailable
	}
	if result.State == "unclear" && (len(result.Matches) > 0 || result.Message == "") {
		return result, recommendation.ErrSearchUnavailable
	}
	seen := map[string]bool{}
	for i, match := range result.Matches {
		if _, err := domain.ParseMaterialKey(match.Key); err != nil || seen[match.Key] || match.Rank != i+1 || !validScore(match.Score) || !validScore(match.Word) || !validScore(match.Character) || !validScore(match.Lexical) || (match.Semantic != nil && !validScore(*match.Semantic)) {
			return result, recommendation.ErrSearchUnavailable
		}
		seen[match.Key] = true
	}
	for _, limitation := range result.Limitations {
		if limitation.Code != "SEMANTIC_UNAVAILABLE" || limitation.Message == "" {
			return result, recommendation.ErrSearchUnavailable
		}
	}
	for _, kind := range append(result.ExcludeKinds, result.AllowedKinds...) {
		if kind != domain.MaterialKindArticle && kind != domain.MaterialKindCourse && kind != domain.MaterialKindDocument && kind != domain.MaterialKindCalculator && kind != domain.MaterialKindTerm && kind != domain.MaterialKindPodcast && kind != domain.MaterialKindWebinar {
			return result, recommendation.ErrSearchUnavailable
		}
	}
	return result, nil
}
func validScore(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0 && value <= 1
}

func (c *Client) Related(ctx context.Context, key domain.MaterialKey, query string) (recommendation.RelatedResult, error) {
	var result recommendation.RelatedResult
	body, err := json.Marshal(struct {
		MaterialKey string `json:"material_key"`
		Query       string `json:"query"`
		CatalogID   string `json:"catalog_id"`
	}{key.String(), query, c.catalogID})
	if err != nil {
		return result, recommendation.ErrSearchUnavailable
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimSuffix(c.url, "/search")+"/related", bytes.NewReader(body))
	if err != nil {
		return result, recommendation.ErrSearchUnavailable
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := c.http.Do(req)
	if err != nil {
		return result, recommendation.ErrSearchUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return result, recommendation.ErrSearchUnavailable
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, 1<<20))
	if err := decoder.Decode(&result); err != nil {
		return result, recommendation.ErrSearchUnavailable
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF || result.CatalogID != c.catalogID || result.SourceKey != key.String() || result.Algorithm != "related-content-v1" || len(result.Matches) > 40 || len(result.Anchors) > 5 {
		return result, recommendation.ErrSearchUnavailable
	}
	seen := map[string]bool{}
	for _, match := range result.Matches {
		if _, err := domain.ParseMaterialKey(match.Key); err != nil || seen[match.Key] || !validScore(match.Score) {
			return result, recommendation.ErrSearchUnavailable
		}
		seen[match.Key] = true
	}
	return result, nil
}
