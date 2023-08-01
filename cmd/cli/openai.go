package cli

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
)

type functionCallType string

const (
	fnCallAuto functionCallType = "auto"
	fnCallNone functionCallType = "none"
)

func (c *oaiClients) openaiGptCompletion(ctx context.Context, prompt *strings.Builder, temp float32) (string, error) {
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

func (c *oaiClients) openaiGptChatCompletion(ctx context.Context, prompt *strings.Builder, temp float32) (string, error) {
	var (
		resp     openai.ChatCompletionResponse
		req      openai.ChatCompletionRequest
		funcName *openai.FunctionCall
		content  string
		err      error
	)

	// if we are using the k8s API, we need to call the functions
	fnCallType := fnCallAuto
	if !*usek8sAPI {
		fnCallType = fnCallNone
	}

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
			Functions: []openai.FunctionDefinition{
				findSchemaNames,
				getSchema,
			},
			FunctionCall: fnCallType,
		}

		resp, err = c.openAIClient.CreateChatCompletion(ctx, req)
		if err != nil {
			return "", err
		}

		funcName = resp.Choices[0].Message.FunctionCall
		// if there is no function call, we are done
		if funcName == nil {
			break
		}
		log.Debugf("calling function: %s", funcName.Name)

		// if there is a function call, we need to call it and get the result
		content, err = funcCall(funcName)
		if err != nil {
			return "", err
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
