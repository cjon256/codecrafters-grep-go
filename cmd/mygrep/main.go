package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
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

	matched, err := matchLine(line, pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	if matched {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}

var (
	digits    = "0123456789"
	alpha     = "abcdefghijklmnopqrstuvwxyz"
	wordChars = alpha + strings.ToUpper(alpha) + digits + "_"
)

type matchPoint struct {
	matchChars string
	inverted   bool
}

func parsePattern(patternIn string) ([]matchPoint, error) {
	index := 0
	parseSetPattern := func(inverted bool) (matchPoint, error) {
		retval := matchPoint{}
		retval.inverted = inverted
		chars := []byte{}
		for index < len(patternIn) {
			switch patternIn[index] {
			case ']':
				retval.matchChars = string(chars[:])
				return retval, nil
			default:
				chars = append(chars, patternIn[index])
			}
			index++
		}
		return retval, errors.New("parse pattern not closed")
	}

	pattern := []matchPoint{}
	for index < len(patternIn) {
		var err error
		var p matchPoint
		switch patternIn[index] {
		case '[':
			index++
			if index < len(patternIn) && patternIn[index] == '^' {
				index++
				p, err = parseSetPattern(true)
			} else {
				p, err = parseSetPattern(false)
			}
			if err != nil {
				os.Exit(3)
			}
			pattern = append(pattern, p)
		default:
			pattern = append(pattern, matchPoint{matchChars: string(patternIn[index])})
		}
		index++
	}
	return pattern, nil
}

func matchLine(line []byte, patternIn string) (bool, error) {
	pattern, err := parsePattern(patternIn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	fmt.Fprintf(os.Stderr, "%+v\n", pattern)
	fmt.Fprintf(os.Stderr, "%s\n", line)

	ok := true
	// if pattern[:1] == "[" && pattern[len(pattern)-1:] == "]" {
	// 	if pattern[1:2] == "^" {
	// 		// wow, this is ugly, there has to be a better way...
	// 		excludedChars := []byte(pattern[2 : len(pattern)-1])
	// 		f := func(r rune) bool {
	// 			s := fmt.Sprintf("%c", r)
	// 			ok := !bytes.ContainsAny(excludedChars, s)
	// 			return ok
	// 		}
	// 		ok := bytes.ContainsFunc(line, f)
	// 		return ok, nil
	// 	} else {
	// 		pattern = pattern[1 : len(pattern)-1]
	// 	}
	// } else if pattern == "\\d" {
	// 	pattern = digits
	// } else if pattern == "\\w" {
	// 	pattern = wordChars
	// } else if utf8.RuneCountInString(pattern) != 1 {
	// 	return false, fmt.Errorf("unsupported pattern: %q", pattern)
	// }
	//
	// ok := bytes.ContainsAny(line, pattern)
	//
	return ok, nil
}
