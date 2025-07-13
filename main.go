package main

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/textarea"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dop251/goja"
)

type errMsg error

type model struct {
	textarea textarea.Model
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
			return m, tea.Quit
		default:
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
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
	return fmt.Sprintf("\n%s\n\n%s",
		m.textarea.View(),
		"(ctrl+g to run, ctrl+c to quit)",
	) + "\n\n"
}

func main() {
	vm := goja.New()
	console := vm.NewObject()
	console.Set("log", func(args ...interface{}) {
		fmt.Println(args...)
	})
	vm.Set("console", console)

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func runJavascript(vm *goja.Runtime, src string) error {
	_, err := vm.RunString(src)
	if err != nil {
		return err
	}
	return nil
}
