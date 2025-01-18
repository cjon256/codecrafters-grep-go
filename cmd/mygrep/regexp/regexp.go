package regexp

import (
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
	debugf("line='%s'\n", line)
	if re.matchStart {
		debugf("whole matched\n")
		return re.mps.matchHere(line, 0)
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
func parseSetPattern(inverted bool, pattern *string, index *int) (matchPoint, error) {
	retval := basicMatchPoint{}
	retval.Inverted = inverted
	chars := []byte{}
	for *index < len(*pattern) {
		switch (*pattern)[*index] {
		case ']':
			retval.MatchChars = string(chars[:])
			return &retval, nil
		default:
			chars = append(chars, (*pattern)[*index])
		}
		(*index)++
	}
	return &retval, errors.New("parse pattern not closed")
}

// used when embedding a basicMatchPoint because it's followed by a ?+*
func makeBasicMatchPoint(b matchPoint) *basicMatchPoint {
	mp, ok := b.(*basicMatchPoint)
	if !ok {
		debugf("Problem parsing Regexp around a '?'\n")
		os.Exit(3)
	}
	return mp
}

const (
	digits    = "0123456789"
	alpha     = "abcdefghijklmnopqrstuvwxyz"
	wordChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" + alpha + digits + "_"
)

// returns a linked list representing the regexp pattern
func ParsePattern(pattern string, start int) matchPoint {
	regex := []matchPoint{}
	var p matchPoint
	for rdx := start; rdx < len(pattern); {
		var err error
		switch pattern[rdx] {
		case '[':
			rdx++
			if rdx < len(pattern) && pattern[rdx] == '^' {
				rdx++
				p, err = parseSetPattern(true, &pattern, &rdx)
			} else {
				p, err = parseSetPattern(false, &pattern, &rdx)
			}
			if err != nil {
				os.Exit(3)
			}

		case '?':
			if len(regex) == 0 {
				// handle + at start of string I guess...
				debugf("questionmark at start %v+??\n", p)
				p = &basicMatchPoint{MatchChars: string(pattern[rdx])}
			} else {
				// set the last mp to be one or more...
				tmp := makeBasicMatchPoint(regex[len(regex)-1])
				regex[len(regex)-1] = &zeroOrOneMatchPoint{*tmp}
				rdx++
				continue // don't add a matchPoint
			}

		case '+':
			if len(regex) == 0 {
				// handle + at start of string I guess...
				debugf("plus at start %v+??\n", p)
				p = &basicMatchPoint{MatchChars: string(pattern[rdx])}
			} else {
				// set the last mp to be one or more...
				tmp := makeBasicMatchPoint(regex[len(regex)-1])
				regex[len(regex)-1] = &oneOrMoreMatchPoint{*tmp}
				rdx++
				continue // don't add a matchPoint
			}

		case '*':
			if len(regex) == 0 {
				// handle * at start of string I guess...
				debugf("asterisk at start %v+??\n", p)
				p = &basicMatchPoint{MatchChars: string(pattern[rdx])}
			} else {
				// set the last tmp to be one or more...
				tmp := makeBasicMatchPoint(regex[len(regex)-1])
				regex[len(regex)-1] = &zeroOrMoreMatchPoint{*tmp}
				rdx++
				continue // don't add a matchPoint
			}

		case '$':
			if rdx == len(pattern)-1 {
				p = &matchEndMatchPoint{}
			} else {
				p = &basicMatchPoint{MatchChars: string(pattern[rdx])}
			}

		case '.':
			p = &basicMatchPoint{"", true, nil}

		case '\\':
			rdx++
			if rdx < len(pattern) {
				switch pattern[rdx] {
				case 'w':
					p = &basicMatchPoint{MatchChars: wordChars}
				case 'd':
					p = &basicMatchPoint{MatchChars: digits}
				default:
					p = &basicMatchPoint{MatchChars: string(pattern[rdx])}
				}
			} else {
				// last character was a backslash....
				// I guess append a backslash character?
				p = &basicMatchPoint{MatchChars: "\\"}
			}

		default:
			p = &basicMatchPoint{MatchChars: string(pattern[rdx])}
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
