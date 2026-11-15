package converter

import "strings"

const summaryMaxRunes = 300

func shortSummary(description *string) *string {
	if description == nil {
		return nil
	}
	text := []rune(*description)
	if len(text) <= summaryMaxRunes {
		return description
	}

	end := summaryMaxRunes - 1
	for end > 0 && text[end] != ' ' {
		end--
	}
	if end == 0 {
		end = summaryMaxRunes - 1
	}
	summary := strings.TrimSpace(string(text[:end])) + "…"
	return &summary
}
