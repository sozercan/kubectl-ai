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
	log "github.com/sirupsen/logrus"
)

const maxRetries = 10

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

	// read from stdin
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		stdin, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&prompt, "Depending on the input, either edit or append to the input YAML. Do not generate new YAML without including the input YAML either original or edited.\nUse the following YAML as the input: \n%s\n", string(stdin))
	}

	for _, p := range prompts {
		fmt.Fprintf(&prompt, "%s", p)
	}

	var resp string
	var err error
	r := retry.WithMaxRetries(maxRetries, retry.NewExponential(1*time.Second))
	if err := retry.Do(ctx, r, func(ctx context.Context) error {
		resp, err = client.openaiGptChatCompletion(ctx, &prompt, temp)

		requestErr := &openai.APIError{}
		if errors.As(err, &requestErr) {
			switch requestErr.HTTPStatusCode {
			case http.StatusTooManyRequests, http.StatusRequestTimeout, http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
				log.Debugf("retrying due to status code %d: %s", requestErr.HTTPStatusCode, requestErr.Message)
				return retry.RetryableError(err)
			}
		}
		return nil
	}); err != nil {
		return "", err
	}

	return resp, nil
}
