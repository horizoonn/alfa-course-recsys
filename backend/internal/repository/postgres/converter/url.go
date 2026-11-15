package converter

import (
	"regexp"
	"strings"
)

var materialSlugPattern = regexp.MustCompile("^[a-zA-Z0-9_-]+$")

func materialURL(section, slug string) *string {
	slug = strings.TrimSpace(slug)
	if !materialSlugPattern.MatchString(slug) {
		return nil
	}

	value := "https://kurs.alfabank.ru/" + section + "/" + slug + "/"
	return &value
}
