package main

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/textarea"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dop251/goja"
)

type errMsg error

type console struct {
	output string
}

func (c *console) Log(messages ...string) {
	for _, msg := range messages {
		c.output += fmt.Sprintf("%s\n", msg)
	}
}

type jsReturnMsg struct {
	err    string
	output string
}

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

func (m *model) runJS() tea.Cmd {
	return func() tea.Msg {
		vm := goja.New()
		c := &console{}

		consoleObj := vm.NewObject()
		consoleObj.Set("log", func(args ...interface{}) {
			messages := make([]string, len(args))
			for i, arg := range args {
				messages[i] = fmt.Sprintf("%v", arg)
			}
			c.Log(messages...)
		})
		vm.Set("console", consoleObj)

		_, err := vm.RunString(m.textarea.Value())
		if err != nil {
			return jsReturnMsg{err: err.Error(), output: ""}
		}
		return jsReturnMsg{err: "", output: c.output}
	}
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
			cmd = m.runJS()
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
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
