package main

import (
	"fmt"
	"log"
	"os"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
)

type CLI struct {
	Filename string `short:"f" help:"Javascript file to run"`
}

type errMsg error

type model struct {
	width    int
	height   int
	textarea textarea.Model
	output   string
	running  bool
	err      error
}

func initialModel() model {
	ti := textarea.New()
	ti.Focus()
	return model{textarea: ti, err: nil}
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
		m.textarea.SetWidth(msg.Width / 2)
		m.textarea.SetHeight(msg.Height - 2)
	case jsReturnMsg:
		m.running = false
		if msg.err != "" {
			m.output = msg.err
		} else {
			m.output = msg.output
		}
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	left := lipgloss.NewStyle().Width(m.width / 2).Render(m.textarea.View())
	right := lipgloss.NewStyle().Width(m.width - m.width/2).Render(m.output)
	top := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	instructions := "(ctrl+g to run, ctrl+c to quit)"
	return lipgloss.JoinVertical(lipgloss.Left, top, instructions)
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
