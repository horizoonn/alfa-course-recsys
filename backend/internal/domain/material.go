package domain

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
	"uuid"
)

type MaterialKind string

const (
	MaterialKindArticle    MaterialKind = "article"
	MaterialKindCourse     MaterialKind = "course"
	MaterialKindDocument   MaterialKind = "document"
	MaterialKindTerm       MaterialKind = "term"
	MaterialKindCalculator MaterialKind = "calculator"
	MaterialKindPodcast    MaterialKind = "podcast"
	MaterialKindWebinar    MaterialKind = "webinar"
)

type Visibility string

const (
	VisibilityListed  Visibility = "listed"
	VisibilityHidden  Visibility = "hidden"
	VisibilityUnknown Visibility = "unknown"
)

type AccessModel string

const (
	AccessModelFree         AccessModel = "free"
	AccessModelSubscription AccessModel = "subscription"
	AccessModelPaid         AccessModel = "paid"
	AccessModelUnknown      AccessModel = "unknown"
)

type MaterialKey struct {
	Kind MaterialKind
	ID   uuid.UUID
}

func (k MaterialKey) String() string {
	return string(k.Kind) + ":" + k.ID.String()
}

func (k MaterialKey) Validate() error {
	switch k.Kind {
	case MaterialKindArticle, MaterialKindCourse, MaterialKindDocument, MaterialKindTerm,
		MaterialKindCalculator, MaterialKindPodcast, MaterialKindWebinar:
	default:
		return fmt.Errorf("unsupported material kind %q", k.Kind)
	}
	if k.ID == uuid.Nil() {
		return errors.New("material ID must not be nil UUID")
	}
	return nil
}

type Label struct {
	ID    uuid.UUID
	Title string
}

type MaterialDuration struct {
	Text    string
	Minutes *int
}

func (d MaterialDuration) Validate() error {
	if d.Minutes != nil && *d.Minutes <= 0 {
		return errors.New("duration minutes must be positive")
	}
	return nil
}

type Material struct {
	Key              MaterialKey
	Title            string
	Summary          *string
	SearchText       string
	Topics           []Label
	Tags             []Label
	Metadata         MaterialMetadata
	Features         MaterialFeatures
	Duration         MaterialDuration
	Rating           *RatingSummary
	Visibility       Visibility
	Access           AccessModel
	URL              *string
	PublishedAt      *time.Time
	SourceModifiedAt time.Time
}

func (m Material) Validate() error {
	if err := m.Key.Validate(); err != nil {
		return fmt.Errorf("key: %w", err)
	}
	if strings.TrimSpace(m.Title) == "" {
		return errors.New("title must not be empty")
	}
	if m.Summary != nil && strings.TrimSpace(*m.Summary) == "" {
		return errors.New("summary must be nonempty or nil")
	}
	if strings.TrimSpace(m.SearchText) == "" {
		return errors.New("search text must not be empty")
	}
	if err := validateLabel(m.Topics); err != nil {
		return fmt.Errorf("topics: %w", err)
	}
	if err := validateLabel(m.Tags); err != nil {
		return fmt.Errorf("tags: %w", err)
	}
	if m.Metadata.Course != nil && m.Key.Kind != MaterialKindCourse {
		return errors.New("course metadata is only allowed for course materials")
	}
	if m.Metadata.File != nil && m.Key.Kind != MaterialKindDocument {
		return errors.New("file metadata is only allowed for document materials")
	}
	if m.Metadata.Term != nil && m.Key.Kind != MaterialKindTerm {
		return errors.New("term metadata is only allowed for term materials")
	}
	if m.Metadata.Webinar != nil && m.Key.Kind != MaterialKindWebinar {
		return errors.New("webinar metadata is only allowed for webinar materials")
	}
	if err := m.Metadata.Validate(); err != nil {
		return fmt.Errorf("metadata: %w", err)
	}
	if err := m.Features.Validate(); err != nil {
		return fmt.Errorf("features: %w", err)
	}
	if err := m.Duration.Validate(); err != nil {
		return fmt.Errorf("duration: %w", err)
	}
	if m.Rating != nil {
		if err := m.Rating.Validate(); err != nil {
			return fmt.Errorf("rating: %w", err)
		}
	}
	switch m.Visibility {
	case VisibilityListed, VisibilityHidden, VisibilityUnknown:
	default:
		return fmt.Errorf("unsupported visibility %q", m.Visibility)
	}
	switch m.Access {
	case AccessModelFree, AccessModelSubscription, AccessModelPaid, AccessModelUnknown:
	default:
		return fmt.Errorf("unsupported access model %q", m.Access)
	}
	if m.URL != nil {
		parsed, err := url.Parse(*m.URL)
		if err != nil || parsed.Hostname() == "" ||
			(!strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https")) ||
			strings.ContainsAny(*m.URL, " \t\r\n") {
			return errors.New("verified URL must be an absolute HTTP(S) URL")
		}
	}
	if m.PublishedAt != nil {
		if err := validateTimestamp(*m.PublishedAt); err != nil {
			return fmt.Errorf("published at: %w", err)
		}
	}
	if err := validateTimestamp(m.SourceModifiedAt); err != nil {
		return fmt.Errorf("source updated at: %w", err)
	}
	return nil
}

func validateLabel(label []Label) error {
	seen := make(map[uuid.UUID]struct{}, len(label))
	for i, taxon := range label {
		if taxon.ID == uuid.Nil() {
			return fmt.Errorf("item %d: ID must not be nil UUID", i)
		}
		if strings.TrimSpace(taxon.Title) == "" {
			return fmt.Errorf("item %d: title must not be empty", i)
		}
		if _, exists := seen[taxon.ID]; exists {
			return fmt.Errorf("item %d: duplicate ID %s", i, taxon.ID)
		}
		seen[taxon.ID] = struct{}{}
	}
	return nil
}

func validateTimestamp(value time.Time) error {
	if value.IsZero() {
		return errors.New("timestamp must not be zero")
	}
	if _, offset := value.Zone(); offset != 0 {
		return errors.New("timestamp must be normalized to UTC")
	}
	return nil
}
