package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"

	"github.com/janeczku/go-spinner"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/walles/env"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const (
	apply     = "Apply"
	dontApply = "Don't Apply"
	reprompt  = "Reprompt"
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

	oaiClients, err := newOAIClients()
	if err != nil {
		return err
	}

	var action, completion string
	for action != apply {
		args = append(args, action)

		s := spinner.NewSpinner("Processing...")
		if !*debug && !*raw {
			s.SetCharset([]string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"})
			s.Start()
		}

		completion, err = gptCompletion(ctx, oaiClients, args)
		if err != nil {
			return err
		}

		s.Stop()

		if *raw {
			fmt.Println(completion)
			return nil
		}

		text := fmt.Sprintf("✨ Attempting to apply the following manifest:\n%s", completion)
		fmt.Println(text)

		action, err = userActionPrompt()
		if err != nil {
			return err
		}

		if action == dontApply {
			return nil
		}
	}

	return applyManifest(completion)
}

func userActionPrompt() (string, error) {
	// if require confirmation is not set, immediately return apply
	if !*requireConfirmation {
		return apply, nil
	}

	var result string
	var err error
	items := []string{apply, dontApply}
	currentContext, err := getCurrentContextName()
	label := fmt.Sprintf("Would you like to apply this? [%[1]s/%[2]s/%[3]s]", reprompt, apply, dontApply)
	if err == nil {
		label = fmt.Sprintf("(context: %[1]s) %[2]s", currentContext, label)
	}

	prompt := promptui.SelectWithAdd{
		Label:    label,
		Items:    items,
		AddLabel: reprompt,
	}
	_, result, err = prompt.Run()
	if err != nil {
		// workaround for bug in promptui when input is piped in from stdin
		// however, this will not block for ui input
		// for now, we will not apply the yaml, but user can pipe to input to kubectl
		if err.Error() == "^D" {
			return dontApply, nil
		}
		return dontApply, err
	}

	return result, nil
}
