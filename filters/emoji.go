package filters

import "github.com/dlclark/regexp2"

type EmojiFilter struct{}

func (e EmojiFilter) Apply(input string) string {
	emojiRegex := regexp2.MustCompile(`[\p{So}\p{Sk}\p{Sc}\p{Sm}\p{S}]`, regexp2.None)
	result, _ := emojiRegex.Replace(input, "", -1, -1)
	return result
}

func (e EmojiFilter) Name() string {
	return "remove_emojis"
}
