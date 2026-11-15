package converter

import (
	"regexp"
	"strings"
	"time"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
)

var durationPattern = regexp.MustCompile(
	`^(?:~\s*)?(?:(\d+(?:[.,]\d+)?)\s*час(?:а|ов)?\s*)?(?:(\d+)\s*минут(?:а|ы)?)?$`,
)

func materialDuration(value *string) domain.MaterialDuration {
	if value == nil {
		return domain.MaterialDuration{}
	}

	result := domain.MaterialDuration{Text: strings.TrimSpace(*value)}
	parts := durationPattern.FindStringSubmatch(strings.ToLower(normalizeText(result.Text)))
	if parts == nil || (parts[1] == "" && parts[2] == "") {
		return result
	}

	var durationText string
	if parts[1] != "" {
		durationText += strings.ReplaceAll(parts[1], ",", ".") + "h"
	}
	if parts[2] != "" {
		durationText += parts[2] + "m"
	}

	duration, err := time.ParseDuration(durationText)
	if err != nil || duration <= 0 || duration%time.Minute != 0 {
		return result
	}

	minutes := int(duration / time.Minute)
	result.Minutes = &minutes
	return result
}
