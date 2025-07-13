package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dop251/goja"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <javascript-file>")
		os.Exit(1)
	}

	filename := os.Args[1]

	jsCode, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading file %s: %v", filename, err)
	}

	vm := goja.New()

	console := vm.NewObject()
	console.Set("log", func(args ...interface{}) {
		fmt.Println(args...)
	})
	vm.Set("console", console)

	_, err = vm.RunString(string(jsCode))
	if err != nil {
		log.Fatalf("Error executing JavaScript: %v", err)
	}
}
