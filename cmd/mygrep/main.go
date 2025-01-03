package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"unicode/utf8"
)

// Ensures gofmt doesn't remove the "bytes" import above (feel free to remove this!)
var _ = bytes.ContainsAny

// Usage: echo <input_text> | your_program.sh -E <pattern>
func main() {
	if len(os.Args) < 3 || os.Args[1] != "-E" {
		fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern>\n")
		os.Exit(2) // 1 means no lines were selected, >1 means error
	}

	pattern := os.Args[2]

	line, err := io.ReadAll(os.Stdin) // assume we're only dealing with a single line
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: read input text: %v\n", err)
		os.Exit(2)
	}

	ok, err := matchLine(line, pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	if !ok {
		os.Exit(1)
	}

	// default exit code is 0 which means success
}

func matchLine(line []byte, pattern string) (bool, error) {
	if pattern[:1] == "[" && pattern[len(pattern)-1:] == "]" {
		if pattern[1:2] == "^" {
			// wow, this is ugly, there has to be a better way...
			excludedChars := []byte(pattern[2 : len(pattern)-1])
			f := func(r rune) bool {
				s := fmt.Sprintf("%c", r)
				ok := !bytes.ContainsAny(excludedChars, s)
				return ok
			}
			ok := bytes.ContainsFunc(line, f)
			return ok, nil
		} else {
			pattern = pattern[1 : len(pattern)-1]
		}
	} else if pattern == "\\d" {
		pattern = "0123456789"
	} else if pattern == "\\w" {
		pattern = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890_"
	} else if utf8.RuneCountInString(pattern) != 1 {
		return false, fmt.Errorf("unsupported pattern: %q", pattern)
	}

	ok := bytes.ContainsAny(line, pattern)

	return ok, nil
}
