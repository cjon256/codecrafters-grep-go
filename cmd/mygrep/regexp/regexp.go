package regexp

import (
	"bytes"
	"errors"
	"fmt"
	"os"
)

type RegExp struct {
	mps        matchPoint
	matchStart bool
}

func (re RegExp) String() string {
	return fmt.Sprintf("#[RegExp(matchStart=%v): '%s']", re.matchStart, re.mps)
}

func (re *RegExp) MatchLine(line []byte) bool {
	// trim any newline off of that in case we forget -n for echo
	line = bytes.TrimRight(line, "\n\r")

	debugf("line='%s'\n", line)
	if re.matchStart {
		matched := re.mps.matchHere(line, 0)
		if matched {
			debugf("whole matched\n")
			return true
		}
		return false
	}
	for ldx := 0; ldx < len(line); ldx++ {
		if re.mps.matchHere(line, ldx) {
			debugf("whole matched")
			return true
		}
	}

	debugf("whole fails\n")
	return false
}

func ParseRegExp(pattern string) RegExp {
	regex := RegExp{}
	if len(pattern) == 0 {
		debugf("pattern must contain at least one character\n")
		os.Exit(-3)
	}
	if pattern[0] == '^' {
		regex.matchStart = true
		pattern = pattern[1:]
	}

	regex.mps = ParsePattern(pattern, 0)
	debugf("regex = '%+v'\n", regex)
	return regex
}

// called when we are inside a [abcd] pattern
func parseSetPattern(pattern *string, rdx *int) (*basicMatchPoint, error) {
	retval := basicMatchPoint{}
	inverted := false
	if *rdx < len(*pattern) && (*pattern)[*rdx] == '^' {
		*rdx++
		inverted = true
	}
	retval.Inverted = inverted
	chars := []byte{}
	for *rdx < len(*pattern) {
		switch (*pattern)[*rdx] {
		case ']':
			retval.MatchChars = string(chars[:])
			return &retval, nil
		default:
			chars = append(chars, (*pattern)[*rdx])
		}
		(*rdx)++
	}
	return &retval, errors.New("parse pattern not closed")
}

const (
	digits    = "0123456789"
	alpha     = "abcdefghijklmnopqrstuvwxyz"
	wordChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" + alpha + digits + "_"
)

// returns a linked list representing the regexp pattern
func ParsePattern(pattern string, start int) matchPoint {
	var rdx int
	glob := func(mp *basicMatchPoint) matchPoint {
		if rdx+1 >= len(pattern) {
			debugf("regex glob: no glob, at end with '%s'\n", string(pattern[rdx]))
			return mp
		}
		debugf("pattern='%s' rdx=%d\n", pattern, rdx)
		switch pattern[rdx+1] {
		case '?':
			rdx++
			debugf("regex glob: got '?'\n")
			return &zeroOrOneMatchPoint{*mp}
		case '+':
			rdx++
			debugf("regex glob: got '+'\n")
			return &oneOrMoreMatchPoint{*mp}
		case '*':
			rdx++
			debugf("regex glob: got '*'\n")
			return &zeroOrMoreMatchPoint{*mp}
		default:
			debugf("regex glob: no glob, got '%s'\n", string(pattern[rdx+1]))
			return mp
		}
	}

	regex := []matchPoint{}
	var p matchPoint
	for rdx = start; rdx < len(pattern); {
		var err error
		switch pattern[rdx] {
		case '[':
			rdx++
			var b *basicMatchPoint
			b, err = parseSetPattern(&pattern, &rdx)
			if err != nil {
				os.Exit(3)
			}
			p = glob(b)

		case '$':
			if rdx == len(pattern)-1 {
				p = &matchEndMatchPoint{}
			} else {
				p = glob(&basicMatchPoint{MatchChars: string(pattern[rdx])})
			}

		case '.':
			debugf("regex parse: got '.'\n")
			p = glob(&basicMatchPoint{"", true, nil})

		case '\\':
			rdx++
			if rdx < len(pattern) {
				switch pattern[rdx] {
				case 'w':
					p = glob(&basicMatchPoint{MatchChars: wordChars})
				case 'd':
					p = glob(&basicMatchPoint{MatchChars: digits})
				default:
					p = glob(&basicMatchPoint{MatchChars: string(pattern[rdx])})
				}
			} else {
				// last character was a backslash....
				// I guess append a backslash character?
				p = glob(&basicMatchPoint{MatchChars: "\\"})
			}

		default:
			p = glob(&basicMatchPoint{MatchChars: string(pattern[rdx])})
		}
		regex = append(regex, p)
		rdx++
	}
	if len(regex) == 0 {
		// XXX should probably just bail if this is the case?
		return nil
	}
	for i := 0; i < len(regex)-1; i++ {
		regex[i].setNext(regex[i+1])
	}
	return regex[0]
}
