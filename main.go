package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
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

	dataSet map[string]map[string]int

	recoverFunc func()
)

func init() {
	defer func() { exitIf(recover()) }()

	// some template caching there
	graph = t.Must(t.ParseFiles("graph.thtml"))
}

func main() {
	defer func() { exitIf(recover()) }()

	// read items to draw from file
	data, err := itemsFromFile("./data.txt")
	exitIf(err)

	// create/truncate output file
	file, err := os.Create("graph.html")
	exitIf(err)
	defer file.Close()

	// execute templete into the file
	err = graph.Execute(file, data)
	exitIf(err)
}

func itemsFromFile(path string) (result dataSet, err error) {
	defer rescue(err)

	// open file
	file, err := os.Open(path)
	panicIf(err)
	defer file.Close()

	fileStat, err := file.Stat()
	panicIf(err)

	// panic if file is empty
	if fileStat.Size() <= 0 {
		msg := fmt.Sprintf("File %s is empty", file.Name())
		panic(errors.New(msg))
	}

	// initialize result
	result = make(dataSet)

	// prepare scanner for file
	line := bufio.NewScanner(file)

	// get first line
	line.Scan()
	// header initialization
	header := strings.Split(line.Text(), TAB)

	// scan file and fill the result
	for line.Scan() {
		record := make(item)

		// map line to item where TAB("\t") is a separator
		for i, word := range strings.Split(line.Text(), TAB) {
			key := header[i]
			if key == "seconds" || key == "ttime" {
				record[key] = word
			}
		}

		// push to result
		result.push(record)
	}

	return
}

func (d dataSet) push(i item) (err error) {
	defer rescue(err)

	key := i["seconds"]

	val, err := strconv.Atoi(i["ttime"])
	panicIf(err)

	if _, ok := d[key]; ok {
		if min, ok := d[key]["min"]; ok {
			if min > val {
				d[key]["min"] = val
			}
		} else {
			d[key]["min"] = val
		}

		if max, ok := d[key]["max"]; ok {
			if max < val {
				d[key]["max"] = val
			}
		} else {
			d[key]["max"] = val
		}
	} else {
		d[key] = map[string]int{"min": val, "max": val}
	}

	return
}

func rescue(err error) recoverFunc {
	return func() {
		if e := recover(); e != nil {
			if _, ok := e.(error); !ok {
				msg := fmt.Sprintf("Something went wrong: %v", e)
				err = errors.New(msg)
			} else {
				err = e.(error)
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
