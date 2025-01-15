package myregexp

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

var (
	digits    = "0123456789"
	alpha     = "abcdefghijklmnopqrstuvwxyz"
	wordChars = alpha + strings.ToUpper(alpha) + digits + "_"
)

type MyRegExp struct {
	mps        []matchPoint
	matchStart bool
	matchEnd   bool
}

type wildcard int

const (
	normal wildcard = iota
	oneOrMore
	zeroOrOne
)

type Color int

type matchPoint struct {
	matchChars string
	inverted   bool
	wildtype   wildcard
}

func (mp matchPoint) matchByte(c byte) bool {
	matches := strings.Contains(mp.matchChars, string(c))
	if mp.inverted {
		matches = !matches
	}
	return matches
}

func (re MyRegExp) matchHere(line []byte, current int, rdx int) bool {
	if rdx >= len(re.mps) {
		// so we're at the end of the regexp
		// we return true unless there was a $ at the end... if that is the case
		// then we need to check we are at the end of the input
		if !re.matchEnd {
			return true
		}
		// ah, matchEnd was set so we succeed if at end and fail if not...
		atEnd := current == len(line)
		if !atEnd {
			fmt.Fprintln(os.Stderr, "match fails because not at end$")
		}
		return atEnd
	}
	mp := re.mps[rdx]
	fmt.Fprintf(os.Stderr, "matchHere(line='%s', current=%d, rdx=%d) with mp = '%v+'\n", string(line)[current:], current, rdx, mp)
	if current >= len(line) {
		fmt.Fprintln(os.Stderr, "at the end...")
		// ah, we're at the end of the string but not the end of the regexp
		// but we might still match if all the remaining items can be zero length
		if mp.wildtype == zeroOrOne {
			return re.matchHere(line, current, rdx+1)
		}
		fmt.Fprintln(os.Stderr, "oops, got to long")
		return false
	}
	switch mp.wildtype {
	case oneOrMore:
		return re.matchOneOrMore(line, current, rdx)
	case zeroOrOne:
		return re.matchZeroOrOne(line, current, rdx)
	default:
		if !mp.matchByte(line[current]) {
			return false
		}
		return re.matchHere(line, current+1, rdx+1)
	}
}

func (re MyRegExp) matchZeroOrOne(line []byte, current int, rdx int) bool {
	mp := re.mps[rdx]
	if !mp.matchByte(line[current]) {
		return re.matchHere(line, current, rdx+1)
	}
	return re.matchHere(line, current+1, rdx+1)
}

func (re MyRegExp) matchOneOrMore(line []byte, current int, rdx int) bool {
	mp := re.mps[rdx]
	fmt.Fprintf(os.Stderr, "matchOneOrMore(line='%s', current=%d, rdx=%d) with mp = '%v+'\n", string(line)[current:], current, rdx, mp)
	if !mp.matchByte(line[current]) {
		return false
	}
	// finding max length that will match and then working backwards
	maxLength := 1
	for ; maxLength+current < len(line); maxLength++ {
		if !mp.matchByte(line[maxLength+current]) {
			// keep incrementing until we fail to match or hit the end
			break
		}
	}
	fmt.Fprintf(os.Stderr, "maxLength: %d\n", maxLength)
	// here is the working backwards
	for trialLength := maxLength; trialLength > 0; trialLength-- {
		fmt.Fprintf(os.Stderr, "trialLength: %d\n", trialLength)
		if re.matchHere(line, current+trialLength, rdx+1) {
			return true
		}
	}
	return false
}

func parseSetPattern(inverted bool, pattern *string, index *int) (matchPoint, error) {
	retval := matchPoint{}
	retval.inverted = inverted
	chars := []byte{}
	for *index < len(*pattern) {
		switch (*pattern)[*index] {
		case ']':
			retval.matchChars = string(chars[:])
			return retval, nil
		default:
			chars = append(chars, (*pattern)[*index])
		}
		(*index)++
	}
	return retval, errors.New("parse pattern not closed")
}

func ParsePattern(pattern string) (MyRegExp, error) {
	index := 0
	regex := MyRegExp{}
	if pattern[index] == '^' {
		regex.matchStart = true
		index++
	}

	limit := len(pattern)
	if pattern[limit-1] == '$' {
		fmt.Fprintf(os.Stderr, "matched $ at end: %s\n", pattern)
		limit--
		regex.matchEnd = true
	}

	var p matchPoint
	for index < limit {
		var err error
		switch pattern[index] {
		case '[':
			index++
			if index < limit && pattern[index] == '^' {
				index++
				p, err = parseSetPattern(true, &pattern, &index)
			} else {
				p, err = parseSetPattern(false, &pattern, &index)
			}
			if err != nil {
				os.Exit(3)
			}
		case '?':
			if len(regex.mps) == 0 {
				// handle + at start of string I guess...
				fmt.Fprintf(os.Stderr, "questionmark at start %v+??\n", p)
				p = matchPoint{matchChars: string(pattern[index])}
			} else {
				// set the last mp to be one or more...
				regex.mps[len(regex.mps)-1].wildtype = zeroOrOne
				index++
				continue // don't add a matchPoint
			}
		case '+':
			if len(regex.mps) == 0 {
				// handle + at start of string I guess...
				fmt.Fprintf(os.Stderr, "plus at start %v+??\n", p)
				p = matchPoint{matchChars: string(pattern[index])}
			} else {
				// set the last mp to be one or more...
				regex.mps[len(regex.mps)-1].wildtype = oneOrMore
				index++
				continue // don't add a matchPoint
			}
		case '.':
			p = matchPoint{"", true, normal}
		case '\\':
			index++
			if index < limit {
				switch pattern[index] {
				case 'w':
					p = matchPoint{matchChars: wordChars}
				case 'd':
					p = matchPoint{matchChars: digits}
				default:
					p = matchPoint{matchChars: string(pattern[index])}
				}
			} else {
				// last character was a backslash....
				// I guess append a backslash character?
				p = matchPoint{matchChars: "\\"}
			}
		default:
			p = matchPoint{matchChars: string(pattern[index])}
		}
		regex.mps = append(regex.mps, p)
		index++
	}
	return regex, nil
}

func (re *MyRegExp) MatchLine(line []byte) bool {
	limit := len(line)
	if re.matchStart {
		limit = 1
	}
	for current := 0; current < limit; current++ {
		matchesHere := re.matchHere(line, current, 0)
		if matchesHere {
			fmt.Fprintln(os.Stderr, "ml whole matched")
			return true
		}
	}

	fmt.Fprintln(os.Stderr, "ml whole fails")
	return false
}
