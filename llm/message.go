package llm

import (
	"fmt"
	"github.com/openai/openai-go"
	"time"
)

type StoredMessage struct {
	Username  string    `json:"username"`
	Source    string    `json:"source"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

func (sm StoredMessage) ToOpenAIMessage() openai.ChatCompletionMessageParamUnion {
	switch sm.Role {
	case "user":
		return openai.UserMessage(sm.JSONify())
	case "assistant":
		return openai.AssistantMessage(sm.Content)
	case "system":
		return openai.SystemMessage(sm.Content)
	default:
		return openai.UserMessage(sm.Content)
	}
}

func (sm StoredMessage) JSONify() string {
	return fmt.Sprintf(`{"username": "%s", "source": "%s", "role": "%s", "content": "%s", "timestamp": "%s"}`,
		sm.Username, sm.Source, sm.Role, sm.Content, sm.Timestamp.Format(time.RFC1123))
}
