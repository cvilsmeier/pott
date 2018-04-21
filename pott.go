package pott

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
)

type Stat struct {
	TotalLines int
	UsedLines  int
}

type RestoreFunc func(silo, key string, jsonData []byte) error

type Pott interface {
	Save(silo, key string, x interface{}) error
	Stat() (*Stat, error)
	Compact() error
	CompactTo(filename string) error
}

func Open(filename string, restoreFn RestoreFunc) (Pott, error) {
	if filename == "" {
		db := &memPott{}
		return db, nil
	}	
	fd, err := loadFileData(filename)
	if err != nil {
		return nil, err
	}
	if restoreFn != nil {
		for silo, datas := range fd.silos {
			for key, data := range datas {
				err := restoreFn(silo, key, []byte(data))
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return &filePott{
		filename: filename,
	}, nil
}

// ------------------------------------------

type memPott struct {
}

func (this *memPott) Save(silo, key string, x interface{}) error {
	verifySiloAndKey(silo, key)
	_, err := marshal(x)
	if err != nil {
		return err
	}
	return nil
}

func (this *memPott) Stat() (*Stat, error) {
	return &Stat{}, nil
}

func (this *memPott) Compact() error {
	return nil
}

func (this *memPott) CompactTo(filename string) error {
	return nil
}

// ------------------------------------------

type filePott struct {
	mu   sync.Mutex
	filename string
	previousErr error
}

func (this *filePott) Save(silo, key string, x interface{}) error {
	verifySiloAndKey(silo, key)
	data, err := marshal(x)
	if err != nil {
		return err
	}
	// build line to append
	line := buildLine(silo, key, data)
	// append line to file
	this.mu.Lock()
	defer this.mu.Unlock()
	if this.previousErr != nil {
		return fmt.Errorf("cannot save, have previous error: %s", this.previousErr)
	}
	file, err := os.OpenFile(this.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write([]byte(line))
	if err != nil {
		this.previousErr = err
		return err
	}
	return nil
}

func (this *filePott) Stat() (*Stat, error) {
	this.mu.Lock()
	defer this.mu.Unlock()
	fd, err := loadFileData(this.filename)
	if err != nil {
		return nil, err
	}
	usedLines := 0
	for _, datas := range fd.silos {
		usedLines += len(datas)
	}
	return &Stat{
		TotalLines: fd.lineCount,
		UsedLines:  usedLines,
	}, nil
}

func (this *filePott) Compact() error {
	return this.CompactTo(this.filename)
}

func (this *filePott) CompactTo(destname string) error {
	this.mu.Lock()
	defer this.mu.Unlock()
	// load file
	fd, err := loadFileData(this.filename)
	if err != nil {
		return err
	}
	// open dest file
	dest, err := os.Create(destname)
	if err != nil {
		return err
	}
	defer dest.Close()
	writer := bufio.NewWriter(dest)
	defer writer.Flush()
	// sort silos
	silos := make([]string, 0, len(fd.silos))
	for silo, _ := range fd.silos {
		silos = append(silos, silo)
	}
	sort.Strings(silos)
	for _, silo := range silos {
		datas := fd.silos[silo]
		// sort keys
		keys := make([]string, 0, len(datas))
		for key, _ := range datas {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			data := datas[key]
			line := buildLine(silo, key, data)
			_, err = writer.WriteString(line)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// -------------------------------------------------

type fileData struct {
	lineCount int
	silos     map[string]map[string]string
}

func loadFileData(filename string) (*fileData, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// file does not exist
		return &fileData{}, nil
	}
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	silos := map[string]map[string]string{}
	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
		text := scanner.Text()
		var siloAndKey string
		var data string
		i := strings.IndexByte(text, '{')
		if i > 0 {
			// silo key {}
			siloAndKey = text[:i-1]
			data = text[i:]
		} else {
			// silo key
			siloAndKey = text
		}
		i = strings.IndexByte(siloAndKey, ' ')
		var silo string
		var key string
		if i > 0 {
			silo = siloAndKey[:i]
			key = siloAndKey[i+1:]
		} else {
			return nil, fmt.Errorf("in line %d: cannot parse %q into silo and key", lineCount, siloAndKey)
		}
		datas := silos[silo]
		if data == "" {
			if datas != nil {
				delete(datas, key)
			}
		} else {
			if datas == nil {
				datas = map[string]string{}
				silos[silo] = datas
			}
			datas[key] = data
		}
	}
	err = scanner.Err()
	if err != nil {
		return nil, err
	}
	return &fileData{lineCount, silos}, nil
}

func verifySiloAndKey(silo, key string) {
	if silo == "" {
		panic("empty silo")
	}
	if strings.ContainsAny(silo, " {}\r\n\t\b") {
		panic("silo " + silo + "contains invalid characters")
	}
	if key == "" {
		panic("empty key")
	}
	if strings.ContainsAny(key, " {}\r\n\t\b") {
		panic("key " + key + "contains invalid characters")
	}
}

func marshal(x interface{}) (string, error) {
	if x == nil {
		return "", nil
	}
	buf, err := json.Marshal(x)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func buildLine(silo, key, data string) string {
	if data == "" {
		return fmt.Sprintf("%s %s\n", silo, key)
	}
	return fmt.Sprintf("%s %s %s\n", silo, key, data)
}
