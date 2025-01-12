package main

import (
	"fmt"
	"internal/myregexp"
	"io"
	"os"
)

// Usage: echo <input_text> | your_program.sh -E <pattern>
func main() {
	if len(os.Args) < 3 || os.Args[1] != "-E" {
		fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern>\n")
		os.Exit(2) // 1 means no lines were selected, >1 means error
	}

	pattern := os.Args[2]
	regex, err := myregexp.ParsePattern(pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	fmt.Fprintf(os.Stderr, "regex = '%+v'\n", regex)

	// XXX ReadAll assumes we're only dealing with a single line
	line, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: read input text: %v\n", err)
		os.Exit(2)
	}
	fmt.Fprintf(os.Stderr, "line = '%s'\n", line)

	matched, err := regex.MatchLine(line)
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
