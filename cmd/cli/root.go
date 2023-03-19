package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"

	"github.com/sozercan/kubectl-ai/pkg/gpt3"
	"github.com/sozercan/kubectl-ai/pkg/tokenizer"
	"github.com/sozercan/kubectl-ai/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const version = "0.0.1"

var (
	KubernetesConfigFlags     *genericclioptions.ConfigFlags
	AzureOpenAIEndpoint       = flag.String("azure-openai-endpoint", utils.LookupEnvOrString("AZURE_OPENAI_ENDPOINT", ""), "The endpoint for Azure OpenAI service")
	AzureOpenAIKey            = flag.String("azure-openai-key", utils.LookupEnvOrString("AZURE_OPENAI_KEY", ""), "The API key for Azure OpenAI service")
	AzureOpenAIDeploymentName = flag.String("azure-openai-deployment-name", utils.LookupEnvOrString("AZURE_OPENAI_DEPLOYMENT_NAME", "text-davinci-003"), "The deployment name used for the model in OpenAI service")
)

var maxTokensMap = map[string]int{
	"gpt-35-turbo-0301": 4097,
	"text-davinci-003":  4097,
}

func InitAndExecute() {
	flag.Parse()
	if err := RootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ai",
		Version: version,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := run(args)
			if err != nil {
				fmt.Fprintf(os.Stderr, "application returned an error: %v\n", err)
				os.Exit(1)
			}

			return nil
		},
	}

	KubernetesConfigFlags = genericclioptions.NewConfigFlags(false)
	KubernetesConfigFlags.AddFlags(cmd.Flags())

	return cmd
}

func run(args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	gptClient := gpt3.NewClient(*AzureOpenAIEndpoint, *AzureOpenAIKey, *AzureOpenAIDeploymentName)

	completion, err := gptCompletion(ctx, gptClient, args, *AzureOpenAIDeploymentName)
	if err != nil {
		return err
	}

	text := fmt.Sprintf("✨ Creating the following object: %s", completion)
	fmt.Println(text)

	err = exec.Command("sh", "-c", completion).Run()
	if err != nil {
		return err
	}

	return nil
}

func gptCompletion(ctx context.Context, client gpt3.Client, prompts []string, deploymentName string) (string, error) {
	maxTokens, err := calculateMaxTokens(prompts, deploymentName)
	if err != nil {
		return "", err
	}

	var prompt strings.Builder
	fmt.Fprintf(&prompt, "You are a shell, do not output anything but shell commands. All responses should be prefixed with 'cat <<EOF | kubectl apply -f -', followed by the Kubernetes object definition generated from the prompt, and finally 'EOF'.")
	for _, p := range prompts {
		fmt.Fprintf(&prompt, "%s\n", p)
	}
	resp, err := client.Completion(ctx, gpt3.CompletionRequest{
		Prompt:    []string{prompt.String()},
		MaxTokens: maxTokens,
		Echo:      false,
		N:         gpt3.ToPtr(1),
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

	encoder, err := tokenizer.NewEncoder()
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
