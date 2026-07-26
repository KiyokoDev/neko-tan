package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stateMenu state = iota
	stateFormat
	stateInput
	stateDownload
	stateResult
)

type formatChoice string

const (
	formatAudio formatChoice = "Audio (mp3)"
	formatVideo formatChoice = "Video (mp4)"
)

type model struct {
	state       state
	cursor      int
	selectedFmt formatChoice
	urlInput    textinput.Model
	output      string
	exitCode    int
	errMsg      string
	destDir     string
	width       int
	height      int
	help        help.Model
	keys        keyMap
	quitting    bool
}

type keyMap struct {
	Up    key.Binding
	Down  key.Binding
	Enter key.Binding
	Back  key.Binding
	Quit  key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Back, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Up, k.Down, k.Enter, k.Back, k.Quit}}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "https://youtube.com/watch?v=..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 60

	return model{
		state:     stateMenu,
		cursor:    0,
		urlInput:  ti,
		help:      help.New(),
		keys:      keys,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

type downloadResult struct {
	output   string
	exitCode int
	errMsg   string
	destDir  string
}

func startDownload(url string, format formatChoice) tea.Cmd {
	return func() tea.Msg {
		args := []string{
			"--newline",
			"--progress",
			"--no-playlist",
		}

		if format == formatAudio {
			args = append(args, "-x", "--audio-format", "mp3")
		} else {
			args = append(args, "-f", "mp4")
		}

		args = append(args, url)

		dir, _ := os.Getwd()
		c := exec.Command("yt-dlp", args...)
		out, err := c.CombinedOutput()

		exitCode := 0
		errMsg := ""
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				exitCode = 1
			}
			errMsg = err.Error()
		}

		return downloadResult{output: string(out), exitCode: exitCode, errMsg: errMsg, destDir: dir}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width

	case downloadResult:
		m.output = msg.output
		m.exitCode = msg.exitCode
		m.errMsg = msg.errMsg
		m.destDir = msg.destDir
		m.state = stateResult
		return m, nil

	case tea.KeyMsg:
		if m.state == stateResult {
			if msg.String() == "r" {
				m.state = stateMenu
				m.cursor = 0
				m.output = ""
				m.exitCode = 0
				m.errMsg = ""
				m.destDir = ""
			} else if msg.String() == "q" || msg.String() == "esc" || msg.String() == "enter" {
				m.quitting = true
				return m, tea.Quit
			}
			return m, nil
		}

		if m.state == stateInput {
			if msg.String() == "enter" {
				url := strings.TrimSpace(m.urlInput.Value())
				if url == "" {
					return m, nil
				}
				m.state = stateDownload
				return m, startDownload(url, m.selectedFmt)
			}
			if msg.String() == "esc" {
				m.state = stateFormat
				m.cursor = 0
				return m, nil
			}
			var cmd tea.Cmd
			m.urlInput, cmd = m.urlInput.Update(msg)
			return m, cmd
		}

		if m.state == stateDownload {
			return m, nil
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Back):
			if m.state == stateFormat {
				m.state = stateMenu
				m.cursor = 0
			}
			return m, nil

		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}

		case key.Matches(msg, m.keys.Down):
			opts := m.currentOptions()
			if m.cursor < len(opts)-1 {
				m.cursor++
			}

		case key.Matches(msg, m.keys.Enter):
			opts := m.currentOptions()
			if m.cursor < len(opts) {
				label := opts[m.cursor]
				switch m.state {
				case stateMenu:
					m.state = stateFormat
					m.cursor = 0
				case stateFormat:
					m.selectedFmt = formatChoice(label)
					m.urlInput.SetValue("")
					m.state = stateInput
					return m, textinput.Blink
				}
			}
		}
	}

	return m, nil
}

func (m model) currentOptions() []string {
	switch m.state {
	case stateMenu:
		return []string{"yutoob"}
	case stateFormat:
		return []string{string(formatAudio), string(formatVideo)}
	}
	return nil
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#f78c6c")).
			Padding(0, 1)

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f78c6c")).
			Bold(true)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#c3e88d")).
			Bold(true)

	optionStyle = lipgloss.NewStyle().
			Padding(0, 1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#546e7a"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#c3e88d")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff5370")).
			Bold(true)

	outputStyle = lipgloss.NewStyle().
			Padding(0, 2)

	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89ddff"))

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#89ddff")).
			Padding(1, 2)

	progressStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#c792ea")).
			Bold(true)

)

func (m model) View() string {
	switch m.state {
	case stateResult:
		return m.resultView()
	case stateDownload:
		return m.downloadView()
	case stateInput:
		return m.inputView()
	default:
		return m.menuView()
	}
}

func (m model) menuView() string {
	opts := m.currentOptions()
	title := "yutoob"

	var header, subtitle string
	switch m.state {
	case stateMenu:
		header = titleStyle.Render("≽^•⩊•^≼")
		subtitle = subtitleStyle.Render("Select an option:")
	case stateFormat:
		header = titleStyle.Render(fmt.Sprintf("≽^•⩊•^≼  /  %s", title))
		subtitle = subtitleStyle.Render("Choose format:")
	}

	var items []string
	for i, opt := range opts {
		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("▸")
		}
		label := optionStyle.Render(opt)
		if i == m.cursor {
			label = selectedStyle.Render(opt)
		}
		items = append(items, fmt.Sprintf("%s %s", cursor, label))
	}

	helpView := m.help.View(m.keys)

	content := lipgloss.JoinVertical(lipgloss.Center,
		header,
		"",
		subtitle,
		"",
		lipgloss.JoinVertical(lipgloss.Center, items...),
		"",
		helpView,
	)

	s := boxStyle.Render(content)
	if m.width > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, s)
	}
	return s
}

func (m model) inputView() string {
	content := lipgloss.JoinVertical(lipgloss.Center,
		titleStyle.Render("≽^•⩊•^≼  /  yutoob  /  "+string(m.selectedFmt)),
		"",
		promptStyle.Render("Paste YouTube URL:"),
		"",
		m.urlInput.View(),
		"",
		subtitleStyle.Render("enter to download · esc to go back"),
	)

	s := boxStyle.Render(content)
	if m.width > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, s)
	}
	return s
}

func (m model) downloadView() string {
	content := lipgloss.JoinVertical(lipgloss.Center,
		titleStyle.Render("≽^•⩊•^≼  /  yutoob"),
		"",
		progressStyle.Render("Downloading..."),
		"",
		subtitleStyle.Render("Please wait while yt-dlp processes your request"),
	)

	s := boxStyle.Render(content)
	if m.width > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, s)
	}
	return s
}

func (m model) resultView() string {
	var status string
	statusStyle := successStyle
	if m.exitCode != 0 {
		statusStyle = errorStyle
		status = fmt.Sprintf("✗ Failed (exit code: %d)", m.exitCode)
		if m.errMsg != "" {
			status += "\n" + m.errMsg
		}
	} else {
		status = "✓ Download complete!"
	}

	content := lipgloss.JoinVertical(lipgloss.Center,
		titleStyle.Render("≽^•⩊•^≼  /  yutoob"),
		statusStyle.Render(status),
		"",
		subtitleStyle.Render(fmt.Sprintf("Destination: %s", m.destDir)),
		"",
		outputStyle.Render(m.output),
		"",
		subtitleStyle.Render("Press [r] to download again · [enter/q/esc] to quit"),
	)

	s := boxStyle.Render(content)
	if m.width > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, s)
	}
	return s
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
