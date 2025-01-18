package regexp

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

type RegExp struct {
	mps        regExpElement
	matchStart bool
}

type regExpElement interface {
	fmt.Stringer
	matchHere(line []byte, ldx int) bool
}

type buildElement interface {
	regExpElement
	setNext(regExpElement)
}

type matchPoint interface {
	buildElement
}

type basicMatchPoint struct {
	MatchChars string
	Inverted   bool
	Next       regExpElement
}

type oneOrMoreMatchPoint struct {
	basicMatchPoint
}

type zeroOrMoreMatchPoint struct {
	basicMatchPoint
}

type zeroOrOneMatchPoint struct {
	basicMatchPoint
}

type matchEndMatchPoint struct{}

func (e matchEndMatchPoint) setNext(_ regExpElement) {
	// should exit as this would be a clear error?
}

var _ buildElement = matchEndMatchPoint{}

type Color int

func (re RegExp) String() string {
	return fmt.Sprintf("#[RegExp: '%s' %v]", re.mps, re.matchStart)
}

func (mp basicMatchPoint) basicString(mytype string) string {
	var remainder string
	if mp.Next == nil {
		remainder = "nil"
	} else {
		remainder = mp.Next.String()
	}
	return fmt.Sprintf("#[%s: '%s' %v %s]", mytype, mp.MatchChars, mp.Inverted, remainder)
}

func (mp basicMatchPoint) String() string {
	return mp.basicString("basicMatchPoint")
}

func (mp oneOrMoreMatchPoint) String() string {
	return mp.basicString("oneOrMoreMatchPoint")
}

func (mp zeroOrMoreMatchPoint) String() string {
	return mp.basicString("zeroOrMoreMatchPoint")
}

func (mp zeroOrOneMatchPoint) String() string {
	return mp.basicString("zeroOrOneMatchPoint")
}

func (e matchEndMatchPoint) String() string {
	return "[matchEnd]"
}

func (mp basicMatchPoint) matchByte(c byte) bool {
	matches := strings.Contains(mp.MatchChars, string(c))
	if mp.Inverted {
		matches = !matches
	}
	return matches
}

func (mp *basicMatchPoint) setNext(n regExpElement) {
	mp.Next = n
}

var (
	_ matchPoint    = &basicMatchPoint{}
	_ regExpElement = &basicMatchPoint{}
	_ buildElement  = &basicMatchPoint{}
)

func (re RegExp) matchHere(line []byte, ldx int) bool {
	fmt.Fprintf(os.Stderr, "matchHere(line='%s', ", string(line)[ldx:])
	fmt.Fprintf(os.Stderr, "lbx=%d)\n", ldx)
	return re.mps.matchHere(line, ldx)
}

func (mp basicMatchPoint) matchHere(line []byte, ldx int) bool {
	fmt.Fprintf(os.Stderr, "mp=%#v\n", mp)
	fmt.Fprintf(os.Stderr, "basicMatchPoint.matchHere('%s', %d)\n", string(line)[ldx:], ldx)
	if ldx >= len(line) {
		fmt.Fprintln(os.Stderr, "oops, got to long")
		return false
	}
	if !mp.matchByte(line[ldx]) {
		fmt.Fprintln(os.Stderr, "no match")
		return false
	}
	if mp.Next == nil {
		fmt.Fprintln(os.Stderr, "finished matching")
		return true
	}
	return mp.Next.matchHere(line, ldx+1)
}

func (mp zeroOrOneMatchPoint) matchHere(line []byte, ldx int) bool {
	fmt.Fprintf(os.Stderr, "mp=%#v\n", mp)
	fmt.Fprintf(os.Stderr, "zeroOrOneMatchPoint.matchHere('%s', %d)\n", string(line)[ldx:], ldx)
	if mp.Next == nil {
		fmt.Fprintln(os.Stderr, "short circuit match")
		return true
	}
	if ldx >= len(line) {
		fmt.Fprintln(os.Stderr, "at end, so trying zero length")
		return mp.Next.matchHere(line, ldx)
	}
	if !mp.matchByte(line[ldx]) {
		fmt.Fprintln(os.Stderr, "no match, so trying zero length")
		return mp.Next.matchHere(line, ldx)
	}
	return mp.Next.matchHere(line, ldx+1)
}

func (mp zeroOrMoreMatchPoint) matchHere(line []byte, ldx int) bool {
	fmt.Fprintf(os.Stderr, "mp=%#v\n", mp)
	fmt.Fprintf(os.Stderr, "zeroOrMoreMatchPoint.matchHere('%s', %d)\n", string(line)[ldx:], ldx)
	if mp.Next == nil {
		fmt.Fprintln(os.Stderr, "short circuit match")
		return true
	}
	if ldx >= len(line) {
		fmt.Fprintln(os.Stderr, "at end, so trying zero length")
		return mp.Next.matchHere(line, ldx)
	}
	// finding max length that will match and then working backwards
	maxLength := 0
	for ; maxLength+ldx < len(line); maxLength++ {
		if !mp.matchByte(line[maxLength+ldx]) {
			// keep incrementing until we fail to match or hit the end
			break
		}
	}
	fmt.Fprintf(os.Stderr, "maxLength: %d\n", maxLength)
	// here is the working backwards
	for trialLength := maxLength; trialLength >= 0; trialLength-- {
		fmt.Fprintf(os.Stderr, "trialLength: %d\n", trialLength)
		if mp.Next.matchHere(line, ldx+trialLength) {
			return true
		}
	}
	return false
}

func (mp oneOrMoreMatchPoint) matchHere(line []byte, ldx int) bool {
	fmt.Fprintf(os.Stderr, "mp=%#v\n", mp)
	fmt.Fprintf(os.Stderr, "zeroOrMoreMatchPoint.matchHere('%s', %d)\n", string(line)[ldx:], ldx)
	if mp.Next == nil {
		fmt.Fprintln(os.Stderr, "short circuit match")
		return true
	}
	if ldx >= len(line) {
		fmt.Fprintln(os.Stderr, "at end, so trying zero length")
		return mp.Next.matchHere(line, ldx)
	}
	// need at least one
	if !mp.matchByte(line[ldx]) {
		fmt.Fprintln(os.Stderr, "no match")
		return false
	}
	// finding max length that will match and then working backwards
	maxLength := 1
	for ; maxLength+ldx < len(line); maxLength++ {
		if !mp.matchByte(line[maxLength+ldx]) {
			// keep incrementing until we fail to match or hit the end
			break
		}
	}
	fmt.Fprintf(os.Stderr, "maxLength: %d\n", maxLength)
	// here is the working backwards
	for trialLength := maxLength; trialLength >= 0; trialLength-- {
		fmt.Fprintf(os.Stderr, "trialLength: %d\n", trialLength)
		if mp.Next.matchHere(line, ldx+trialLength) {
			return true
		}
	}
	return false
}

func (e matchEndMatchPoint) matchHere(line []byte, ldx int) bool {
	fmt.Fprintf(os.Stderr, "mp=%#v\n", e)
	fmt.Fprintf(os.Stderr, "matchEndMatchPoint.matchHere('%s', %d)\n", string(line)[ldx:], ldx)
	atEnd := ldx == len(line)
	if atEnd {
		fmt.Fprintf(os.Stderr, "Matched at end and regexp has $\n")
		return true
	}
	fmt.Fprintf(os.Stderr, "Not at end and regexp has $\n")
	return false
}

func ParseRegExp(pattern string) RegExp {
	regex := RegExp{}
	if len(pattern) == 0 {
		fmt.Fprintf(os.Stderr, "pattern must contain at least one character\n")
		os.Exit(-3)
	}
	if pattern[0] == '^' {
		regex.matchStart = true
		pattern = pattern[1:]
	}

	regex.mps = ParsePattern(pattern, 0)
	fmt.Fprintf(os.Stderr, "regex = '%+v'\n", regex)
	return regex
}

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

func makeBasicMatchPoint(b buildElement) *basicMatchPoint {
	mp, ok := b.(*basicMatchPoint)
	if !ok {
		fmt.Fprintf(os.Stderr, "Problem parsing Regexp around a '?'\n")
		os.Exit(3)
	}
	return mp
}

func ParsePattern(pattern string, start int) regExpElement {
	regex := []buildElement{}
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
				fmt.Fprintf(os.Stderr, "questionmark at start %v+??\n", p)
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
				fmt.Fprintf(os.Stderr, "plus at start %v+??\n", p)
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
				fmt.Fprintf(os.Stderr, "asterisk at start %v+??\n", p)
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

func (re *RegExp) MatchLine(line []byte) bool {
	limit := len(line)
	if re.matchStart {
		limit = 1
	}
	for current := 0; current < limit; current++ {
		matchesHere := re.matchHere(line, current)
		if matchesHere {
			fmt.Fprintln(os.Stderr, "ml whole matched")
			return true
		}
	}

	fmt.Fprintln(os.Stderr, "ml whole fails")
	return false
}
