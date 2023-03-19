package main

import (
	"github.com/sozercan/kubectl-ai/cmd/cli"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	cli.InitAndExecute()
}
