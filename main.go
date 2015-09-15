package main

import (
	"fmt"
	"os"

	t "html/template"
)

var graph *t.Template

type data struct {
	Items []item
}

type item map[string]string

func init() {
	defer func() { exitIf(recover()) }()

	// some template caching there
	graph = t.Must(t.ParseFiles("graph.thtml"))
}

func main() {
	defer func() { exitIf(recover()) }()

	file, err := os.Create("graph.html")
	defer file.Close()
	exitIf(err)

	// data := parse("path/to/file")

	dt := &data{Items: []item{
		item{"x": "1,2", "y": "0"},
		item{"x": "2,3", "y": "1"},
		item{"x": "3,4", "y": "3"},
		item{"x": "5,5", "y": "4"}}}

	err = graph.Execute(file, dt)
	exitIf(err)
}

func exitIf(err interface{}) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
}
