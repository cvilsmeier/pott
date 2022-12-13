package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/cvilsmeier/pott"
)

type Todo struct {
	Id   string `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
}

func main() {
	filename := "/tmp/todos"
	log.Printf("using %s", filename)
	os.Remove(filename)
	// insert some todos
	pott.MustAppend(filename, []pott.Record{
		{Typ: "todo", Id: "0", Data: mustMarshal(Todo{"0", "Write Documentation", false})},
		{Typ: "todo", Id: "1", Data: mustMarshal(Todo{"1", "Ship Binary", false})},
		{Typ: "todo", Id: "2", Data: mustMarshal(Todo{"2", "Install", false})},
	})
	// update todo 0 and delete todo 2
	pott.MustAppend(filename, []pott.Record{
		{Typ: "todo", Id: "0", Data: mustMarshal(Todo{"0", "Write Documentation", true})},
		{Typ: "todo", Id: "2", Data: ""},
	})
	// read todos
	records := pott.MustRead(filename)
	for _, rec := range records {
		if rec.Typ == "todo" {
			var todo Todo
			mustUnmarshal(rec.Data, &todo)
			if todo.Done {
				log.Printf("Done: %s", todo.Text)
			} else {
				log.Printf("Todo: %s", todo.Text)
			}
		}
	}
}

func mustMarshal(todo Todo) string {
	data, err := json.Marshal(todo)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func mustUnmarshal(data string, v any) {
	err := json.Unmarshal([]byte(data), v)
	if err != nil {
		panic(err)
	}
}
