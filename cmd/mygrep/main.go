package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/codecrafters-io/grep-starter-go/cmd/mygrep/regexp"
)

// debugf convenience function
func debugf(format string, args ...interface{}) {
	if regexp.DEBUG {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}

// Usage: echo <input_text> | your_program.sh -E <pattern>
func main() {
	if len(os.Args) < 3 || os.Args[1] != "-E" {
		fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern>\n")
		os.Exit(2) // 1 means no lines were selected, >1 means error
	}

	pattern := os.Args[2]
	regex := regexp.ParseRegExp(pattern)

	// XXX ReadAll assumes we're only dealing with a single line
	line, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: read input text: %v\n", err)
		os.Exit(2)
	}
	// trim any newline off of that in case we forget -n for echo
	line = bytes.TrimRight(line, "\n\r")
	debugf("line = '%s'\n", line)

	matched := regex.MatchLine(line)
	if matched {
		debugf("whole matched")
		os.Exit(0)
	} else {
		debugf("whole fails")
		os.Exit(1)
	}
}
