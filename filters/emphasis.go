package filters

import "regexp"

type EmphasisFilter struct{}

func (e EmphasisFilter) Apply(input string) string {
	emphasisRegex := regexp.MustCompile(`\*.*?\*`) // Matches text between asterisks
	return emphasisRegex.ReplaceAllString(input, "")
}

func (e EmphasisFilter) Name() string {
	return "remove_emphasis"
}
