package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"pott"
)

type Todo struct {
	Id   string `json:"id"`
	Text string `json:"text"`
}

var myFilename string = ""
var myTodos map[string]Todo = make(map[string]Todo)
var myPott pott.Pott
var myScanner *bufio.Scanner

func main() {
	fmt.Printf("-=> Pott TODO <=-\n")
	fmt.Printf("a sample app using pott\n")
	flag.Parse()
	myFilename = flag.Arg(0)
	if myFilename == "" {
		fmt.Printf("operating in memory-only mode, will not write any file")
	}
	// open pott file
	var err error
	restoreFn := func(silo, key string, jsonData []byte) error {
		fmt.Printf("restore silo=%q, key=%q, json=%s\n", silo, key, string(jsonData))
		if silo == "todos" {
			todo := Todo{}
			err := json.Unmarshal(jsonData, &todo)
			if err != nil {
				return err
			}
			myTodos[todo.Id] = todo
		}
		// silently ignore all other silos
		return nil
	}
	myPott, err = pott.Open(myFilename, restoreFn)
	die(err)
	// execute command loop
	doHelp()
	fmt.Printf("> ")
	myScanner = bufio.NewScanner(os.Stdin)
	for myScanner.Scan() {
		switch strings.TrimSpace(myScanner.Text()) {
		case "":
			// do nothing
		case "l", "list":
			doList()
		case "a", "add":
			doAdd()
		case "d", "del":
			doDel()
		case "sa", "sample":
			doSample()
		case "st", "stat":
			doStat()
		case "co", "compact":
			doCompact()
		case "?", "help":
			doHelp()
		case "q", "quit":
			fmt.Printf("bye.")
			os.Exit(0)
		default:
			fmt.Printf("what?\n")
		}
		fmt.Printf("> ")
	}
}

func openPott() {
}

func doHelp() {
	fmt.Printf("Commands:\n")
	fmt.Printf("  l, list     list all todos\n")
	fmt.Printf("  a, add      add todo\n")
	fmt.Printf("  d, del      delete todo\n")
	fmt.Printf("  sa, sample  insert sample todos\n")
	fmt.Printf("  st, stat    show pott stat\n")
	fmt.Printf("  co, compact compact pott file\n")
	fmt.Printf("  ?, help     print help\n")
	fmt.Printf("  q, quit     quit\n")
}

func doList() {
	// sort them by id
	todos := []Todo{}
	for _, todo := range myTodos {
		todos = append(todos, todo)
	}
	sort.Slice(todos, func(i, j int) bool {
		return todos[i].Id < todos[j].Id
	})
	fmt.Printf("You have %d todos\n", len(todos))
	for _, todo := range todos {
		fmt.Printf("  %s - %s\n", todo.Id, todo.Text)
	}
}

func doAdd() {
	fmt.Printf("Enter Id > ")
	if myScanner.Scan() {
		id := myScanner.Text()
		fmt.Printf("Enter Text > ")
		if myScanner.Scan() {
			text := myScanner.Text()
			todo := Todo{id, text}
			myTodos[id] = todo
			myPott.Save("todos", id, todo)
			fmt.Printf("Added Todo " + todo.Id + " - " + todo.Text + "\n")
		}
	}
}

func doDel() {
	fmt.Printf("Enter Id > ")
	if myScanner.Scan() {
		id := myScanner.Text()
		if _, found := myTodos[id]; found {
			delete(myTodos, id)
			myPott.Save("todos", id, nil)
			fmt.Printf("Deleted Todo %s\n", id)
		} else {
			fmt.Printf("Id %s not found\n", id)
		}
	}
}

func doSample() {
	// insert sample todos, with fixed ids
	for i, text := range []string{"Fix Documentation", "Prepare README file", "Build Binary", "Ship everything to customer"} {
		id := fmt.Sprintf("%d", i)
		todo := Todo{id, text}
		myTodos[id] = todo
		err := myPott.Save("todos", todo.Id, todo)
		die(err)
	}
}

func doStat() {
	stat, err := myPott.Stat()
	die(err)
	fmt.Printf("%3d lines total\n", stat.TotalLines)
	fmt.Printf("%3d lines used\n", stat.UsedLines)
}

func doCompact() {
	err := myPott.Compact()
	die(err)
	fmt.Printf("compacted\n")
}

func die(x interface{}) {
	if x != nil {
		fmt.Printf("an unrecoverable error occurred: %s", x)
		panic(x)
	}
}
