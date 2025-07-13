package main

import (
	"fmt"
	"log"
	"os"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/bubbles/textarea"

	tea "github.com/charmbracelet/bubbletea"
)

type CLI struct {
	Fullscreen bool   `help:"Run fullscreen"`
	Filename   string `short:"f" help:"Javascript file to run"`
}

type errMsg error

type model struct {
	textarea   textarea.Model
	statusLine string
	running    bool
	err        error
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
	case jsReturnMsg:
		m.running = false
		if msg.err != "" {
			// xxx this should show up in a status line or output window in red?
			m.statusLine = msg.err
		} else {
			// xxx this should show up in output window
			m.statusLine = msg.output
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
	return fmt.Sprintf("\n%s\n\n%s %s",
		m.textarea.View(),
		"(ctrl+g to run, ctrl+c to quit)",
		m.statusLine,
	) + "\n\n"
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
	if cli.Fullscreen {
		p = tea.NewProgram(initialModel(), tea.WithAltScreen())
	} else {
		p = tea.NewProgram(initialModel())
	}
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
