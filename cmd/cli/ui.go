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
	"unicode"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	openai "github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
)

const (
	defaultWidth = 20
	tabWidth     = 4
	listHeight   = 8
	showList     = "Show List"
	showError    = "Show Error"
	apply        = "Apply"
	autoApply    = "Auto Apply"
	dontApply    = "Don't Apply"
	reprompt     = "Reprompt"
	rawOutput    = "Raw Output"
)

var (
	viewDefaultStyle  = lipgloss.NewStyle().Margin(0, 2).PaddingTop(1)
	titleStyle        = lipgloss.NewStyle().MarginLeft(0)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(2).PaddingBottom(1)
	contextStyle      = lipgloss.NewStyle().PaddingLeft(2).MarginBottom(1).Foreground(lipgloss.Color("170"))
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
	state   string
	loading bool
	error   *uiError

	spinner   spinner.Model
	glam      *glamour.TermRenderer
	list      list.Model
	choice    string
	textInput textinput.Model

	k8sContext    string
	autoApply     bool
	promptArgs    []string
	completion    string
	renderedYaml  string
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

func newModel(promptArgs []string, k8sContext string, autoApply bool) model {
	s := spinner.New(spinner.WithSpinner(spinner.Dot))

	items := []list.Item{
		item(apply),
		item(dontApply),
		item(reprompt),
	}

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Would you like to apply this?"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.SetShowPagination(false)
	l.Styles.HelpStyle = helpStyle

	ti := textinput.New()
	ti.Placeholder = "Enter your new prompt"
	ti.CharLimit = 156
	ti.Width = 38

	// Discard the error, because we are using the auto style which always exists.
	gr, _ := glamour.NewTermRenderer(glamour.WithAutoStyle())

	return model{
		state:      showList,
		config:     defaultConfig(),
		oaiClients: newOAIClients(),
		spinner:    s,
		glam:       gr,
		list:       l,
		textInput:  ti,
		promptArgs: promptArgs,
		k8sContext: k8sContext,
		autoApply:  autoApply,
	}
}

// Init implements tea.Model.
func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, textinput.Blink, m.readStdinCmd())
}

// Update implements tea.Model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds         []tea.Cmd
		listCmd      tea.Cmd
		textInputCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, m.quit
		case "enter":
			if m.state == showList {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					m.choice = string(i)
					switch m.choice {
					case apply:
						m.state = apply
						return m, m.quit
					case dontApply:
						m.state = dontApply
						return m, m.quit
					case reprompt:
						m.textInput.SetValue("")
						m.state = reprompt
						cmds = append(cmds, m.textInput.Focus())
					}
				}
			} else if m.state == reprompt {
				val := m.textInput.Value()
				m.promptArgs = append(m.promptArgs, removeWhitespace(val))
				m.state = showList
				m.loading = true
				return m, m.startCompletionCmd(m.promptArgs)
			}
		}

	case completionInput:
		m.loading = true

		if removeWhitespace(msg.content) != "" {
			m.promptArgs = append(m.promptArgs, removeWhitespace(msg.content))
		}

		return m, m.startCompletionCmd(m.promptArgs)

	case completionOutput:
		m.loading = false

		// update the model with the latest completion
		m.completion = msg.content

		// handle the raw flag
		if *raw {
			m.state = rawOutput
			return m, m.quit
		}

		return m, m.renderWithGlamour(msg.content)

	case renderedYamlMsg:
		m.renderedYaml = string(msg)

		// handle the auto apply setting
		if m.autoApply {
			m.state = autoApply
			return m, m.quit
		}

	case uiError:
		m.error = &msg
		m.state = showError
		return m, m.quit

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	switch m.state {
	case showList:
		m.list, listCmd = m.list.Update(msg)
		cmds = append(cmds, listCmd)
	case reprompt:
		m.textInput, textInputCmd = m.textInput.Update(msg)
		cmds = append(cmds, textInputCmd)
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model.
func (m model) View() string {
	s := strings.Builder{}

	if m.state == rawOutput {
		s.WriteString(trimTicks(m.completion))
		return s.String()
	}

	if m.loading {
		s.WriteString(m.spinner.View() + "Processing..." + "\n")
		return s.String()
	}

	contentView := lipgloss.JoinVertical(
		lipgloss.Left,
		viewDefaultStyle.Render("✨ Attempting to apply the following manifest:"),
		m.renderedYaml,
	)

	s.WriteString(contentView + "\n")

	if m.k8sContext != "" {
		s.WriteString(
			lipgloss.JoinHorizontal(lipgloss.Top,
				contextStyle.Render("☸️ Context: "),
				m.k8sContext,
			) + "\n",
		)
	}

	if m.state == autoApply {
		return s.String()
	}

	if m.state == reprompt {
		s.WriteString(m.textInput.View())
		return s.String()
	}

	s.WriteString(m.list.View())

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

func (m *model) quit() tea.Msg {
	if m.cancelRequest != nil {
		m.cancelRequest()
	}
	return tea.Quit()
}

func (m *model) readStdinCmd() tea.Cmd {
	return func() tea.Msg {
		if !isInputTTY() {
			reader := bufio.NewReader(os.Stdin)
			stdinBytes, err := io.ReadAll(reader)
			if err != nil {
				return uiError{err, "Unable to read stdin."}
			}

			var prompt strings.Builder
			pipedInput := string(stdinBytes)
			fmt.Fprintf(&prompt, "Depending on the input, either edit or append to the input YAML. Do not generate new YAML without including the input YAML either original or edited.\nUse the following YAML as the input: \n%s\n", pipedInput)

			return completionInput{prompt.String()}
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

func (m *model) startCompletionCmd(promptArgs []string) tea.Cmd {
	return func() tea.Msg {
		temp := float32(*temperature)
		var prompt strings.Builder

		for _, arg := range promptArgs {
			fmt.Fprintf(&prompt, "%s\n", arg)
		}

		ctx, cancel := context.WithCancel(context.Background())
		m.cancelRequest = cancel

		resp, err := m.oaiClients.openaiGptChatCompletion(ctx, &prompt, temp)
		if err != nil {
			return m.handleRequestError(err, prompt.String())
		}

		return completionOutput{content: resp}
	}
}

func (m *model) handleRequestError(err error, content string) tea.Msg {
	ae := &openai.APIError{}
	if errors.As(err, &ae) {
		switch ae.HTTPStatusCode {
		case http.StatusTooManyRequests, http.StatusRequestTimeout, http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			log.Debugf("retrying due to status code %d: %s", ae.HTTPStatusCode, ae.Message)
			return m.retry(content, uiError{err: err, reason: fmt.Sprintf("retrying due to status code %d: %s", ae.HTTPStatusCode, ae.Message)})
		}
	}
	return uiError{ae, fmt.Sprintf(
		"There was a problem with the API request: %s",
		err.Error(),
	)}
}

type renderedYamlMsg string

func (m *model) renderWithGlamour(md string) tea.Cmd {
	return func() tea.Msg {
		s, err := m.glam.Render(md)
		if err != nil {
			return uiError{
				err:    err,
				reason: fmt.Sprintf("There was an error rendering the completion: %s", err),
			}
		}

		s = strings.TrimRightFunc(s, unicode.IsSpace)
		s = strings.ReplaceAll(s, "\t", strings.Repeat(" ", tabWidth))
		s += "\n"

		return renderedYamlMsg(s)
	}
}

// if the input is whitespace only, make it empty.
func removeWhitespace(s string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	return s
}
