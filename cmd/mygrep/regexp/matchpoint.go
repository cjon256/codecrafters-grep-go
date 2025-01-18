package regexp

import (
	"fmt"
	"os"
	"strings"
)

type matchPoint interface {
	fmt.Stringer
	matchHere(line []byte, ldx int) bool
	setNext(matchPoint)
}

type basicMatchPoint struct {
	MatchChars string
	Inverted   bool
	Next       matchPoint
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
	_ matchPoint = matchEndMatchPoint{}
)

func (mp basicMatchPoint) recursiveString(mytype string) string {
	var remainder string
	if mp.Next == nil {
		remainder = "nil"
	} else {
		remainder = mp.Next.String()
	}
	return fmt.Sprintf("#[%s: '%s' %v %s]", mytype, mp.MatchChars, mp.Inverted, remainder)
}

func (mp basicMatchPoint) String() string {
	return mp.recursiveString("basicMatchPoint")
}

func (mp oneOrMoreMatchPoint) String() string {
	return mp.recursiveString("oneOrMoreMatchPoint")
}

func (mp zeroOrMoreMatchPoint) String() string {
	return mp.recursiveString("zeroOrMoreMatchPoint")
}

func (mp zeroOrOneMatchPoint) String() string {
	return mp.recursiveString("zeroOrOneMatchPoint")
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

func (mp basicMatchPoint) matchHere(line []byte, ldx int) bool {
	debugf("mp=%#v\n", mp)
	debugf("basicMatchPoint.matchHere('%s', %d)\n", string(line)[ldx:], ldx)
	if ldx >= len(line) {
		debugf("oops, got to long\n")
		return false
	}
	if !mp.matchByte(line[ldx]) {
		debugf("no match\n")
		return false
	}
	if mp.Next == nil {
		debugf("finished matching\n")
		return true
	}
	return mp.Next.matchHere(line, ldx+1)
}

func (mp zeroOrOneMatchPoint) matchHere(line []byte, ldx int) bool {
	debugf("mp=%#v\n", mp)
	debugf("zeroOrOneMatchPoint.matchHere('%s', %d)\n", string(line)[ldx:], ldx)
	if mp.Next == nil {
		debugf("short circuit match\n")
		return true
	}
	if ldx >= len(line) {
		debugf("at end, so trying zero length\n")
		return mp.Next.matchHere(line, ldx)
	}
	if !mp.matchByte(line[ldx]) {
		debugf("no match, so trying zero length\n")
		return mp.Next.matchHere(line, ldx)
	}
	return mp.Next.matchHere(line, ldx+1)
}

func (mp zeroOrMoreMatchPoint) matchHere(line []byte, ldx int) bool {
	debugf("mp=%#v\n", mp)
	debugf("zeroOrMoreMatchPoint.matchHere('%s', %d)\n", string(line)[ldx:], ldx)
	if mp.Next == nil {
		debugf("short circuit match\n")
		return true
	}
	if ldx >= len(line) {
		debugf("at end, so trying zero length\n")
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
	debugf("maxLength: %d\n", maxLength)
	// here is the working backwards
	for trialLength := maxLength; trialLength >= 0; trialLength-- {
		debugf("trialLength: %d\n", trialLength)
		if mp.Next.matchHere(line, ldx+trialLength) {
			return true
		}
	}
	return false
}

func (mp oneOrMoreMatchPoint) matchHere(line []byte, ldx int) bool {
	debugf("mp=%#v\n", mp)
	debugf("zeroOrMoreMatchPoint.matchHere('%s', %d)\n", string(line)[ldx:], ldx)
	if mp.Next == nil {
		debugf("short circuit match\n")
		return true
	}
	if ldx >= len(line) {
		debugf("at end, so trying zero length\n")
		return mp.Next.matchHere(line, ldx)
	}
	// need at least one
	if !mp.matchByte(line[ldx]) {
		debugf("no match\n")
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
	debugf("maxLength: %d\n", maxLength)
	// here is the working backwards
	for trialLength := maxLength; trialLength >= 0; trialLength-- {
		debugf("trialLength: %d\n", trialLength)
		if mp.Next.matchHere(line, ldx+trialLength) {
			return true
		}
	}
	return false
}

func (e matchEndMatchPoint) matchHere(line []byte, ldx int) bool {
	debugf("mp=%#v\n", e)
	debugf("matchEndMatchPoint.matchHere('%s', %d)\n", string(line)[ldx:], ldx)
	atEnd := ldx == len(line)
	if atEnd {
		debugf("Matched at end and regexp has $\n")
		return true
	}
	debugf("Not at end and regexp has $\n")
	return false
}

func (mp *basicMatchPoint) setNext(n matchPoint) {
	mp.Next = n
}

func (e matchEndMatchPoint) setNext(_ matchPoint) {
	// should exit as this would be a clear error?
	os.Exit(3)
}
