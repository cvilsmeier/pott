package pott

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestParseEmpty(t *testing.T) {
	r := strings.NewReader("\n")
	recs, err := parse(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 0 {
		t.Fatal()
	}
}

func TestParseValid(t *testing.T) {
	txt := strings.Join([]string{
		"user 1 data1",
		"user 2 data2",
		"user 3 data3",
		"-",
		"user 1 data1b",
		"user 2",
		"-",
		"user 3 data3b",
		"-",
	}, "\n") + "\n"
	r := strings.NewReader(txt)
	records, err := parse(r)
	assertNil(t, err)
	assertEq(t, 2, len(records))
	assertEq(t, "user/1/data1b", dump(records[0]))
	assertEq(t, "user/3/data3b", dump(records[1]))
}

func TestReadNotFound(t *testing.T) {
	records, err := Read("/tmp/this-file-does-not-exist")
	assertNil(t, err)
	assertEq(t, 0, len(records))
}

func TestAppendAndRead(t *testing.T) {
	filename, err := mkTemp(t.Name())
	assertNil(t, err)
	defer os.Remove(filename)
	Append(filename, []Record{
		{"user", "1", "data1"},
		{"user", "2", "data2"},
		{"user", "3", "data3"},
	})
	Append(filename, []Record{
		{"user", "1", "data1b"},
		{"user", "2", ""},
		{"user", "3", "data3b"},
	})
	data, err := os.ReadFile(filename)
	assertNil(t, err)
	want := strings.Join([]string{
		"user 1 data1",
		"user 2 data2",
		"user 3 data3",
		"-",
		"user 1 data1b",
		"user 2",
		"user 3 data3b",
		"-",
	}, "\n") + "\n"
	assertEq(t, want, string(data))
	// read
	records, err := Read(filename)
	assertNil(t, err)
	assertEq(t, 2, len(records))
	assertEq(t, "user/1/data1b", dump(records[0]))
	assertEq(t, "user/3/data3b", dump(records[1]))
}

func BenchmarkAppend(b *testing.B) {
	for n := 0; n < b.N; n++ {
		filename, _ := mkTemp(b.Name())
		defer os.Remove(filename)
		// insert 1_000_000 users takes ~ 2835ms
		count := 1_000_000
		for i := 0; i < count; i++ {
			rec := Record{"user", "user_" + strconv.Itoa(i), "data"}
			Append(filename, []Record{rec})
		}
	}
}

func BenchmarkRead(b *testing.B) {
	filename, _ := mkTemp(b.Name())
	defer os.Remove(filename)
	// insert many users
	count := 1_000_000
	for i := 0; i < count; i++ {
		rec := Record{"user", "user_" + strconv.Itoa(i), "data"}
		Append(filename, []Record{rec})
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		// read 1_000_000 users takes ~ 855ms
		records, err := Read(filename)
		if err != nil {
			b.Fatalf("want nil but have %s", err)
		}
		if len(records) != count {
			b.Fatalf("want %v but have %v", count, len(records))
		}
	}
}

func mkTemp(name string) (string, error) {
	f, err := os.CreateTemp("", name)
	if err != nil {
		return "", err
	}
	filename := f.Name()
	f.Close()
	return filename, nil
}

func dump(r Record) string {
	return fmt.Sprintf("%s/%s/%s", r.Typ, r.Id, r.Data)
}

func assertEq(t *testing.T, want, have any) {
	if want != have {
		t.Helper()
		t.Fatalf("\nwant %#v\nhave %#v", want, have)
	}
}

func assertNil(t *testing.T, v any) {
	if v != nil {
		t.Helper()
		t.Fatalf("want nil, have %v", v)
	}
}
