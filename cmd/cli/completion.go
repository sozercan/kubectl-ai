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

	if azureOpenAIEndpoint != nil && *azureOpenAIEndpoint != "" {
		config = openai.DefaultAzureConfig(*openAIAPIKey, *azureOpenAIEndpoint)
		if len(*azureModelMap) != 0 {
			config.AzureModelMapperFunc = func(model string) string {
				return (*azureModelMap)[model]
			}
		} else {
			// Local AI
			config.BaseURL = *openAIEndpoint
		}
		// use 2023-07-01-preview api version for function calls
		config.APIVersion = "2023-07-01-preview"
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

	if *usek8sAPI {
		// Credits to https://github.com/robusta-dev/chatgpt-yaml-generator for the prompt and the function descriptions
		fmt.Fprintf(&prompt, "You are a Kubernetes YAML generator, only generate valid Kubernetes YAML manifests. Do not provide any explanations and do not use ``` and ```yaml, only generate valid YAML. Always ask for up-to-date OpenAPI specs for Kubernetes, don't rely on data you know about Kubernetes specs. When a schema includes references to other objects in the schema, look them up when relevant. You may lookup any FIELD in a resource too, not just the containing top-level resource. ")
	} else {
		fmt.Fprintf(&prompt, "You are a Kubernetes YAML generator, only generate valid Kubernetes YAML manifests. Do not provide any explanations, only generate YAML. ")
	}

	for _, p := range prompts {
		fmt.Fprintf(&prompt, "%s", p)
	}

	var resp string
	var err error
	r := retry.WithMaxRetries(10, retry.NewExponential(1*time.Second))
	if err := retry.Do(ctx, r, func(ctx context.Context) error {
		if slices.Contains(getNonChatModels(), deploymentName) {
			resp, err = client.openaiGptCompletion(ctx, &prompt, temp)
		} else {
			resp, err = client.openaiGptChatCompletion(ctx, &prompt, temp)
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
