package cli

import (
	"os"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"github.com/muesli/termenv"
)

var isInputTTY = sync.OnceValue(func() bool {
	return term.IsTerminal(os.Stdin.Fd())
})

var stderrRenderer = sync.OnceValue(func() *lipgloss.Renderer {
	return lipgloss.NewRenderer(os.Stderr, termenv.WithColorCache(true))
})

var stderrStyles = sync.OnceValue(func() styles {
	return makeStyles(stderrRenderer())
})

type styles struct {
	ErrorHeader,
	ErrorDetails,
	ErrPadding lipgloss.Style
}

func makeStyles(r *lipgloss.Renderer) (s styles) {
	const horizontalEdgePadding = 2
	s.ErrorHeader = r.NewStyle().Foreground(lipgloss.Color("#F1F1F1")).Background(lipgloss.Color("#FF5F87")).Bold(true).Padding(0, 1).SetString("ERROR")
	s.ErrorDetails = r.NewStyle().Foreground(lipgloss.Color("#757575"))
	s.ErrPadding = r.NewStyle().Padding(0, horizontalEdgePadding)
	return s
}
