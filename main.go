package main

import (
	"bufio"
	"errors"
	"flag"
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

var (
	graph     *t.Template
	abLogPath string
	psLogPath string
)

type (
	item map[string]string

	dataItem map[string]float64

	abDataSet map[string]dataItem

	psDataSet map[string]dataItem

	recoverFunc func()

	dt struct {
		Ab []dataItem
		Ps []dataItem
	}
)

func init() {
	defer func() { exitIf(recover()) }()

	flag.Usage = func() {
		fmt.Printf("Usage: %s -ab=path_to_ab.log -ps=path_to_ps.log\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.StringVar(&abLogPath, "ab", "", "Specifies a path to the ApacheBench log")
	flag.StringVar(&psLogPath, "ps", "", "Specifies a path to the ps log")

	flag.Parse()

	// check required flags count
	if flag.NFlag() == 0 {
		flag.Usage()
		panic("at least one flag should be specified.")
	}

	// some template caching there
	graph = t.Must(t.ParseFiles("graph.thtml"))
}

func main() {
	defer func() { exitIf(recover()) }()

	// read items to draw from ab log
	abData, err := abDataFromFile(abLogPath)
	exitIf(err)

	// read items to draw from ps log
	psData, err := psDataFromFile(psLogPath)
	exitIf(err)

	// create/truncate output file
	file, err := os.Create("graph.html")
	exitIf(err)
	defer file.Close()

	// convert abData to array
	abArr, err := abData.array()
	exitIf(err)

	// convert psData to array
	psArr, err := psData.array()
	exitIf(err)

	// execute templete into the file
	err = graph.Execute(file, &dt{Ab: abArr, Ps: psArr})
	exitIf(err)
}

func psDataFromFile(path string) (result psDataSet, err error) {
	defer rescue(err)()

	// ignore empty path
	if len(path) == 0 {
		return
	}

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
	result = make(psDataSet)

	// prepare scanner for file
	line := bufio.NewScanner(file)

	// get first line
	line.Scan()

	// header initialization
	header := strings.Split(line.Text(), TAB)

	// needed fields
	fields := map[string]struct{}{
		"seconds": {},
		"rss":     {},
		"psr":     {},
		"pcpu":    {},
	}

	// scan file and fill the result
	for line.Scan() {
		record := make(item)

		// map line to item where TAB("\t") is a separator
		for i, word := range strings.Split(line.Text(), TAB) {
			key := header[i]

			if _, ok := fields[key]; ok {
				record[key] = word
			}
		}

		// push to result
		result.push(record)
	}

	return
}

func abDataFromFile(path string) (result abDataSet, err error) {
	defer rescue(err)()

	// ignore empty path
	if len(path) == 0 {
		return
	}

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
	result = make(abDataSet)

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

func (d psDataSet) push(i item) (err error) {
	defer rescue(err)()

	key, ok := i["seconds"]
	if !ok {
		return errors.New("Unexpected format, 'seconds' key not found")
	}

	psr, ok := i["psr"]
	if !ok {
		return errors.New("Unexpected format, 'psr' key not found")
	}
	psrKey := "psr_" + psr

	mem, err := strconv.ParseFloat(i["rss"], 64)
	panicIf(err)

	pcpu, err := strconv.ParseFloat(i["pcpu"], 64)
	panicIf(err)

	if _, ok := d[key]; ok {
		if _, ok := d[key][psrKey]; ok {
			d[key][psrKey] += pcpu
		} else {
			d[key][psrKey] = pcpu
		}
	} else {
		d[key] = dataItem{
			psrKey: pcpu,
			"mem":  mem,
		}
	}

	return
}

func (d psDataSet) array() (res []dataItem, err error) {
	defer rescue(err)()

	res = make([]dataItem, 0)

	for key, val := range d {
		s, err := strconv.ParseFloat(key, 64)
		panicIf(err)

		data := dataItem{"seconds": s}

		for k, v := range val {
			data[k] = v
		}

		res = append(res, data)
	}

	return
}

func (d abDataSet) push(i item) (err error) {
	defer rescue(err)()

	key := i["seconds"]

	val, err := strconv.ParseFloat(i["ttime"], 64)
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
		d[key] = dataItem{"min": val, "max": val}
	}

	return
}

func (d abDataSet) array() (res []dataItem, err error) {
	defer rescue(err)()

	res = make([]dataItem, 0)

	for key, val := range d {
		s, err := strconv.ParseFloat(key, 64)
		panicIf(err)

		res = append(res,
			dataItem{
				"seconds": s,
				"min":     val["min"],
				"max":     val["max"]})
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
