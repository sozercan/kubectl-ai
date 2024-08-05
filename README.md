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

`kubectl-ai` requires a valid Kubernetes configuration and one of the following:

- [OpenAI API key](https://platform.openai.com/overview)
- [Azure OpenAI Service](https://aka.ms/azure-openai) API key and endpoint
- [OpenAI API-compatible endpoint](#set-up-a-local-openai-api-compatible-endpoint) (such as [AIKit](https://github.com/sozercan/aikit) or [LocalAI](https://localai.io/))

For OpenAI, Azure OpenAI or OpenAI API compatible endpoint, you can use the following environment variables:

```shell
export OPENAI_API_KEY=<your OpenAI key>
export OPENAI_DEPLOYMENT_NAME=<your OpenAI deployment/model name. defaults to "gpt-3.5-turbo-0301">
export OPENAI_ENDPOINT=<your OpenAI endpoint, like "https://my-aoi-endpoint.openai.azure.com" or "http://localhost:8080/v1">
```

If `OPENAI_ENDPOINT` variable is set, then it will use the endpoint. Otherwise, it will use OpenAI API.

Azure OpenAI service does not allow certain characters, such as `.`, in the deployment name. Consequently, `kubectl-ai` will automatically replace `gpt-3.5-turbo` to `gpt-35-turbo` for Azure. However, if you use an Azure OpenAI deployment name completely different from the model name, you can set `AZURE_OPENAI_MAP` environment variable to map the model name to the Azure OpenAI deployment name. For example:

```shell
export AZURE_OPENAI_MAP="gpt-3.5-turbo=my-deployment"
```

### Set up a local OpenAI API-compatible endpoint

If you don't have OpenAI API access, you can set up a local OpenAI API-compatible endpoint using [AIKit](https://github.com/sozercan/aikit) on your local machine without any GPUs! For more information, see the [AIKit documentaton](https://sozercan.github.io/aikit/).

```shell
docker run -d --rm -p 8080:8080 ghcr.io/sozercan/llama3.1:8b
export OPENAI_ENDPOINT="http://localhost:8080/v1"
export OPENAI_DEPLOYMENT_NAME="llama-3.1-8b-instruct"
export OPENAI_API_KEY="n/a"
```

After setting up the environment like above, you can use `kubectl-ai` as usual.

### Flags and environment variables

- `--require-confirmation` flag or `REQUIRE_CONFIRMATION` environment varible can be set to prompt the user for confirmation before applying the manifest. Defaults to true.

- `--temperature` flag or `TEMPERATURE` environment variable can be set between 0 and 1. Higher temperature will result in more creative completions. Lower temperature will result in more deterministic completions. Defaults to 0.

- `--use-k8s-api` flag or `USE_K8S_API` environment variable can be set to use Kubernetes OpenAPI Spec to generate the manifest. This will result in very accurate completions including CRDs (if present in configured cluster). This setting will use more OpenAI API calls and it requires [function calling](https://openai.com/blog/function-calling-and-other-api-updates) which is available in `0613` or later models only. Defaults to false. However, this is recommended for accuracy and completeness.

- `--k8s-openapi-url` flag or `K8S_OPENAPI_URL` environment variable can be set to use a custom Kubernetes OpenAPI Spec URL. This is only used if `--use-k8s-api` is set. By default, `kubectl-ai` will use the configured Kubernetes API Server to get the spec unless this setting is configured. You can use the [default Kubernetes OpenAPI Spec](https://raw.githubusercontent.com/kubernetes/kubernetes/master/api/openapi-spec/swagger.json) or generate a custom spec for completions that includes custom resource definitions (CRDs). You can generate custom OpenAPI Spec by using `kubectl get --raw /openapi/v2 > swagger.json`.

### Pipe Input and Output

Kubectl AI can be used with pipe input and output. For example:

```shell
$ cat foo-deployment.yaml | kubectl ai "change replicas to 5" --raw | kubectl apply -f -
```

#### Save to file

```shell
$ cat foo-deployment.yaml | kubectl ai "change replicas to 5" --raw > my-deployment-updated.yaml
```

#### Use with external editors

If you want to use an external editor to edit the generated manifest, you can set the `--raw` flag and pipe to the editor of your choice. For example:

```shell
# Visual Studio Code
$ kubectl ai "create a foo namespace" --raw | code -

# Vim
$ kubectl ai "create a foo namespace" --raw | vim -
```

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
