package converter

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

func normalizeText(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func htmlToText(value string) (string, error) {
	document, err := html.Parse(strings.NewReader(value))
	if err != nil {
		return "", fmt.Errorf("parse HTML: %w", err)
	}

	var text strings.Builder
	writeHTMLText(&text, document)
	return normalizeText(text.String()), nil
}

func writeHTMLText(text *strings.Builder, node *html.Node) {
	if node.Type == html.TextNode {
		text.WriteString(node.Data)
		return
	}

	if node.Type == html.ElementNode {
		switch node.Data {
		case "head", "script", "style", "template":
			return
		}
	}

	separate := separatesText(node)
	if separate {
		text.WriteByte(' ')
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		writeHTMLText(text, child)
	}

	if separate {
		text.WriteByte(' ')
	}
}

func separatesText(node *html.Node) bool {
	if node.Type != html.ElementNode {
		return false
	}

	switch node.Data {
	case "address", "article", "aside", "blockquote", "br", "caption",
		"dd", "details", "div", "dl", "dt", "fieldset", "figcaption",
		"figure", "footer", "form", "h1", "h2", "h3", "h4", "h5", "h6",
		"header", "hgroup", "hr", "li", "main", "nav", "ol", "p", "pre",
		"section", "summary", "table", "tbody", "td", "tfoot", "th",
		"thead", "tr", "ul":
		return true
	default:
		return false
	}
}

func joinNormalizedText(parts ...string) string {
	result := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))

	for _, part := range parts {
		if part == "" {
			continue
		}
		if _, exists := seen[part]; exists {
			continue
		}

		seen[part] = struct{}{}
		result = append(result, part)
	}

	return strings.Join(result, "\n\n")
}

func normalizeOptionalText(value *string) *string {
	if value == nil {
		return nil
	}

	text := normalizeText(*value)
	if text == "" {
		return nil
	}
	return &text
}

func optionalHTMLToText(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	text, err := htmlToText(*value)
	if err != nil {
		return nil, err
	}
	if text == "" {
		return nil, nil
	}
	return &text, nil
}
