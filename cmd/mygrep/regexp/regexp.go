package regexp

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
)

///////////////////////////////////////////////////////////
// A useful function for debugging

// DEBUG flag, set to true to enable debug mode
var DEBUG bool = false

// debugf convenience function
func debugf(format string, args ...interface{}) {
	if DEBUG {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}

///////////////////////////////////////////////////////////
// RegExp class and constructor function for it

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
		matched, _ := re.mps.matchHere(line, 0, false)
		if matched {
			debugf("whole matched\n")
			return true
		}
		return false
	}
	for ldx := 0; ldx < len(line); ldx++ {
		matched, _ := re.mps.matchHere(line, ldx, false)
		if matched {
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

	regex.mps = parsePattern(pattern)
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
	retval.inverted = inverted
	chars := []byte{}
	for *rdx < len(*pattern) {
		switch (*pattern)[*rdx] {
		case ']':
			retval.matchChars = string(chars[:])
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

type groupHead struct {
	heads     []matchPoint
	tails     []matchPoint
	groupTail matchPoint
}

type groupOneOrMoreHead struct {
	groupHead
}

type groupZeroOrOneHead struct {
	groupHead
}

type groupZeroOrMoreHead struct {
	groupHead
}

type groupTail struct {
	next matchPoint
}

func (gh groupHead) matchHere(line []byte, ldx int, isSpecial bool) (bool, int) {
	for i := 0; i < len(gh.heads); i++ {
		matched, bytesUsed := gh.heads[i].matchHere(line, ldx, isSpecial)
		if matched {
			return true, bytesUsed
		}
	}
	return false, 0
}

func (gt groupTail) matchHere(line []byte, ldx int, isSpecial bool) (bool, int) {
	if isSpecial {
		return true, 0
	}
	return gt.next.matchHere(line, ldx, isSpecial)
}

func (gt groupTail) String() string {
	remainder := ""
	if gt.next != nil {
		remainder = gt.next.String()
	}
	return "[groupTail] " + remainder
}

func (gh groupHead) String() string {
	fmt.Fprintf(os.Stderr, "got here %p %#v\n", gh, gh)
	remainder := ""
	if gh.heads[0] != nil {
		remainder = gh.heads[0].String()
	}
	return "[groupHead] " + remainder
}

func (gt *groupTail) setNext(n matchPoint) {
	gt.next = n
}

func (gh *groupHead) setNext(n matchPoint) {
	gh.groupTail = n
}

// returns a linked list representing the regexp pattern
func parsePattern(pattern string) matchPoint {
	var rdx int
	var parseHere func(bool) (matchPoint, matchPoint)

	parseGroup := func() (matchPoint, matchPoint) {
		gh := groupHead{}
		gt := groupTail{}
		for {
			head, tail := parseHere(true)
			if head == nil || rdx >= len(pattern) {
				// reached end of string I guess?
				fmt.Fprintf(os.Stderr, "error when parsing a group\n")
				os.Exit(3)
			}
			fmt.Fprintf(os.Stderr, "head = got this %s\n", head)
			gh.heads = append(gh.heads, head)
			gh.tails = append(gh.tails, tail)
			tail.setNext(&gt)
			fmt.Fprintf(os.Stderr, "gh (%p) = have this %s\n", &gh, gh)
			switch pattern[rdx] {
			case ')':
				return &gh, &gt
				// incrementing rdx handled by caller
			case '|':
				rdx++
			default:
				fmt.Fprintf(os.Stderr, "unexpected character: '%s' - error when parsing a group\n", string(pattern[rdx]))
				os.Exit(3)
			}
		}
	}

	// handles ? + * characters when they glob
	// will not be used if at start of line or after \
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

	parseHere = func(isGroup bool) (matchPoint, matchPoint) {
		regex := []matchPoint{}
		var p matchPoint
	loop:
		for rdx < len(pattern) {
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

			case '(':
				rdx++ // move past (
				p, q := parseGroup()
				regex = append(regex, p)
				regex = append(regex, q)
				fmt.Fprintf(os.Stderr, "pos = %s\n", pattern[rdx:])
				rdx++ // move past )
				continue

			case '|':
				fallthrough
			case ')':
				if isGroup {
					break loop
				} else {
					p = glob(&basicMatchPoint{matchChars: string(pattern[rdx])})
				}

			case '$':
				if rdx == len(pattern)-1 {
					p = &matchEndMatchPoint{}
				} else {
					p = glob(&basicMatchPoint{matchChars: string(pattern[rdx])})
				}

			case '.':
				debugf("regex parse: got '.'\n")
				p = glob(&basicMatchPoint{"", true, nil})

			case '\\':
				rdx++
				if rdx < len(pattern) {
					switch pattern[rdx] {
					case 'w':
						p = glob(&basicMatchPoint{matchChars: wordChars})
					case 'd':
						p = glob(&basicMatchPoint{matchChars: digits})
					default:
						p = glob(&basicMatchPoint{matchChars: string(pattern[rdx])})
					}
				} else {
					// last character was a backslash....
					// I guess append a backslash character?
					p = glob(&basicMatchPoint{matchChars: "\\"})
				}

			default:
				p = glob(&basicMatchPoint{matchChars: string(pattern[rdx])})
			}
			regex = append(regex, p)
			rdx++
		}
		if len(regex) == 0 {
			// XXX should probably just bail if this is the case?
			return nil, nil
		}
		for i := 0; i < len(regex)-1; i++ {
			regex[i].setNext(regex[i+1])
		}
		return regex[0], regex[len(regex)-1]
	}
	retval, _ := parseHere(false)
	return retval
}

///////////////////////////////////////////////////////////
// matchPoints performs matching at a single point

type matchPoint interface {
	fmt.Stringer
	matchHere(line []byte, ldx int, isSpecial bool) (bool, int)
	setNext(matchPoint)
}

type basicMatchPoint struct {
	matchChars string
	inverted   bool
	next       matchPoint
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

// checking interfaces are implemented fully
var (
	_ matchPoint = &basicMatchPoint{}
	_ matchPoint = &oneOrMoreMatchPoint{}
	_ matchPoint = &zeroOrMoreMatchPoint{}
	_ matchPoint = &zeroOrOneMatchPoint{}
	_ matchPoint = matchEndMatchPoint{}
	_ matchPoint = &groupHead{}
	_ matchPoint = &groupTail{}
)

func (mp basicMatchPoint) recursiveString(mytype string) string {
	var remainder string
	if mp.next == nil {
		remainder = ""
	} else {
		remainder = ", " + mp.next.String()
	}
	invChar := ""
	if mp.inverted {
		invChar = "^"
	}
	return fmt.Sprintf("%s: [%s%s]%s", mytype, invChar, mp.matchChars, remainder)
}

func (mp basicMatchPoint) String() string {
	return mp.recursiveString("basic")
}

func (mp oneOrMoreMatchPoint) String() string {
	return mp.recursiveString("oneOrMore")
}

func (mp zeroOrMoreMatchPoint) String() string {
	return mp.recursiveString("zeroOrMore")
}

func (mp zeroOrOneMatchPoint) String() string {
	return mp.recursiveString("zeroOrOne")
}

func (e matchEndMatchPoint) String() string {
	return "[end '$']"
}

func (mp basicMatchPoint) matchByte(c byte) bool {
	matches := strings.Contains(mp.matchChars, string(c))
	if mp.inverted {
		matches = !matches
	}
	return matches
}

func (mp basicMatchPoint) matchHere(line []byte, ldx int, isSpecial bool) (bool, int) {
	debugf("mp=%#v\n", mp)
	debugf("basicMatchPoint.matchHere('%s', %d)\n", string(line)[ldx:], ldx)
	if ldx >= len(line) {
		debugf("oops, got to long\n")
		return false, 0
	}
	if !mp.matchByte(line[ldx]) {
		debugf("no match\n")
		return false, 0
	}
	if mp.next == nil {
		debugf("finished matching\n")
		return true, 1
	}
	return mp.next.matchHere(line, ldx+1, isSpecial)
}

func (mp zeroOrOneMatchPoint) matchHere(line []byte, ldx int, isSpecial bool) (bool, int) {
	debugf("mp=%#v\n", mp)
	debugf("zeroOrOneMatchPoint.matchHere('%s', %d)\n", string(line)[ldx:], ldx)
	// XXX ah, but we don't want to short circuit if inGroup
	if !isSpecial && mp.next == nil {
		debugf("short circuit match\n")
		return true, 0
	}
	if ldx >= len(line) {
		debugf("at end, so trying zero length\n")
		return mp.next.matchHere(line, ldx, isSpecial)
	}
	if !mp.matchByte(line[ldx]) {
		debugf("no match, so trying zero length\n")
		return mp.next.matchHere(line, ldx, isSpecial)
	}
	return mp.next.matchHere(line, ldx+1, isSpecial)
}

func (mp zeroOrMoreMatchPoint) matchHere(line []byte, ldx int, isSpecial bool) (bool, int) {
	debugf("mp=%#v\n", mp)
	debugf("zeroOrMoreMatchPoint.matchHere('%s', %d)\n", string(line)[ldx:], ldx)
	if !isSpecial && mp.next == nil {
		debugf("short circuit match\n")
		return true, 0
	}
	if ldx >= len(line) {
		debugf("at end, so trying zero length\n")
		return mp.next.matchHere(line, ldx, isSpecial)
	}
	// finding max length that will match and then working backwards
	maxLength := 0
	for ; maxLength+ldx < len(line); maxLength++ {
		if !mp.matchByte(line[maxLength+ldx]) {
			// keep incrementing until we fail to match or hit the end
			break
		}
	}
	debugf("maxLength: %d\n", maxLength)
	// here is the working backwards
	for trialLength := maxLength; trialLength >= 0; trialLength-- {
		debugf("trialLength: %d\n", trialLength)
		matched, bytesUsed := mp.next.matchHere(line, ldx+trialLength, isSpecial)
		if matched {
			return true, bytesUsed
		}
	}
	return false, 0
}

func (mp oneOrMoreMatchPoint) matchHere(line []byte, ldx int, isSpecial bool) (bool, int) {
	debugf("mp=%#v\n", mp)
	debugf("zeroOrMoreMatchPoint.matchHere('%s', %d)\n", string(line)[ldx:], ldx)
	if ldx >= len(line) {
		debugf("at end, so trying zero length\n")
		return mp.next.matchHere(line, ldx, isSpecial)
	}
	// need at least one
	if !mp.matchByte(line[ldx]) {
		debugf("no match\n")
		return false, 0
	}
	// finding max length that will match and then working backwards
	maxLength := 1
	for ; maxLength+ldx < len(line); maxLength++ {
		if !mp.matchByte(line[maxLength+ldx]) {
			// keep incrementing until we fail to match or hit the end
			break
		}
	}
	debugf("maxLength: %d\n", maxLength)
	// here is the working backwards
	for trialLength := maxLength; trialLength >= 0; trialLength-- {
		debugf("trialLength: %d\n", trialLength)
		matched, bytesUsed := mp.next.matchHere(line, ldx+trialLength, isSpecial)
		if matched {
			return true, bytesUsed
		}
	}
	return false, 0
}

func (e matchEndMatchPoint) matchHere(line []byte, ldx int, _ bool) (bool, int) {
	debugf("mp=%#v\n", e)
	debugf("matchEndMatchPoint.matchHere('%s', %d)\n", string(line)[ldx:], ldx)
	atEnd := ldx == len(line)
	if atEnd {
		debugf("Matched at end and regexp has $\n")
		return true, 0
	}
	debugf("Not at end and regexp has $\n")
	return false, 0
}

func (mp *basicMatchPoint) setNext(n matchPoint) {
	mp.next = n
}

func (e matchEndMatchPoint) setNext(_ matchPoint) {
	// should exit as this would be a clear error?
	os.Exit(3)
}
