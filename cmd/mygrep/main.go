package main

import (
	"fmt"
	"io"
	"os"

	"github.com/codecrafters-io/grep-starter-go/cmd/mygrep/regexp"
)

// Usage: echo <input_text> | your_program.sh -E <pattern>
func main() {
	if len(os.Args) < 3 || os.Args[1] != "-E" {
		fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern>\n")
		os.Exit(2) // 1 means no lines were selected, >1 means error
	}

	pattern := os.Args[2]
	regex := regexp.ParseRegExp(pattern)
	fmt.Fprintf(os.Stderr, "%s", regex)

	// XXX ReadAll assumes we're only dealing with a single line
	line, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: read input text: %v\n", err)
		os.Exit(2)
	}

	matched := regex.MatchLine(line)
	if matched {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
