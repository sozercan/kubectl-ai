package cli

import (
	"os"
	"sync"

	"github.com/charmbracelet/x/term"
)

var isInputTTY = sync.OnceValue(func() bool {
	return term.IsTerminal(os.Stdin.Fd())
})
