package recommendation

import (
	"strings"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
)

type TextField struct {
	Name string `json:"name"`
	Text string `json:"text"`
}

type Document struct {
	Key    string      `json:"key"`
	Fields []TextField `json:"fields"`
}

func Documents(materials []domain.Material) []Document {
	result := make([]Document, 0, len(materials))
	for _, material := range materials {
		if material.Visibility != domain.VisibilityListed {
			continue
		}
		doc := Document{Key: material.Key.String(), Fields: []TextField{}}
		seen := map[string]bool{}
		add := func(name string, text *string) {
			if text == nil {
				return
			}
			value := analysisText(*text)
			key := name + ":" + strings.ToLower(value)
			if value != "" && !seen[key] {
				seen[key] = true
				doc.Fields = append(doc.Fields, TextField{
					Name: name,
					Text: value,
				})
			}
		}
		fragments := func(items []domain.ContentFragment) {
			for _, f := range items {
				name := string(f.Role)
				if name == "" {
					name = "body"
				}
				add(name, &f.Text)
			}
		}
		add("title", &material.Title)
		add("summary", material.Summary)
		m := material.Metadata
		for _, text := range []*string{
			m.Description, m.AdditionalDescription, m.CardDescription,
			m.AlternativeTitle, m.AlternativeDescription, m.NavigationTitle,
		} {
			add("metadata", text)
		}
		add("body", m.Content)
		for _, heading := range m.Headings {
			add("metadata", &heading)
		}
		for _, labels := range [][]domain.Label{material.Topics, material.Tags, m.AdditionalLabels} {
			for _, label := range labels {
				add("metadata", &label.Title)
			}
		}
		if m.Category != nil {
			add("metadata", &m.Category.Title)
		}
		fragments(m.Fragments)
		if course := m.Course; course != nil {
			add("metadata", course.Target)
			for _, info := range course.Info {
				add("metadata", &info.Title)
				add("metadata", &info.Type.Title)
				add("metadata", info.Description)
			}
			for _, lesson := range course.Program {
				if lesson.Visibility != domain.VisibilityListed {
					continue
				}
				add("body", &lesson.Title)
				add("body", lesson.NavigationTitle)
				add("body", lesson.Description)
				add("body", lesson.Content)
				fragments(lesson.Fragments)
			}
		}
		if m.Content == nil && len(m.Fragments) == 0 && m.Course == nil {
			add("body", &material.SearchText)
		}
		result = append(result, doc)
	}
	return result
}
