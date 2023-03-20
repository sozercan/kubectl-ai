# Kubectl OpenAI plugin ✨

This project is a `kubectl` plugin to generate and apply Kubernetes manifests using OpenAI GPT-3.

Main motivation for me is to avoid finding random manifests when dev/testing things.

## Usage

### Prerequisites

`kubectl-ai` requires an [OpenAI API key](https://platform.openai.com/overview) or an [Azure OpenAI Service](https://aka.ms/azure-openai) API key and endpoint. Also, `kubectl` is required with a valid kubeconfig.

> I don't have a non-Azure OpenAI API key, non-Azure OpenAI API is not tested.

For both OpenAI and Azure OpenAI, you can use the following environment variables:

```shell
export OPENAI_API_KEY=<your OpenAI key>
```

For Azure OpenAI Service, you can use the following environment variables:

```shell
export AZURE_OPENAI_ENDPOINT=<your Azure OpenAI endpoint, like https://my-aoi-endpoint.openai.azure.com>
export OPENAI_DEPLOYMENT_NAME=<your OpenAI deployment/model name. defaults to "text-davinci-003">
```

If `AZURE_OPENAI_ENDPOINT` variable is set, then it will use the Azure OpenAI Service. Otherwise, it will use OpenAI API.

### Installation

- Download the binary from [GitHub releases](https://github.com/sozercan/kubectl-ai/releases).

- If you want to use this as a [`kubectl` plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/), then copy `kubectl-ai` binary to your `PATH`. If not, you can also use the binary standalone.

### Flags and environment variables

- `--require-confirmation` flag or `REQUIRE_CONFIRMATION` environment varible can be set to prompt the user for confirmation before applying the manifest. Defaults to false.

- `--temperature` flag or `TEMPERATURE` environment variable can be set between 0 and 1. Higher temperature will result in more creative completions. Lower temperature will result in more deterministic completions. Defaults to 0.

## Examples

Creating objects with specific values:

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
EOF
```

```shell
$ kubectl ai "scale nginx-deployment to 5 replicas"
✨ Attempting to apply the following manifest:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
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
          image: nginx
EOF
```

Optional `--require-confirmation` flag:

```shell
$ kubectl ai "create a service with type LoadBalancer with selector as 'app:nginx'" --require-confirmation
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
Use the arrow keys to navigate: ↓ ↑ → ←
? Would you like to apply this? [Yes/No]:
  ▸ Yes
    No
```

Multiple objects:

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
EOF
```

## Acknowledgements and Credits

Thanks to @simongottschlag for their work on Azure OpenAI fork in https://github.com/simongottschlag/azure-openai-gpt-slack-bot
which is based on https://github.com/PullRequestInc/go-gpt3
