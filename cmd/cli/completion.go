package cli

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sethvargo/go-retry"
	"golang.org/x/exp/slices"
)

type oaiClients struct {
	openAIClient openai.Client
}

func newOAIClients() (oaiClients, error) {
	var config openai.ClientConfig
	config = openai.DefaultConfig(*openAIAPIKey)

	if openAIEndpoint != &openaiAPIURLv1 {
		// Azure OpenAI
		if strings.Contains(*openAIEndpoint, "openai.azure.com") {
			config = openai.DefaultAzureConfig(*openAIAPIKey, *openAIEndpoint)
			if len(*azureModelMap) != 0 {
				config.AzureModelMapperFunc = func(model string) string {
					return (*azureModelMap)[model]
				}
			}
		} else {
			// Local AI
			config.BaseURL = *openAIEndpoint
		}
	}

	clients := oaiClients{
		openAIClient: *openai.NewClientWithConfig(config),
	}
	return clients, nil
}

func getNonChatModels() []string {
	return []string{"code-davinci-002", "text-davinci-003"}
}

func gptCompletion(ctx context.Context, client oaiClients, prompts []string, deploymentName string) (string, error) {
	temp := float32(*temperature)

	var prompt strings.Builder
	fmt.Fprintf(&prompt, "You are a Kubernetes YAML generator, only generate valid Kubernetes YAML manifests. Do not provide any explanations, only generate YAML.")
	for _, p := range prompts {
		fmt.Fprintf(&prompt, "%s\n", p)
	}

	var resp string
	var err error
	r := retry.WithMaxRetries(10, retry.NewExponential(1*time.Second))
	if err := retry.Do(ctx, r, func(ctx context.Context) error {
		if slices.Contains(getNonChatModels(), deploymentName) {
			resp, err = client.openaiGptCompletion(ctx, prompt, temp)
		} else {
			resp, err = client.openaiGptChatCompletion(ctx, prompt, temp)
		}

		requestErr := &openai.RequestError{}
		if errors.As(err, &requestErr) {
			if requestErr.HTTPStatusCode == http.StatusTooManyRequests {
				return retry.RetryableError(err)
			}
		}
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return "", err
	}

	return resp, nil
}
