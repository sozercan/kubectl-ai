package cli

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

func (c *oaiClients) openaiGptCompletion(ctx context.Context, prompt strings.Builder, temp float32) (string, error) {
	req := openai.CompletionRequest{
		Prompt:      []string{prompt.String()},
		Echo:        false,
		N:           1,
		Temperature: temp,
	}

	resp, err := c.openAIClient.CreateCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) != 1 {
		return "", fmt.Errorf("expected choices to be 1 but received: %d", len(resp.Choices))
	}

	return resp.Choices[0].Text, nil
}

func (c *oaiClients) openaiGptChatCompletion(ctx context.Context, prompt strings.Builder, temp float32) (string, error) {
	req := openai.ChatCompletionRequest{
		Model: *openAIDeploymentName,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt.String(),
			},
		},
		N:           1,
		Temperature: temp,
	}

	resp, err := c.openAIClient.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) != 1 {
		return "", fmt.Errorf("expected choices to be 1 but received: %d", len(resp.Choices))
	}

	return resp.Choices[0].Message.Content, nil
}
