package pott

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestTodo struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func TestPottSave(t *testing.T) {
	// tmp
	filename := "testdata/TestPottSave.tmp"
	os.Remove(filename)
	defer os.Remove(filename)
	// open db
	db, err := Open(filename, nil)
	assert.Nil(t, err)
	assert.NotNil(t, db)
	// append some records
	db.Save("todos", "a", TestTodo{"a", "Alice"})
	db.Save("todos", "b", TestTodo{"b", "Bob"})
	db.Save("todos", "a", TestTodo{"a", "Alice2"})
	db.Save("todos", "b", nil)
	// check file content
	buffer, _ := ioutil.ReadFile(filename)
	expected := ""
	expected += "todos a {\"id\":\"a\",\"name\":\"Alice\"}\n"
	expected += "todos b {\"id\":\"b\",\"name\":\"Bob\"}\n"
	expected += "todos a {\"id\":\"a\",\"name\":\"Alice2\"}\n"
	expected += "todos b\n"
	assert.Equal(t, expected, string(buffer))
}

func TestPottRestore(t *testing.T) {
	// tmp
	filename := "testdata/TestPottRestore.tmp"
	os.Remove(filename)
	defer os.Remove(filename)
	// open db
	{
		db, err := Open(filename, nil)
		assert.Nil(t, err)
		assert.NotNil(t, db)
		// append some records
		db.Save("todos", "a", TestTodo{"a", "Alice"})
		db.Save("todos", "b", TestTodo{"b", "Bob"})
		db.Save("todos", "a", TestTodo{"a", "Alice2"})
		db.Save("todos", "c", TestTodo{"c", "Carol"})
		db.Save("todos", "b", nil)
	}
	// re-open db
	{
		todos := []TestTodo{}
		restore := func(silo, key string, jsonData []byte) error {
			assert.Equal(t, "todos", silo)
			todo := TestTodo{}
			err := json.Unmarshal(jsonData, &todo)
			assert.Nil(t, err)
			todos = append(todos, todo)
			return nil
		}
		db, err := Open(filename, restore)
		assert.Nil(t, err)
		assert.NotNil(t, db)
		assert.Equal(t, 2, len(todos))
		// sort them
		sort.Slice(todos, func(i, j int) bool {
			return todos[i].Id < todos[j].Id
		})
		assert.Equal(t, "a", todos[0].Id)
		assert.Equal(t, "Alice2", todos[0].Name)
		assert.Equal(t, "c", todos[1].Id)
		assert.Equal(t, "Carol", todos[1].Name)
	}
}

func TestPottStat(t *testing.T) {
	// tmp
	filename := "testdata/TestPottStat.tmp"
	os.Remove(filename)
	defer os.Remove(filename)
	// open db
	{
		db, err := Open(filename, nil)
		assert.Nil(t, err)
		assert.NotNil(t, db)
		// stat
		stat, err := db.Stat()
		assert.Nil(t, err)
		assert.Equal(t, 0, stat.TotalLines)
		assert.Equal(t, 0, stat.UsedLines)
		// append some records
		db.Save("todos", "a", TestTodo{"a", "A"})
		db.Save("todos", "b", TestTodo{"b", "B"})
		db.Save("todos", "a", TestTodo{"a", "A2"})
		db.Save("todos", "c", TestTodo{"c", "C"})
		db.Save("todos", "b", nil)
		// stat
		stat, err = db.Stat()
		assert.Nil(t, err)
		assert.Equal(t, 5, stat.TotalLines)
		assert.Equal(t, 2, stat.UsedLines)
		// append some more
		db.Save("todos", "a", nil)
		db.Save("todos", "b", TestTodo{"b", "B"})
		// stat
		stat, err = db.Stat()
		assert.Nil(t, err)
		assert.Equal(t, 7, stat.TotalLines)
		assert.Equal(t, 2, stat.UsedLines)
	}
	// re-open db
	{
		db, err := Open(filename, nil)
		assert.Nil(t, err)
		assert.NotNil(t, db)
		// stat
		stat, err := db.Stat()
		assert.Nil(t, err)
		assert.Equal(t, 7, stat.TotalLines)
		assert.Equal(t, 2, stat.UsedLines)
	}
}

func TestPottInMemory(t *testing.T) {
	// open
	db, _ := Open("", nil)
	// stat
	stat, _ := db.Stat()
	assert.Equal(t, 0, stat.TotalLines)
	assert.Equal(t, 0, stat.UsedLines)
	db.Save("todos", "a", TestTodo{Id: "a", Name: "Alice"})
	db.Save("todos", "b", TestTodo{Id: "b", Name: "Bob"})
	db.Save("todos", "c", TestTodo{Id: "c", Name: "Carol"})
	db.Save("todos", "a", TestTodo{Id: "a", Name: "Alice X."})
	db.Save("todos", "a", TestTodo{Id: "a", Name: "Alice Y."})
	db.Save("todos", "a", TestTodo{Id: "a", Name: "Alice Z."})
	db.Save("todos", "b", nil)
	// stat
	stat, _ = db.Stat()
	assert.Equal(t, 0, stat.TotalLines)
	assert.Equal(t, 0, stat.UsedLines)
	// compact
	err := db.Compact()
	assert.Nil(t, err)
	// stat again
	stat, _ = db.Stat()
	assert.Equal(t, 0, stat.TotalLines)
	assert.Equal(t, 0, stat.UsedLines)
}

func TestPottCompact(t *testing.T) {
	// tmp
	filename := "testdata/TestPottCompact.tmp"
	os.Remove(filename)
	defer os.Remove(filename)
	// open db
	{
		db, err := Open(filename, nil)
		assert.Nil(t, err)
		assert.NotNil(t, db)
		db.Save("todos", "a", TestTodo{Id: "a", Name: "Alice"})
		db.Save("todos", "b", TestTodo{Id: "b", Name: "Bob"})
		db.Save("todos", "c", TestTodo{Id: "c", Name: "Carol"})
		db.Save("todos", "a", TestTodo{Id: "a", Name: "Alice X."})
		db.Save("todos", "a", TestTodo{Id: "a", Name: "Alice Y."})
		db.Save("todos", "a", TestTodo{Id: "a", Name: "Alice Z."})
		db.Save("todos", "c", nil)
		// stat
		stat, _ := db.Stat()
		assert.Equal(t, 7, stat.TotalLines)
		assert.Equal(t, 2, stat.UsedLines)
		// compact
		err = db.Compact()
		assert.Nil(t, err)
		// stat
		stat, _ = db.Stat()
		assert.Equal(t, 2, stat.TotalLines)
		assert.Equal(t, 2, stat.UsedLines)
	}
	// open db
	{
		db, err := Open(filename, nil)
		assert.Nil(t, err)
		assert.NotNil(t, db)
		// stat
		stat, _ := db.Stat()
		assert.Equal(t, 2, stat.TotalLines)
		assert.Equal(t, 2, stat.UsedLines)
	}
}

func BenchmarkPottSaveFile(b *testing.B) {
	// tmp
	filename := "testdata/BenchmarkPottSaveFile.tmp"
	os.Remove(filename)
	defer os.Remove(filename)
	db, err := Open(filename, nil)
	assert.Nil(b, err)
	todo := TestTodo{Id: "42", Name: "abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz"}
	// start benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.Save("todos", todo.Id, todo)
	}
}

func BenchmarkPottSaveMemory(b *testing.B) {
	db, err := Open("", nil)
	assert.Nil(b, err)
	todo := TestTodo{Id: "42", Name: "abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz"}
	// start benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.Save("todos", todo.Id, todo)
	}
}

func BenchmarkPottRestoreFile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		restoreCount := 0
		restore := func(silo, id string, jsonData []byte) error {
			if silo == "todos" {
				todo := TestTodo{}
				err := json.Unmarshal(jsonData, &todo)
				if err != nil {
					return err
				}
				restoreCount++
				return nil
			}
			return fmt.Errorf("wrong silo %q", silo)
		}
		_, err := Open("testdata/todos.json", restore)
		assert.Nil(b, err)
		assert.Equal(b, 4, restoreCount)
	}
}
