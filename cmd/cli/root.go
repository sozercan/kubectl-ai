package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/walles/env"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	openaiAPIURLv1        = "https://api.openai.com/v1"
	version               = "dev"
	kubernetesConfigFlags = genericclioptions.NewConfigFlags(false)

	openAIDeploymentName = flag.String("openai-deployment-name", env.GetOr("OPENAI_DEPLOYMENT_NAME", env.String, "gpt-3.5-turbo-0301"), "The deployment name used for the model in OpenAI service.")
	openAIAPIKey         = flag.String("openai-api-key", env.GetOr("OPENAI_API_KEY", env.String, ""), "The API key for the OpenAI service. This is required.")
	openAIEndpoint       = flag.String("openai-endpoint", env.GetOr("OPENAI_ENDPOINT", env.String, openaiAPIURLv1), "The endpoint for OpenAI service. Defaults to"+openaiAPIURLv1+". Set this to your Local AI endpoint or Azure OpenAI Service, if needed.")
	azureModelMap        = flag.StringToString("azure-openai-map", env.GetOr("AZURE_OPENAI_MAP", env.Map(env.String, "=", env.String, ""), map[string]string{}), "The mapping from OpenAI model to Azure OpenAI deployment. Defaults to empty map. Example format: gpt-3.5-turbo=my-deployment.")
	requireConfirmation  = flag.Bool("require-confirmation", env.GetOr("REQUIRE_CONFIRMATION", strconv.ParseBool, true), "Whether to require confirmation before executing the command. Defaults to true.")
	temperature          = flag.Float64("temperature", env.GetOr("TEMPERATURE", env.WithBitSize(strconv.ParseFloat, 64), 0.0), "The temperature to use for the model. Range is between 0 and 1. Set closer to 0 if your want output to be more deterministic but less creative. Defaults to 0.0.")
	raw                  = flag.Bool("raw", false, "Prints the raw YAML output immediately. Defaults to false.")
	usek8sAPI            = flag.Bool("use-k8s-api", env.GetOr("USE_K8S_API", strconv.ParseBool, false), "Whether to use the Kubernetes API to create resources with function calling. Defaults to false.")
	k8sOpenAPIURL        = flag.String("k8s-openapi-url", env.GetOr("K8S_OPENAPI_URL", env.String, ""), "The URL to a Kubernetes OpenAPI spec. Only used if use-k8s-api flag is true.")
	debug                = flag.Bool("debug", env.GetOr("DEBUG", strconv.ParseBool, false), "Whether to print debug logs. Defaults to false.")
)

func InitAndExecute() {
	if *openAIAPIKey == "" {
		fmt.Println("Please provide an OpenAI key.")
		os.Exit(1)
	}

	if err := RootCmd().Execute(); err != nil {
		handleError(err)
		os.Exit(1)
	}
}

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "kubectl-ai",
		Short:        "kubectl-ai",
		Long:         "kubectl-ai is a plugin for kubectl that allows you to interact with OpenAI GPT API.",
		Version:      version,
		SilenceUsage: true,
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			if *debug {
				log.SetLevel(log.DebugLevel)
				printDebugFlags()
			}
		},
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("prompt must be provided")
			}

			err := run(args)
			if err != nil {
				return err
			}

			return nil
		},
	}

	kubernetesConfigFlags.AddFlags(cmd.PersistentFlags())

	return cmd
}

func printDebugFlags() {
	log.Debugf("openai-endpoint: %s", *openAIEndpoint)
	log.Debugf("openai-deployment-name: %s", *openAIDeploymentName)
	log.Debugf("azure-openai-map: %s", *azureModelMap)
	log.Debugf("temperature: %f", *temperature)
	log.Debugf("use-k8s-api: %t", *usek8sAPI)
	log.Debugf("k8s-openapi-url: %s", *k8sOpenAPIURL)
}

func run(args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	var k8sContext string
	currentContext, err := getCurrentContextName()
	if err == nil {
		log.Debugf("current-context: %s", currentContext)
		k8sContext = currentContext
	}

	p := tea.NewProgram(newModel(args, k8sContext, !*requireConfirmation), tea.WithContext(ctx))
	m, err := p.Run()
	if err != nil {
		return uiError{err, "Couldn't start Bubble Tea program."}
	}

	model, ok := m.(model)
	if !ok {
		return fmt.Errorf("unexpected model type %T", m)
	} else if model.error != nil {
		return *model.error
	}

	// Create a manifest from the last completion
	manifest := trimTicks(model.completion)

	if model.state == apply || model.state == autoApply {
		return applyManifest(manifest)
	}

	return nil
}

func handleError(err error) {
	format := "\n%s\n\n"

	var args []interface{}
	var merr uiError

	if errors.As(err, &merr) {
		args = []interface{}{
			stderrStyles().ErrPadding.Render(stderrStyles().ErrorHeader.String(), merr.reason),
		}
	} else {
		args = []interface{}{
			stderrStyles().ErrPadding.Render(stderrStyles().ErrorDetails.Render(err.Error())),
		}
	}

	fmt.Fprintf(os.Stderr, format, args...)
}
