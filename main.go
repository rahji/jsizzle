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
	err      error
}

func initialModel() model {
	ta := textarea.New()
	ta.Focus()
	vp := viewport.New(0, 0)
	return model{textarea: ta, viewport: vp, err: nil, focus: left}
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
		usableHeight := msg.Height - 8 // minus status line
		m.textarea.SetWidth(halfWidth)
		m.textarea.SetHeight(usableHeight)
		m.viewport.Width = msg.Width - halfWidth
		m.viewport.Height = usableHeight
		m.viewport.Width = msg.Width - halfWidth
		m.viewport.Height = usableHeight
	case jsReturnMsg:
		m.running = false
		if msg.err != "" {
			m.viewport.SetContent(msg.err)
		} else {
			m.viewport.SetContent(msg.output)
		}
	case errMsg:
		m.err = msg
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
	left := lipgloss.NewStyle().Width(m.width / 2).Render(m.textarea.View())
	right := lipgloss.NewStyle().Width(m.width - m.width/2).Render(m.viewport.View())
	top := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	heading := "\n\nJSizzle\n\n"
	instructions := "\n\n// ctrl+g to run // shift+tab to switch sides // ctrl+c to quit //\n\n"
	return lipgloss.JoinVertical(lipgloss.Left, heading, top, instructions)
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
