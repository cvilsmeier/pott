
# Pott DB

Pott is an append-only JSON data store in pure Go.  It's well suited for
apps that hold all data in memory and have fewer than 1 millon datasets.

It stores data in a text file that is appended line-by-line:

    todos 0 {"id":"0","text":"Fix Documentation"}
    todos 1 {"id":"1","text":"Prepare README file"}
    todos 2 {"id":"2","text":"Build Binary"}
    todos 3 {"id":"3","text":"Ship"}
    todos 2 {"id":"2","text":"Build Release Binary"}
    todos 2

The first column is the 'silo' ('table' in SQL), indicating the type of the
encoded JSON. The second column is a unique key for that data line. The third
column is the data object, encoded as JSON. if the third column is empty, 
that object was deleted. In the sample above, todo 2 was deleted.

Each time an object is inserted, updated or deleted, pott will append a line to
its file. After a while, unused lines will accumulate. Pott provides a command
line tool to report file statistics (how many lines in total, how many lines
are actually used), and to compact a pott file. Compaction removes all unused
lines from the file.

A sample command line app (pott-todo) is included to demonstrate the pott API.

Pott is durable, as it writes to a file. However: If the process of writing to
the file is unexpectedly interrupted (e.g. power loss), the file may get
corrupted and must manually be fixed again.

A Pott instance is safe to be used concurrently by multiple goroutines.
However, if two pott instances write to the same file, file corruptions may
occur. Pott does not use file-locking mechanisms, as RMDBS systems do.



## Usage

In pott, every object is restored at startup time. A restore function must be
supplied that unmarshals JSON strings into domain objects, as shown here:

```go
import (
    "encoding/json"

    "pott"
)

type Todo struct {
    Id string
    Text string
}

func main() {
    todos := []Todo{}
    // restore todos from file "my_data"
    restoreFn := func(silo, key string, jsonData []byte) error {
        if silo == "todos" {
            // unmarshal Todo JSON
            x := Todo{}
            err := json.Unmarshal(jsonData, &x)
            if err != nil {
                return err
            }
            todos = append(todos, x)
            return nil
        }
        return nil // silently ignore all unknown silos
    }
    db, err := pott.Open("my_data", restoreFn)
    check(err)
    
    // insert a todo
    todo := Todo{
        Id: uuid(),
        Text: "Some new Todo item",
    }
    err := db.Save("todos", todo.Id, todo)
    check(err)

    // delete a todo by updating it to 'nil'
    firstId := todos[0].Id
    err := db.Save("todos", firstId, nil)
    check(err)
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


