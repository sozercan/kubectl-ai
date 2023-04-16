# Kubectl OpenAI plugin ✨

This project is a `kubectl` plugin to generate and apply Kubernetes manifests using OpenAI GPT.

My main motivation is to avoid finding and collecting random manifests when dev/testing things.

## Demo

[![asciicast](https://asciinema.org/a/MEXrlAqUjo7DMnfoyQearpVQ7.svg)](https://asciinema.org/a/MEXrlAqUjo7DMnfoyQearpVQ7)

## Installation

### Homebrew

Add to `brew` tap and install with:

```shell
brew tap sozercan/kubectl-ai https://github.com/sozercan/kubectl-ai
brew install kubectl-ai
```

### Krew

Add to `krew` index and install with:

```shell
kubectl krew index add kubectl-ai https://github.com/sozercan/kubectl-ai
kubectl krew install kubectl-ai/kubectl-ai
```

### GitHub release
- Download the binary from [GitHub releases](https://github.com/sozercan/kubectl-ai/releases).

- If you want to use this as a [`kubectl` plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/), then copy `kubectl-ai` binary to your `PATH`. If not, you can also use the binary standalone.

## Usage

### Prerequisites

`kubectl-ai` requires an [OpenAI API key](https://platform.openai.com/overview) or an [Azure OpenAI Service](https://aka.ms/azure-openai) API key and endpoint, and a valid Kubernetes configuration.

For both OpenAI and Azure OpenAI, you can use the following environment variables:

```shell
export OPENAI_API_KEY=<your OpenAI key>
export OPENAI_DEPLOYMENT_NAME=<your OpenAI deployment/model name. defaults to "gpt-3.5-turbo">
```

> Following models are supported:
> - `code-davinci-002`
> - `text-davinci-003`
> - `gpt-3.5-turbo-0301` (deployment must be named `gpt-35-turbo-0301` for Azure )
> - `gpt-3.5-turbo`
> - `gpt-35-turbo-0301`
> - `gpt-4-0314`
> - `gpt-4-32k-0314`

For Azure OpenAI Service, you can use the following environment variables:

```shell
export AZURE_OPENAI_ENDPOINT=<your Azure OpenAI endpoint, like "https://my-aoi-endpoint.openai.azure.com">
```

If `AZURE_OPENAI_ENDPOINT` variable is set, then it will use the Azure OpenAI Service. Otherwise, it will use OpenAI API.

### Flags and environment variables

- `--require-confirmation` flag or `REQUIRE_CONFIRMATION` environment varible can be set to prompt the user for confirmation before applying the manifest. Defaults to true.

- `--temperature` flag or `TEMPERATURE` environment variable can be set between 0 and 1. Higher temperature will result in more creative completions. Lower temperature will result in more deterministic completions. Defaults to 0.

## Examples

### Creating objects with specific values

```shell
$ kubectl ai "create an nginx deployment with 3 replicas"
✨ Attempting to apply the following manifest:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        ports:
        - containerPort: 80
Use the arrow keys to navigate: ↓ ↑ → ←
? Would you like to apply this? [Reprompt/Apply/Don't Apply]:
+   Reprompt
  ▸ Apply
    Don't Apply
```

### Reprompt to refine your prompt

```shell
...
Reprompt: update to 5 replicas and port 8080
✨ Attempting to apply the following manifest:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 5
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        ports:
        - containerPort: 8080
Use the arrow keys to navigate: ↓ ↑ → ←
? Would you like to apply this? [Reprompt/Apply/Don't Apply]:
+   Reprompt
  ▸ Apply
    Don't Apply
```

### Multiple objects

```shell
$ kubectl ai "create a foo namespace then create nginx pod in that namespace"
✨ Attempting to apply the following manifest:
apiVersion: v1
kind: Namespace
metadata:
  name: foo
---
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  namespace: foo
spec:
  containers:
  - name: nginx
    image: nginx:latest
Use the arrow keys to navigate: ↓ ↑ → ←
? Would you like to apply this? [Reprompt/Apply/Don't Apply]:
+   Reprompt
  ▸ Apply
    Don't Apply
```

### Optional `--require-confirmation` flag

```shell
$ kubectl ai "create a service with type LoadBalancer with selector as 'app:nginx'" --require-confirmation=false
✨ Attempting to apply the following manifest:
apiVersion: v1
kind: Service
metadata:
  name: nginx-service
spec:
  selector:
    app: nginx
  ports:
  - port: 80
    targetPort: 80
  type: LoadBalancer
```

> Please note that the plugin does not know the current state of the cluster (yet?), so it will always generate the full manifest.

## Acknowledgements and Credits

Thanks to @simongottschlag for their work on Azure OpenAI fork in https://github.com/simongottschlag/azure-openai-gpt-slack-bot
which is based on https://github.com/PullRequestInc/go-gpt3
