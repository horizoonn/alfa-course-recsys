package recommendation

import (
	"strings"

	"golang.org/x/net/html"
)

func analysisText(value string) string {
	if !strings.ContainsAny(value, "<>") {
		return strings.Join(strings.Fields(value), " ")
	}
	document, err := html.Parse(strings.NewReader(value))
	if err != nil {
		return ""
	}
	var result strings.Builder
	var visit func(*html.Node)
	visit = func(node *html.Node) {
		if node.Type == html.ElementNode {
			switch node.Data {
			case "head", "script", "style", "template":
				return
			}
		}
		if node.Type == html.TextNode {
			result.WriteString(node.Data)
			result.WriteByte(' ')
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(document)
	return strings.Join(strings.Fields(result.String()), " ")
}
