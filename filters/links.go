package filters

import "regexp"

type LinkFilter struct{}

func (l LinkFilter) Apply(input string) string {
	linkRegex := regexp.MustCompile(`https?://\S+`) // Matches URLs
	return linkRegex.ReplaceAllString(input, "")
}

func (l LinkFilter) Name() string {
	return "remove_links"
}
