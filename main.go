package main

import (
	"fmt"
	"os"

	t "html/template"
)

var graph *t.Template

func init() {
	defer func() { ExitIf(recover()) }()

	// some template caching there
	graph = t.Must(t.ParseFiles("graph.thtml"))
}

func main() {
	defer func() { ExitIf(recover()) }()

	file, err := os.Create("graph.html")
	defer file.Close()
	ExitIf(err)

	err = graph.Execute(file, "gopher")
	ExitIf(err)
}

func ExitIf(err interface{}) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
}
