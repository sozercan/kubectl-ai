package cli

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"

	"github.com/manifoldco/promptui"
	"github.com/sozercan/kubectl-ai/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const version = "0.0.1"

var (
	kubernetesConfigFlags *genericclioptions.ConfigFlags

	openAIDeploymentName = flag.String("openai-deployment-name", utils.LookupEnvOrString("OPENAI_DEPLOYMENT_NAME", "text-davinci-003"), "The deployment name used for the model in OpenAI service.")
	openAIAPIKey         = flag.String("openai-api-key", utils.LookupEnvOrString("OPENAI_API_KEY", ""), "The API key for the OpenAI service. This is required.")
	azureOpenAIEndpoint  = flag.String("azure-openai-endpoint", utils.LookupEnvOrString("AZURE_OPENAI_ENDPOINT", ""), "The endpoint for Azure OpenAI service. If provided, Azure OpenAI service will be used instead of OpenAI service.")
	requireConfirmation  = flag.Bool("require-confirmation", false, "Whether to require confirmation before executing the command. Defaults to false.")
	temperature          = flag.Float64("temperature", 1, "The temperature to use for the model. Range is between 0 and 1. Set closer to 0 if your want output to be more deterministic but less creative. Defaults to 1.")
)

var maxTokensMap = map[string]int{
	"text-davinci-003": 4097,
	"code-davinci-002": 8001,
}

func InitAndExecute() {
	flag.Parse()

	if *openAIAPIKey == "" {
		fmt.Println("Please provide an OpenAI key.")
		os.Exit(1)
	}

	if err := RootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "kubectl-ai",
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("prompt must be provided")
			}

			err := run(args)
			if err != nil {
				return fmt.Errorf("application returned an error: %w", err)
			}

			return nil
		},
	}

	kubernetesConfigFlags = genericclioptions.NewConfigFlags(false)
	kubernetesConfigFlags.AddFlags(cmd.Flags())
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	return cmd
}

func run(args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	oaiClients, err := newOAIClients()
	if err != nil {
		return err
	}

	completion, err := gptCompletion(ctx, oaiClients, args, *openAIDeploymentName)
	if err != nil {
		return err
	}

	text := fmt.Sprintf("✨ Attempting to run the following command: %s", completion)
	fmt.Println(text)

	conf, err := yesNo()
	if err != nil {
		return err
	}

	if conf {
		var stderr bytes.Buffer
		cmd := exec.Command("sh", "-c", completion)
		cmd.Stderr = &stderr

		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("%w: %s", err, stderr.String())
		}
	}
	return nil
}

func yesNo() (bool, error) {
	result := "Yes"
	var err error
	if *requireConfirmation {
		prompt := promptui.Select{
			Label: "Would you like to apply this? [Yes/No]",
			Items: []string{"Yes", "No"},
		}
		_, result, err = prompt.Run()
		if err != nil {
			return false, err
		}
	}
	return result == "Yes", nil
}
