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
	setWildtype(wildcard)
}

type matchPoint interface {
	buildElement
}

type basicMatchPoint struct {
	MatchChars string
	Inverted   bool
	Wildtype   wildcard
	Next       regExpElement
}

type zeroOrOneMatchPoint struct {
	basicMatchPoint
}

type matchEndMatchPoint struct{}

func (e matchEndMatchPoint) setNext(_ regExpElement) {
	// should exit as this would be a clear error?
}

func (e matchEndMatchPoint) String() string {
	return "[matchEnd]"
}

func (e matchEndMatchPoint) matchHere(line []byte, ldx int) bool {
	atEnd := ldx == len(line)
	if atEnd {
		fmt.Fprintf(os.Stderr, "Matched at end and regexp has $\n")
		return true
	}
	fmt.Fprintf(os.Stderr, "Not at end and regexp has $\n")
	return false
}

func (e matchEndMatchPoint) setWildtype(_ wildcard) {
}

var _ buildElement = matchEndMatchPoint{}

type wildcard int

const (
	normal wildcard = iota
	oneOrMore
	zeroOrOne
	zeroOrMore
)

type Color int

func (re RegExp) String() string {
	return fmt.Sprintf("#[RegExp: '%s' %v]", re.mps, re.matchStart)
}

func (mp basicMatchPoint) String() string {
	if mp.Next == nil {
		return fmt.Sprintf("#[basicMatchPoint: '%s' %v %v nil]", mp.MatchChars, mp.Inverted, mp.Wildtype)
	}
	remainder := mp.Next.String()
	return fmt.Sprintf("#[basicMatchPoint: '%s' %v %v %s]", mp.MatchChars, mp.Inverted, mp.Wildtype, remainder)
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

func (mp *basicMatchPoint) setWildtype(w wildcard) {
	mp.Wildtype = w
}

var (
	_ matchPoint    = &basicMatchPoint{}
	_ regExpElement = &basicMatchPoint{}
	_ buildElement  = &basicMatchPoint{}
)

func (re RegExp) matchHere(line []byte, ldx int) bool {
	fmt.Fprintf(os.Stderr, "matchHere(line='%s', ", string(line)[ldx:])
	fmt.Fprintf(os.Stderr, "current=%d)\n", ldx)
	return re.mps.matchHere(line, ldx)
}

func (mp basicMatchPoint) matchHere(line []byte, ldx int) bool {
	fmt.Fprintf(os.Stderr, "basicMatchPoint.matchHere('%s', %d\n", string(line)[ldx:], ldx)
	if ldx >= len(line) {
		fmt.Fprintln(os.Stderr, "at the end...")
		// ah, we're at the end of the string but not the end of the regexp
		// but we might still match if all the remaining items can be zero length
		if mp.Wildtype == zeroOrOne {
			if mp.Next == nil {
				return true
			}
			return mp.Next.matchHere(line, ldx)
		}
		fmt.Fprintln(os.Stderr, "oops, got to long")
		return false
	}
	switch mp.Wildtype {
	case oneOrMore:
		return mp.matchOneOrMore(line, ldx)
	case zeroOrOne:
		return mp.matchZeroOrOne(line, ldx)
	case zeroOrMore:
		return mp.matchZeroOrMore(line, ldx)
	default:
		if !mp.matchByte(line[ldx]) {
			return false
		}
		if mp.Next == nil {
			return true
		}
		return mp.Next.matchHere(line, ldx+1)
	}
}

func (mp basicMatchPoint) matchZeroOrOne(line []byte, ldx int) bool {
	if mp.Next == nil {
		return true
	}
	if !mp.matchByte(line[ldx]) {
		return mp.Next.matchHere(line, ldx)
	}
	return mp.Next.matchHere(line, ldx+1)
}

func (mp basicMatchPoint) matchZeroOrMore(line []byte, ldx int) bool {
	fmt.Fprintf(os.Stderr, "matchOneOrMore(line='%s', current=%d) with mp = '%v+'\n", string(line)[ldx:], ldx, mp)
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
		if mp.Next == nil {
			return true
		}
		if mp.Next.matchHere(line, ldx+trialLength) {
			return true
		}
	}
	return false
}

func (mp basicMatchPoint) matchOneOrMore(line []byte, current int) bool {
	fmt.Fprintf(os.Stderr, "matchOneOrMore(line='%s', current=%d) with mp = '%v+'\n", string(line)[current:], current, mp)
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
		if mp.Next == nil {
			return true
		}
		if mp.Next.matchHere(line, current+trialLength) {
			return true
		}
	}
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
				regex[len(regex)-1].setWildtype(zeroOrOne)
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
				regex[len(regex)-1].setWildtype(oneOrMore)
				rdx++
				continue // don't add a matchPoint
			}

		case '*':
			if len(regex) == 0 {
				// handle * at start of string I guess...
				fmt.Fprintf(os.Stderr, "asterisk at start %v+??\n", p)
				p = &basicMatchPoint{MatchChars: string(pattern[rdx])}
			} else {
				// set the last mp to be one or more...
				regex[len(regex)-1].setWildtype(zeroOrMore)
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
			p = &basicMatchPoint{"", true, normal, nil}

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
