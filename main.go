package main

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/textarea"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dop251/goja"
)

type errMsg error

type jsReturnMsg struct {
	err string
}

type model struct {
	textarea  textarea.Model
	jsOutput  string
	status    string
	runningJs bool
	err       error
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
		console := vm.NewObject()
		console.Set("log", func(args ...interface{}) {
			m.jsOutput = fmt.Sprintln(args...)
		})
		vm.Set("console", console)
		_, err := vm.RunString(m.textarea.Value())
		if err != nil {
			return jsReturnMsg{err: err.Error()}
		}
		return jsReturnMsg{err: ""}
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
			// fmt.Println(m.jsOutput) //xxx
			cmds = append(cmds, cmd)
			m.runningJs = true
		default:
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}
	case jsReturnMsg:
		m.runningJs = false
		if msg.err != "" {
			// xxx this should show up in a status line or output window in red?
			m.status = msg.err
		} else {
			// xxx this should show up in output window
			m.status = m.jsOutput
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
		m.status,
	) + "\n\n"
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
