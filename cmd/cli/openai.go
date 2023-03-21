package cli

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/PullRequestInc/go-gpt3"
	gptEncoder "github.com/samber/go-gpt-3-encoder"
	azureopenai "github.com/sozercan/kubectl-ai/pkg/gpt3"
	"github.com/sozercan/kubectl-ai/pkg/utils"
)

type oaiClients struct {
	azureClient  azureopenai.Client
	openAIClient openai.Client
}

func newOAIClients() (oaiClients, error) {
	var oaiClient openai.Client
	var azureClient azureopenai.Client
	var err error

	if azureOpenAIEndpoint == nil || *azureOpenAIEndpoint == "" {
		oaiClient = openai.NewClient(*openAIAPIKey)
	} else {
		azureClient, err = azureopenai.NewClient(*azureOpenAIEndpoint, *openAIAPIKey, *openAIDeploymentName)
		if err != nil {
			return oaiClients{}, err
		}
	}

	clients := oaiClients{
		azureClient:  azureClient,
		openAIClient: oaiClient,
	}
	return clients, nil
}

func gptCompletion(ctx context.Context, client oaiClients, prompts []string, deploymentName string) (string, error) {
	temp := float32(*temperature)
	maxTokens, err := calculateMaxTokens(prompts, deploymentName)
	if err != nil {
		return "", err
	}

	var prompt strings.Builder
	fmt.Fprintf(&prompt, "You are a Kubernetes YAML generator, only generate valid Kubernetes YAML manifests.")
	for _, p := range prompts {
		fmt.Fprintf(&prompt, "%s\n", p)
	}

	if azureOpenAIEndpoint == nil || *azureOpenAIEndpoint == "" {
		resp, err := client.openAIClient.CompletionWithEngine(ctx, *openAIDeploymentName, openai.CompletionRequest{
			Prompt:      []string{prompt.String()},
			MaxTokens:   maxTokens,
			Echo:        false,
			N:           utils.ToPtr(1),
			Temperature: &temp,
		})
		if err != nil {
			return "", err
		}

		if len(resp.Choices) != 1 {
			return "", fmt.Errorf("expected choices to be 1 but received: %d", len(resp.Choices))
		}

		return resp.Choices[0].Text, nil
	}

	resp, err := client.azureClient.Completion(ctx, azureopenai.CompletionRequest{
		Prompt:      []string{prompt.String()},
		MaxTokens:   maxTokens,
		Echo:        false,
		N:           utils.ToPtr(1),
		Temperature: &temp,
	})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) != 1 {
		return "", fmt.Errorf("expected choices to be 1 but received: %d", len(resp.Choices))
	}

	return resp.Choices[0].Text, nil
}

func calculateMaxTokens(prompts []string, deploymentName string) (*int, error) {
	maxTokens, ok := maxTokensMap[deploymentName]
	if !ok {
		return nil, fmt.Errorf("deploymentName %q not found in max tokens map", deploymentName)
	}

	encoder, err := gptEncoder.NewEncoder()
	if err != nil {
		return nil, err
	}

	// start at 100 since the encoder at times doesn't get it exactly correct
	totalTokens := 100
	for _, prompt := range prompts {
		tokens, err := encoder.Encode(prompt)
		if err != nil {
			return nil, err
		}
		totalTokens += len(tokens)
	}

	remainingTokens := maxTokens - totalTokens
	return &remainingTokens, nil
}
