package main

import (
	"fmt"
	"log"
	"os"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
)

type CLI struct {
	Filename string `short:"f" help:"Javascript file to run"`
}

type errMsg error

type focusArea int

const (
	left focusArea = iota
	right
)

type model struct {
	width    int
	height   int
	textarea textarea.Model
	viewport viewport.Model
	focus    focusArea
	running  bool
	err      string
}

func initialModel() model {
	ta := textarea.New()
	ta.Focus()
	vp := viewport.New(0, 0)
	return model{textarea: ta, viewport: vp, focus: left}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if m.textarea.Focused() {
				m.textarea.Blur()
			}
		case tea.KeyShiftTab:
			if m.focus == left {
				m.focus = right
				m.textarea.Blur()
			} else {
				m.focus = left
				m.textarea.Focus()
			}
			return m, nil
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyCtrlG:
			m.err = ""
			cmd = runJs(m.textarea.Value())
			cmds = append(cmds, cmd)
			m.running = true
		default:
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		halfWidth := msg.Width / 2
		usableHeight := msg.Height - 4
		m.textarea.SetWidth(halfWidth - 4)
		m.textarea.SetHeight(usableHeight - 4)
		m.viewport.Width = msg.Width - halfWidth - 8
		m.viewport.Height = usableHeight - 8
	case jsReturnMsg:
		m.running = false
		if msg.err != "" {
			m.viewport.SetContent("")
			m.err = msg.err
		} else {
			m.viewport.SetContent(msg.output)
		}
	case errMsg:
		m.err = msg.Error()
		return m, nil
	}

	// Update only the focused component
	switch m.focus {
	case left:
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)
	case right:
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	headingStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		PaddingTop(1).
		PaddingLeft(2)
	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#FF5F87")).
		Padding(0, 1).
		MarginRight(1)
	leftStyle := lipgloss.NewStyle().
		Padding(2)
	rightStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#4C4E52")).
		Padding(4)
	errorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#9E3BEC")).
		Padding(4)

	left := leftStyle.Render(m.textarea.View())
	var right string
	if m.err != "" {
		right = errorStyle.Render(m.err)
	} else {
		right = rightStyle.Render(m.viewport.View())
	}
	top := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headingStyle.Render("JSizzle"),
		top,
		instructionStyle.Render("ctrl+g to run // shift+tab to switch sides // ctrl+c to quit"),
	)
}

func main() {
	var cli CLI
	kong.Parse(&cli)

	if cli.Filename != "" {
		src, err := readJSFile(cli.Filename)
		if err != nil {
			log.Fatalf("couldn't open file %s", cli.Filename)
		}
		cmd := runJs(src)
		ret := cmd().(jsReturnMsg) // a bit hacky, but it's a bubbletea command so...
		if ret.err != "" {
			log.Fatal(ret.err)
		}
		fmt.Print(ret.output)
		os.Exit(0)
	}

	var p *tea.Program
	p = tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
