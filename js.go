package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dop251/goja"
)

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

func readJSFile(f string) (string, error) {
	b, err := os.ReadFile(f)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func runJs(src string) tea.Cmd {
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

		_, err := vm.RunString(src)
		if err != nil {
			return jsReturnMsg{err: err.Error(), output: ""}
		}
		return jsReturnMsg{err: "", output: c.output}
	}
}
