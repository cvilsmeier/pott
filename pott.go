package pott

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type Record struct {
	Typ  string
	Id   string
	Data string
}

// typ => id => data
type typeIdMap map[string]map[string]string

func MustRead(filename string) []Record {
	records, err := Read(filename)
	if err != nil {
		panic(err)
	}
	return records
}

func Read(filename string) ([]Record, error) {
	f, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	return parse(f)
}

func parse(r io.Reader) ([]Record, error) {
	timap := make(typeIdMap)
	var tx []Record
	sca := bufio.NewScanner(r)
	for sca.Scan() {
		line := sca.Text()
		if line == "" {
			// ignore empty lines
		} else if line == "-" {
			// apply tx to records
			for _, rec := range tx {
				imap := timap[rec.Typ]
				if imap == nil {
					imap = make(map[string]string)
					timap[rec.Typ] = imap
				}
				if rec.Data == "" {
					// delete
					delete(imap, rec.Id)
				} else {
					// add
					imap[rec.Id] = rec.Data
				}
			}
			tx = tx[:0]
		} else {
			rec, err := parseRecord(line)
			if err != nil {
				return nil, err
			}
			tx = append(tx, rec)
		}
	}
	err := sca.Err()
	if err != nil {
		return nil, err
	}
	if len(tx) > 0 {
		return nil, fmt.Errorf("unfinished tx at end of input")
	}
	var records []Record
	for typ, imap := range timap {
		for id, data := range imap {
			records = append(records, Record{typ, id, data})
		}
	}
	sort.Slice(records, func(i, j int) bool {
		a := records[i]
		b := records[j]
		if a.Typ != b.Typ {
			return a.Typ < b.Typ
		}
		return a.Id < b.Id
	})
	return records, nil
}

func parseRecord(line string) (Record, error) {
	// line = "user 1 data"
	// line = "user 1"
	i := strings.IndexRune(line, ' ')
	if i < 0 {
		return Record{}, fmt.Errorf("invalid line: first space not found")
	}
	typ := line[:i]
	line = line[i+1:] // id data
	i = strings.IndexRune(line, ' ')
	if i < 0 {
		return Record{typ, line, ""}, nil
	}
	id := line[:i]
	line = line[i+1:] // data
	return Record{typ, id, line}, nil
}

// MustAppend is like Append except it panics on error.
func MustAppend(filename string, records []Record) {
	err := Append(filename, records)
	if err != nil {
		panic(err)
	}
}

// Append appends records to a file. Each Append invocation is a transaction.
// Append returns any error encountered. It is possible that Append appends only
// a portion of the provided records. In that case the database is corrupted.
func Append(filename string, records []Record) error {
	// print records to memory
	var buf bytes.Buffer
	var err error
	for _, rec := range records {
		if rec.Data == "" {
			_, err = fmt.Fprintf(&buf, "%s %s\n", rec.Typ, rec.Id)
		} else {
			_, err = fmt.Fprintf(&buf, "%s %s %s\n", rec.Typ, rec.Id, rec.Data)
		}
		if err != nil {
			return err
		}
	}
	_, err = buf.WriteString("-\n")
	if err != nil {
		return err
	}
	// write memory buffer to file
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(buf.Bytes())
	return err
}

func Compact(srcfile, destfile string) error {
	records, err := Read(srcfile)
	if err != nil {
		return err
	}
	return Append(destfile, records)
}
