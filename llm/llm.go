package llm

import (
	"context"
	"errors"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"strings"
)

type OpenAIClient struct {
	client   *openai.Client
	endpoint string
	model    string
}

func NewOpenAIClient(apiKey, endpoint, model string) *OpenAIClient {
	opts := []option.RequestOption{
		option.WithAPIKey(apiKey),
		option.WithBaseURL(endpoint + "/v1/"),
	}
	c := openai.NewClient(opts...)

	return &OpenAIClient{
		client:   c,
		endpoint: endpoint,
		model:    model,
	}
}

func (o *OpenAIClient) GenerateResponse(ctx context.Context, messages []StoredMessage) (string, error) {
	sysMessage := openai.SystemMessage(
		NeighBotPrompt + "\n" + strings.ReplaceAll(BotPersonaPrompt, "{{.persona}}", DefaultPersona),
	)

	converted := []openai.ChatCompletionMessageParamUnion{sysMessage}
	for _, m := range messages {
		converted = append(converted, m.ToOpenAIMessage())
	}

	params := openai.ChatCompletionNewParams{
		Messages: openai.F(converted),
		Model:    openai.F(o.model),
	}

	completion, err := o.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", err
	}

	if len(completion.Choices) == 0 {
		return "", errors.New("no completions returned")
	}

	choice := completion.Choices[0].Message
	return choice.Content, nil
}

func (o *OpenAIClient) CountTokens(ctx context.Context, messages []StoredMessage) (int, error) {
	var concatted strings.Builder
	for _, m := range messages {
		concatted.WriteString(m.Content)
	}

	// Construct params interface
	params := map[string]interface{}{
		"content":     concatted.String(),
		"add_special": false,
		"with_pieces": false,
	}

	// Construct response interface
	response := struct {
		Tokens []int `json:"tokens"`
	}{}

	// POST to /tokenize endpoint to get token count
	if err := o.client.Post(ctx, o.endpoint+"/tokenize", params, &response); err != nil {
		return 0, err
	}

	return len(response.Tokens), nil
}
