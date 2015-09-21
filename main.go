package main

import (
	"bufio"
	"errors"
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
	item map[string]string

	data []item
)

func init() {
	defer func() { exitIf(recover()) }()

	// some template caching there
	graph = t.Must(t.ParseFiles("graph.thtml"))
}

func main() {
	defer func() { exitIf(recover()) }()

	// read items to draw from file
	items, err := itemsFromFile("./data.txt")
	exitIf(err)

	// create/truncate output file
	file, err := os.Create("graph.html")
	defer file.Close()
	exitIf(err)

	// execute templete into the file
	err = graph.Execute(file, items)
	exitIf(err)
}

func itemsFromFile(path string) (result data, err error) {
	defer func() {
		if e := recover(); e != nil {
			if _, ok := e.(error); !ok {
				msg := fmt.Sprintf("Something went wrong: %v", e)
				err = errors.New(msg)
			} else {
				err = e.(error)
			}
		}
	}()

	// open file
	file, err := os.Open(path)
	defer file.Close()
	panicIf(err)

	// count lines in file
	fSize, err := countLinesIn(file)
	panicIf(err)

	// panic if file is empty
	if fSize <= 0 {
		msg := fmt.Sprintf("File %s is empty", file.Name())
		panicIf(errors.New(msg))
	}

	// rewind file
	_, err = file.Seek(0, 0)
	panicIf(err)

	// initialize result
	result = make(data, fSize-1)

	// prepare scanner for file
	line := bufio.NewScanner(file)

	// get first line
	line.Scan()
	// header initialization
	header := strings.Split(line.Text(), TAB)

	// line number counter
	var n int
	// scan file and fill the result
	for line.Scan() {
		record := make(item)

		// map line to item where TAB("\t") is a separator
		for i, word := range strings.Split(line.Text(), TAB) {
			record[header[i]] = word
		}

		// push to result
		result[n] = record
		// increment line number
		n++
	}

	return
}

func countLinesIn(r io.Reader) (n int, err error) {
	var count int

	// initializing read buffer
	buf := make([]byte, 8<<10) // 8Kb

	// loops forever
	for {
		// read len(buf) bytes
		count, err = r.Read(buf)

		// exit without error if end of file reached
		if err == io.EOF {
			return n, nil
		}

		// exit with error if something went wrong
		if err != nil {
			return
		}

		// counting lines
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

func panicIf(err interface{}) {
	if err != nil {
		panic(err)
	}
}
