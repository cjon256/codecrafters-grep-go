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

type regExpElement interface {
	fmt.Stringer
	matchHere(line []byte, ldx int) bool
}

type buildElement interface {
	regExpElement
	setNext(regExpElement)
	setWildtype(wildcard)
}

type RegExp struct {
	mps        regExpElement
	matchStart bool
	matchEnd   bool
}

type wildcard int

const (
	normal wildcard = iota
	oneOrMore
	zeroOrOne
	zeroOrMore
)

type Color int

type matchPoint struct {
	matchChars string
	inverted   bool
	wildtype   wildcard
	next       regExpElement
}

func (re RegExp) String() string {
	return fmt.Sprintf("#[RegExp: '%s' %v %v]", re.mps, re.matchStart, re.matchEnd)
}

func (mp matchPoint) String() string {
	if mp.next == nil {
		return fmt.Sprintf("#[matchPoint: '%s' %v %v nil]", mp.matchChars, mp.inverted, mp.wildtype)
	}
	remainder := fmt.Sprintf("%s", mp.next)
	return fmt.Sprintf("#[matchPoint: '%s' %v %v %s]", mp.matchChars, mp.inverted, mp.wildtype, remainder)
}

func (mp matchPoint) matchByte(c byte) bool {
	matches := strings.Contains(mp.matchChars, string(c))
	if mp.inverted {
		matches = !matches
	}
	return matches
}

func (mp matchPoint) matchHere(line []byte, ldx int) bool {
	fmt.Fprintf(os.Stderr, "matchPoint.matchHere('%s', %d\n", string(line)[ldx:], ldx)
	return true
}

func (mp *matchPoint) setNext(n regExpElement) {
	mp.next = n
}

func (mp *matchPoint) setWildtype(w wildcard) {
	mp.wildtype = w
}

var (
	_ regExpElement = &matchPoint{}
	_ buildElement  = &matchPoint{}
)

func (re RegExp) matchHere(line []byte, current int, rdx int) bool {
	fmt.Fprintf(os.Stderr, "matchHere(line='%s', ", string(line)[current:])
	fmt.Fprintf(os.Stderr, "current=%d, ", current)
	fmt.Fprintf(os.Stderr, "rdx=%d)\n", rdx)
	return true
}

// 	if rdx >= len(re.mps) {
// 		// so we're at the end of the regexp
// 		// we return true unless there was a $ at the end...
// 		if !re.matchEnd {
// 			//
// 			return true
// 		}
// 		// matchEnd was set so we succeed if we matched at end, fail if not
// 		// so we need to check we are at the end of the input
// 		atEnd := current == len(line)
// 		if !atEnd {
// 			fmt.Fprintln(os.Stderr, "match fails because not at end$")
// 		}
// 		return atEnd
// 	}
// 	mp := re.mps[rdx]
// 	fmt.Fprintf(os.Stderr, "matchHere(line='%s', ", string(line)[current:])
// 	fmt.Fprintf(os.Stderr, "current=%d, ", current)
// 	fmt.Fprintf(os.Stderr, "rdx=%d) with ", rdx)
// 	fmt.Fprintf(os.Stderr, "mp = '%v'\n", mp)
// 	if current >= len(line) {
// 		fmt.Fprintln(os.Stderr, "at the end...")
// 		// ah, we're at the end of the string but not the end of the regexp
// 		// but we might still match if all the remaining items can be zero length
// 		if mp.wildtype == zeroOrOne {
// 			return re.matchHere(line, current, rdx+1)
// 		}
// 		fmt.Fprintln(os.Stderr, "oops, got to long")
// 		return false
// 	}
// 	switch mp.wildtype {
// 	case oneOrMore:
// 		return re.matchOneOrMore(line, current, rdx)
// 	case zeroOrOne:
// 		return re.matchZeroOrOne(line, current, rdx)
// 	case zeroOrMore:
// 		return re.matchZeroOrMore(line, current, rdx)
// 	default:
// 		if !mp.matchByte(line[current]) {
// 			return false
// 		}
// 		return re.matchHere(line, current+1, rdx+1)
// 	}
// }
//
// func (re MyRegExp) matchZeroOrOne(line []byte, current int, rdx int) bool {
// 	mp := re.mps[rdx]
// 	if !mp.matchByte(line[current]) {
// 		return re.matchHere(line, current, rdx+1)
// 	}
// 	return re.matchHere(line, current+1, rdx+1)
// }
//
// func (re MyRegExp) matchZeroOrMore(line []byte, current int, rdx int) bool {
// 	mp := re.mps[rdx]
// 	fmt.Fprintf(os.Stderr, "matchOneOrMore(line='%s', current=%d, rdx=%d) with mp = '%v+'\n", string(line)[current:], current, rdx, mp)
// 	// finding max length that will match and then working backwards
// 	maxLength := 0
// 	for ; maxLength+current < len(line); maxLength++ {
// 		if !mp.matchByte(line[maxLength+current]) {
// 			// keep incrementing until we fail to match or hit the end
// 			break
// 		}
// 	}
// 	fmt.Fprintf(os.Stderr, "maxLength: %d\n", maxLength)
// 	// here is the working backwards
// 	for trialLength := maxLength; trialLength >= 0; trialLength-- {
// 		fmt.Fprintf(os.Stderr, "trialLength: %d\n", trialLength)
// 		if re.matchHere(line, current+trialLength, rdx+1) {
// 			return true
// 		}
// 	}
// 	return false
// }
//
// func (re MyRegExp) matchOneOrMore(line []byte, current int, rdx int) bool {
// 	mp := re.mps[rdx]
// 	fmt.Fprintf(os.Stderr, "matchOneOrMore(line='%s', current=%d, rdx=%d) with mp = '%v+'\n", string(line)[current:], current, rdx, mp)
// 	if !mp.matchByte(line[current]) {
// 		return false
// 	}
// 	// finding max length that will match and then working backwards
// 	maxLength := 1
// 	for ; maxLength+current < len(line); maxLength++ {
// 		if !mp.matchByte(line[maxLength+current]) {
// 			// keep incrementing until we fail to match or hit the end
// 			break
// 		}
// 	}
// 	fmt.Fprintf(os.Stderr, "maxLength: %d\n", maxLength)
// 	// here is the working backwards
// 	for trialLength := maxLength; trialLength > 0; trialLength-- {
// 		fmt.Fprintf(os.Stderr, "trialLength: %d\n", trialLength)
// 		if re.matchHere(line, current+trialLength, rdx+1) {
// 			return true
// 		}
// 	}
// 	return false
// }

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

	if pattern[len(pattern)-1] == '$' {
		fmt.Fprintf(os.Stderr, "matched $ at end: %s\n", pattern)
		regex.matchEnd = true
		pattern = pattern[:len(pattern)-1]
	}
	regex.mps = ParsePattern(pattern, 0)
	return regex
}

func parseSetPattern(inverted bool, pattern *string, index *int) (*matchPoint, error) {
	retval := matchPoint{}
	retval.inverted = inverted
	chars := []byte{}
	for *index < len(*pattern) {
		switch (*pattern)[*index] {
		case ']':
			retval.matchChars = string(chars[:])
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
	var p *matchPoint
	for index := start; index < len(pattern); {
		var err error
		switch pattern[index] {
		case '[':
			index++
			if index < len(pattern) && pattern[index] == '^' {
				index++
				p, err = parseSetPattern(true, &pattern, &index)
			} else {
				p, err = parseSetPattern(false, &pattern, &index)
			}
			if err != nil {
				os.Exit(3)
			}
		case '?':
			if len(regex) == 0 {
				// handle + at start of string I guess...
				fmt.Fprintf(os.Stderr, "questionmark at start %v+??\n", p)
				p = &matchPoint{matchChars: string(pattern[index])}
			} else {
				// set the last mp to be one or more...
				regex[len(regex)-1].setWildtype(zeroOrOne)
				index++
				continue // don't add a matchPoint
			}
		case '+':
			if len(regex) == 0 {
				// handle + at start of string I guess...
				fmt.Fprintf(os.Stderr, "plus at start %v+??\n", p)
				p = &matchPoint{matchChars: string(pattern[index])}
			} else {
				// set the last mp to be one or more...
				regex[len(regex)-1].setWildtype(oneOrMore)
				index++
				continue // don't add a matchPoint
			}
		case '*':
			if len(regex) == 0 {
				// handle * at start of string I guess...
				fmt.Fprintf(os.Stderr, "asterisk at start %v+??\n", p)
				p = &matchPoint{matchChars: string(pattern[index])}
			} else {
				// set the last mp to be one or more...
				regex[len(regex)-1].setWildtype(zeroOrMore)
				index++
				continue // don't add a matchPoint
			}
		case '.':
			p = &matchPoint{"", true, normal, nil}
		case '\\':
			index++
			if index < len(pattern) {
				switch pattern[index] {
				case 'w':
					p = &matchPoint{matchChars: wordChars}
				case 'd':
					p = &matchPoint{matchChars: digits}
				default:
					p = &matchPoint{matchChars: string(pattern[index])}
				}
			} else {
				// last character was a backslash....
				// I guess append a backslash character?
				p = &matchPoint{matchChars: "\\"}
			}
		default:
			p = &matchPoint{matchChars: string(pattern[index])}
		}
		regex = append(regex, p)
		index++
	}
	if len(regex) == 0 {
		// XXX should probably just bail if this is the case?
		return nil
	}
	for i := 0; i < len(regex)-1; i++ {
		fmt.Println(regex[i])
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
		matchesHere := re.matchHere(line, current, 0)
		if matchesHere {
			fmt.Fprintln(os.Stderr, "ml whole matched")
			return true
		}
	}

	fmt.Fprintln(os.Stderr, "ml whole fails")
	return false
}
