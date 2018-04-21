package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"pott"
)

const VERSION = "1.0.0"

const USAGE = `Pott is a tool for managing pott files.

Usage:

    pott command [arguments]

The commands are:

    stat <filename>
        Stat reports the total number of lines,
        the number of used lines, and the number of unused
        lines of a pott file.
        It is safe to run analyze against a file
        that's in use by another process.

    compact <source> <dest> 
        Compact reads the source file, removes unused data
        lines, and writes the result to the dest file.
        Compact is used to get rid of unused data lines that
        accumulate over time. Source and dest can be the same.
        It is NOT safe to run compact against a file that's
        in use by another process.

    help
        Show this help page

    version
        Show program version

`

func main() {
	flag.Parse()
	command := flag.Arg(0)
	if command == "stat" {
		doStat()
	} else if command == "compact" {
		doCompact()
	} else if command == "" || command == "help" {
		doHelp(0)
	} else if command == "version" {
		doVersion()
	} else {
		fmt.Printf("unknown command %q\n", command)
		doHelp(2)
	}
}

func doStat() {
	filename := flag.Arg(1)
	if filename == "" {
		fmt.Printf("no filename\n")
		os.Exit(2)
	}
	fmt.Printf("stat %s\n", filename)
	db, err := pott.Open(filename, nil)
	if err != nil {
		fmt.Printf("cannot open %s: %s\n", filename, err)
		os.Exit(1)
	}
	stat, err := db.Stat()
	if err != nil {
		fmt.Printf("cannot stat %s: %s\n", filename, err)
		os.Exit(1)
	}
	if stat.TotalLines == 0 {
		fmt.Printf("no lines found, file %s seems to be empty", filename)
	} else {
		totalPercent := 100
		usedPercent := (100 * stat.UsedLines) / stat.TotalLines
		unusedLines := stat.TotalLines - stat.UsedLines
		unusedPercent := totalPercent - usedPercent
		fmt.Printf("total  lines: %6d (%d%%)\n", stat.TotalLines, totalPercent)
		fmt.Printf("used   lines: %6d (%d%%)\n", stat.UsedLines, usedPercent)
		fmt.Printf("unused lines: %6d (%d%%)\n", unusedLines, unusedPercent)
	}
	os.Exit(0)
}

func doCompact() {
	infilename := flag.Arg(1)
	if infilename == "" {
		fmt.Printf("no source file\n")
		os.Exit(2)
	}
	outfilename := flag.Arg(2)
	if outfilename == "" {
		fmt.Printf("no dest file\n")
		os.Exit(2)
	}
	fmt.Printf("compact %s to %s\n", infilename, outfilename)
	db, err := pott.Open(infilename, nil)
	if err != nil {
		fmt.Printf("cannot open %s: %s\n", infilename, err)
		os.Exit(1)
	}
	err = db.CompactTo(outfilename)
	if err != nil {
		fmt.Printf("cannot compact %s: %s\n", outfilename, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func doVersion() {
	fmt.Print(VERSION)
	os.Exit(0)
}

func doHelp(exitCode int) {
	usage := strings.Replace(USAGE, "{{VERSION}}", VERSION, -1)
	fmt.Print(usage)
	os.Exit(exitCode)
}
