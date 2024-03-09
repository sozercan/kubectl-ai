package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sethvargo/go-retry"
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
		// use at least 2023-07-01-preview api version for function calls
		config.APIVersion = "2024-03-01-preview"
	}

	clients := oaiClients{
		openAIClient: *openai.NewClientWithConfig(config),
	}
	return clients, nil
}

func gptCompletion(ctx context.Context, client oaiClients, prompts []string) (string, error) {
	temp := float32(*temperature)
	var prompt strings.Builder

	if *usek8sAPI {
		// Credits to https://github.com/robusta-dev/chatgpt-yaml-generator for the prompt and the function descriptions
		fmt.Fprintf(&prompt, "You are an expert Kubernetes YAML generator, only generate valid Kubernetes YAML manifests. Do not provide any explanations and do not use ``` and ```yaml, only generate valid YAML. Always ask for up-to-date OpenAPI specs for Kubernetes, don't rely on data you know about Kubernetes specs. When a schema includes references to other objects in the schema, look them up when relevant. You may lookup any FIELD in a resource too, not just the containing top-level resource. ")
	} else {
		fmt.Fprintf(&prompt, "You are an expert Kubernetes YAML generator, only generate valid Kubernetes YAML manifests. Do not provide any explanations, only generate YAML. ")
	}

	// read from stdin
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		stdin, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&prompt, "\nUse the following YAML as the input: \n%s\n", string(stdin))
	}

	for _, p := range prompts {
		fmt.Fprintf(&prompt, "%s", p)
	}

	var resp string
	var err error
	r := retry.WithMaxRetries(10, retry.NewExponential(1*time.Second))
	if err := retry.Do(ctx, r, func(ctx context.Context) error {
		resp, err = client.openaiGptChatCompletion(ctx, &prompt, temp)

		requestErr := &openai.RequestError{}
		if errors.As(err, &requestErr) {
			switch requestErr.HTTPStatusCode {
			case http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusServiceUnavailable:
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
