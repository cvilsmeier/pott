Pott
===============================================================================

[![GoDoc](https://godoc.org/github.com/cvilsmeier/pott?status.svg)](https://godoc.org/github.com/cvilsmeier/pott)

Pott is an append-only data store in Go. It's well suited for
apps that hold all data in memory and have fewer than 1 millon datasets.

It stores data in a text file that is appended line-by-line:

    todos 0 {"text":"Fix Documentation", "done":false}
    todos 1 {"text":"Prepare README file", "done":false}
    todos 2 {"text":"Ship", "done":false}
    -
    todos 0 {"text":"Fix Documentation", "done":true}
    todos 1 {"text":"Prepare README file", "done":true}
    -
    todos 0
    -

The first column is the record type.
The second column is a unique id for that record.
The third column is the data, for example a JSON string.
If the third column is empty, that record was deleted.
In the sample above, todo 0 was deleted.

Lines consisting only of `-` are transaction boundaries.

Each time a record is inserted, updated or deleted, pott will append a line to
its file. After a while, unused lines will accumulate. Pott provides a 
tool to compact a data file. Compaction removes all unused
lines from the file.

Pott files are durable. However: If the process of writing to
the file is unexpectedly interrupted (e.g. power loss), the file may get
corrupted and must manually be fixed again.

Pott does not use file-locking mechanisms, synchronization has to be done
by the caller.



## Usage

    go get github.com/cvilsmeier/pott


```go
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

```

## Author

C. Vilsmeier


## License

This is free and unencumbered software released into the public domain.

Anyone is free to copy, modify, publish, use, compile, sell, or
distribute this software, either in source code form or as a compiled
binary, for any purpose, commercial or non-commercial, and by any
means.

In jurisdictions that recognize copyright laws, the author or authors
of this software dedicate any and all copyright interest in the
software to the public domain. We make this dedication for the benefit
of the public at large and to the detriment of our heirs and
successors. We intend this dedication to be an overt act of
relinquishment in perpetuity of all present and future rights to this
software under copyright law.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS BE LIABLE FOR ANY CLAIM, DAMAGES OR
OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.

For more information, please refer to <http://unlicense.org/>
