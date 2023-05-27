package cli

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"golang.org/x/exp/slices"
)

type oaiClients struct {
	openAIClient openai.Client
}

func newOAIClients() (oaiClients, error) {
	var config openai.ClientConfig
	config = openai.DefaultConfig(*openAIAPIKey)

	if azureOpenAIEndpoint != nil || *azureOpenAIEndpoint != "" {
		config = openai.DefaultAzureConfig(*openAIAPIKey, *azureOpenAIEndpoint)
		if len(*azureModelMap) != 0 {
			config.AzureModelMapperFunc = func(model string) string {
				return (*azureModelMap)[model]
			}
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

	if slices.Contains(getNonChatModels(), deploymentName) {
		resp, err := client.openaiGptCompletion(ctx, prompt, temp)
		if err != nil {
			return "", err
		}
		return resp, nil
	}

	resp, err := client.openaiGptChatCompletion(ctx, prompt, temp)
	if err != nil {
		return "", err
	}
	return resp, nil
}
