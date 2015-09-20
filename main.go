package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	t "html/template"
)

const (
	TAB = "\t"
	EOL = '\n'
)

var graph *t.Template

type (
	data struct {
		Items []item
	}

	item map[string]string
)

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

	fmt.Printf("%#v", newDataFromFile("./data.txt"))
	// err = graph.Execute(file, &data{newDataFromFile("./data.txt")})
	// exitIf(err)
}

func newDataFromFile(path string) []item {
	file, err := os.Open(path)
	exitIf(err)

	defer file.Close()

	fSize, err := countLinesIn(file)
	exitIf(err)

	_, err = file.Seek(0, 0)
	exitIf(err)

	result := make([]item, fSize-1)

	// prepare scanner for file
	line := bufio.NewScanner(file)

	// header initializtion
	line.Scan()
	header := strings.Split(line.Text(), TAB)

	var n int // line number counter
	// scan file and fill the result
	for line.Scan() {
		record := make(item)

		// map line to item where TAB("\t") is separator
		for i, word := range strings.Split(line.Text(), TAB) {
			record[header[i]] = word
		}

		result[n] = record // push to result
		n++                // increment line number
	}

	return result
}

func countLinesIn(r io.Reader) (n int, err error) {
	var count int

	buf := make([]byte, 8<<10) // 8Kb

	for {
		count, err = r.Read(buf)

		if err == io.EOF {
			return n, nil
		}

		if err != nil {
			return
		}

		for _, b := range buf[:count] {
			if b == EOL {
				n++
			}
		}
	}
}

func exitIf(err interface{}) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
}
