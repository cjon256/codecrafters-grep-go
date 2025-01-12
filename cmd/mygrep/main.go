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
	regex, err := parsePattern(pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	fmt.Fprintf(os.Stderr, "regex = '%+v'\n", regex)
	fmt.Fprintf(os.Stderr, "line = '%s'\n", line)

	matched, err := matchLine(line, regex)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	if matched {
		fmt.Fprintln(os.Stderr, "whole matched")
		os.Exit(0)
	} else {
		fmt.Fprintln(os.Stderr, "whole fails")
		os.Exit(1)
	}
}

var (
	digits    = "0123456789"
	alpha     = "abcdefghijklmnopqrstuvwxyz"
	wordChars = alpha + strings.ToUpper(alpha) + digits + "_"
)

type regExp struct {
	pattern []matchPoint
}

type matchPoint struct {
	matchChars string
	inverted   bool
}

func (re regExp) matchHere(line []byte, current int) bool {
	for i := 0; i < len(re.pattern); i++ {
		if current >= len(line) {
			fmt.Fprintln(os.Stderr, "oops, got to long")
			return false
		}
		if !re.pattern[i].matchOnce(line, current) {
			fmt.Fprintln(os.Stderr, "match fails in regExp::matchHere")
			return false
		}
		current++
	}
	return true
}

func (mp *matchPoint) matchOnce(line []byte, current int) bool {
	matches := strings.Contains(mp.matchChars, string((line)[current]))
	fmt.Fprintf(os.Stderr, "matchHere(line='%s', current=%d) with mp = '%v+'\n", string(line), current, mp)
	if mp.inverted {
		matches = !matches
	}
	if matches {
		// fmt.Fprintln("matches")
		return true
	} else {
		// fmt.Fprintln("fails")
		return false
	}
}

func parsePattern(patternIn string) (regExp, error) {
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
		case '\\':
			index++
			if index < len(patternIn) {
				switch patternIn[index] {
				case 'w':
					pattern = append(pattern, matchPoint{matchChars: wordChars})
				case 'd':
					pattern = append(pattern, matchPoint{matchChars: digits})
				default:
					pattern = append(pattern, matchPoint{matchChars: string(patternIn[index])})
				}
			} else {
				// last character was a backslash....
				// I guess append a backslash character?
				pattern = append(pattern, matchPoint{matchChars: "\\"})
			}
		default:
			pattern = append(pattern, matchPoint{matchChars: string(patternIn[index])})
		}
		index++
	}
	return regExp{pattern}, nil
}

func matchLine(line []byte, pattern regExp) (bool, error) {
	for current := 0; current < len(line); current++ {
		matchesHere := pattern.matchHere(line, current)
		if matchesHere {
			fmt.Fprintln(os.Stderr, "ml whole matched")
			return true, nil
		}
	}

	fmt.Fprintln(os.Stderr, "ml whole fails")
	return false, nil
}
