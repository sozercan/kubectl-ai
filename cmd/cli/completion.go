package cli

import (
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

const maxRetries = 10

type oaiClients struct {
	openAIClient openai.Client
}

func newOAIClients() oaiClients {
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
	return clients
}
