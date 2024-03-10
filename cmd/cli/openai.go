package cli

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
)

type toolChoiceType string

const (
	toolChoiceAuto toolChoiceType = "auto"
	toolChoiceNone toolChoiceType = "none"
)

func (c *oaiClients) openaiGptChatCompletion(ctx context.Context, prompt *strings.Builder, temp float32) (string, error) {
	var (
		resp    openai.ChatCompletionResponse
		req     openai.ChatCompletionRequest
		content string
		err     error
	)

	// if we are using the k8s API, we need to call the functions
	toolChoiseType := toolChoiceAuto

	for {
		prompt.WriteString(content)
		log.Debugf("prompt: %s", prompt.String())

		req = openai.ChatCompletionRequest{
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

		if *usek8sAPI {
			req.Tools = []openai.Tool{
				{
					Type:     "function",
					Function: &findSchemaNames,
				},
				{
					Type:     "function",
					Function: &getSchema,
				},
			}
			req.ToolChoice = toolChoiseType
		}

		resp, err = c.openAIClient.CreateChatCompletion(ctx, req)
		if err != nil {
			return "", err
		}

		if len(resp.Choices[0].Message.ToolCalls) == 0 {
			break
		}

		for _, tool := range resp.Choices[0].Message.ToolCalls {
			log.Debugf("calling tool: %s", tool.Function.Name)

			// if there is a tool call, we need to call it and get the result
			content, err = callTool(tool)
			if err != nil {
				return "", err
			}
		}
	}

	if len(resp.Choices) != 1 {
		return "", fmt.Errorf("expected choices to be 1 but received: %d", len(resp.Choices))
	}

	result := resp.Choices[0].Message.Content
	log.Debugf("result: %s", result)

	// remove unnessary backticks if they are in the output
	result = trimTicks(result)

	return result, nil
}

func trimTicks(str string) string {
	trimStr := []string{"```yaml", "```"}
	for _, t := range trimStr {
		str = strings.ReplaceAll(str, t, "")
	}
	return str
}
