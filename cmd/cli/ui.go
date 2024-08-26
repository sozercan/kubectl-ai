package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	openai "github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
)

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type model struct {
	list         list.Model
	choice       string
	spinner      spinner.Model
	glam         *glamour.TermRenderer
	renderedYaml string

	pipedInput    string
	promptArgs    []string
	completion    string
	retries       int
	oaiClients    oaiClients
	cancelRequest context.CancelFunc

	config *config
}

type config struct {
	maxRetries int
}

func defaultConfig() *config {
	return &config{
		maxRetries: maxRetries,
	}
}

func newModel(promptArgs []string) model {
	s := spinner.New(spinner.WithSpinner(spinner.Dot))

	items := []list.Item{
		item("Apply"),
		item("Don't Apply"),
		item("Reprompt"),
	}

	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Would you like to apply this?"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	// Discard the error, because we can recover from it by falling back to the default style.
	gr, _ := glamour.NewTermRenderer(glamour.WithAutoStyle())

	return model{
		config:     defaultConfig(),
		oaiClients: newOAIClients(),
		spinner:    s,
		glam:       gr,
		list:       l,
		promptArgs: promptArgs,
	}
}

// Init implements tea.Model.
func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.readStdinCmd())
}

// Update implements tea.Model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds    []tea.Cmd
		listCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
			}
			return m, tea.Quit
		}
	case completionInput:
		return m, m.startCompletionCmd(msg.content)
	case completionOutput:
		// update the model with the latest completion
		m.completion = msg.content

		// TODO - we'll probably handle the raw flag here.
		// if *raw {
		// 	fmt.Println(msg.content)
		// 	return nil
		// }

		renderedYaml, err := m.glam.Render(msg.content)
		if err != nil {
			// TODO - maybe return a uiError here instead, so we can abort
			renderedYaml = fmt.Sprintf("There was an error rendering the completion: %s", err)
		}
		m.renderedYaml = renderedYaml
	case uiError:
		// TODO - should I add the error rather than return?
		// TODO - ensure the state is set to show any errors
		//return m, m.errorCmd(msg.err, msg.reason)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	m.list, listCmd = m.list.Update(msg)
	cmds = append(cmds, listCmd)
	return m, tea.Batch(cmds...)
}

// View implements tea.Model.
func (m model) View() string {
	s := strings.Builder{}

	s.WriteString("✨ Attempting to apply the following manifest:\n")

	s.WriteString(m.spinner.View() + "Processing..." + "\n")

	s.WriteString(m.renderedYaml + "\n")

	s.WriteString(m.list.View())

	// currentContext, err := getCurrentContextName()
	// label := fmt.Sprintf("Would you like to apply this? [%[1]s/%[2]s/%[3]s]", reprompt, apply, dontApply)
	// if err == nil {
	// 	label = fmt.Sprintf("(context: %[1]s) %[2]s", currentContext, label)
	// }

	return s.String()
}

// completionInput is a tea.Msg that wraps the content read from stdin.
type completionInput struct {
	content string
}

// completionOutput a tea.Msg that wraps the content returned from OpenAI.
type completionOutput struct {
	content string
}

// uiError is a wrapper around an error that adds additional context.
type uiError struct {
	err    error
	reason string
}

func (u uiError) Error() string {
	return u.err.Error()
}

func (u uiError) Reason() string {
	return u.reason
}

func (m *model) readStdinCmd() tea.Cmd {
	return func() tea.Msg {
		if !isInputTTY() {
			reader := bufio.NewReader(os.Stdin)
			stdinBytes, err := io.ReadAll(reader)
			if err != nil {
				return uiError{err, "Unable to read stdin."}
			}

			return completionInput{increaseIndent(string(stdinBytes))}
		}
		return completionInput{""}
	}
}

func (m *model) retry(content string, err uiError) tea.Msg {
	m.retries++
	if m.retries >= m.config.maxRetries {
		return err
	}
	wait := time.Millisecond * 100 * time.Duration(math.Pow(2, float64(m.retries))) //nolint:mnd
	time.Sleep(wait)
	return completionInput{content}
}

func (m *model) startCompletionCmd(content string) tea.Cmd {
	return func() tea.Msg {
		temp := float32(*temperature)
		var prompt strings.Builder

		// did we receive any piped input?
		if m.pipedInput != "" {
			fmt.Fprintf(&prompt, "Depending on the input, either edit or append to the input YAML. Do not generate new YAML without including the input YAML either original or edited.\nUse the following YAML as the input: \n%s\n", m.pipedInput)
		}

		for _, p := range m.promptArgs {
			fmt.Fprintf(&prompt, "%s", p)
		}

		ctx, cancel := context.WithCancel(context.Background())
		m.cancelRequest = cancel

		resp, err := m.oaiClients.openaiGptChatCompletion(ctx, &prompt, temp)
		if err != nil {
			return m.handleRequestError(err, content)
		}

		// TODO - do this later, because we need them in the output for glamour to render properly
		// remove unnecessary backticks if they are in the output
		//cleanedResp := trimTicks(resp)

		return completionOutput{content: resp}
	}
}

func (m *model) handleRequestError(err error, content string) tea.Msg {
	ae := &openai.APIError{}
	if errors.As(err, &ae) {
		//return m.handleAPIError(ae, mod, content)
		switch ae.HTTPStatusCode {
		case http.StatusTooManyRequests, http.StatusRequestTimeout, http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			log.Debugf("retrying due to status code %d: %s", ae.HTTPStatusCode, ae.Message)
			//return retry.RetryableError(err)
			return m.retry(content, uiError{err: err, reason: fmt.Sprintf("retrying due to status code %d: %s", ae.HTTPStatusCode, ae.Message)})
		}
	}
	return uiError{ae, fmt.Sprintf(
		"There was a problem with the API request: %s",
		err.Error(),
	)}
}

func increaseIndent(s string) string {
	lines := strings.Split(s, "\n")
	for i := 0; i < len(lines); i++ {
		lines[i] = "\t" + lines[i]
	}
	return strings.Join(lines, "\n")
}
