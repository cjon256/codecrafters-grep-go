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
}

type matchPoint struct {
	matchChars string
	inverted   bool
}

func (re MyRegExp) matchHere(line []byte, current int) bool {
	for i := 0; i < len(re.mps); i++ {
		if current >= len(line) {
			fmt.Fprintln(os.Stderr, "oops, got to long")
			return false
		}
		if !re.mps[i].matchOnce(line, current) {
			fmt.Fprintln(os.Stderr, "match fails in regExp::matchHere")
			return false
		}
		current++
	}
	return true
}

func (mp *matchPoint) matchOnce(line []byte, current int) bool {
	matches := strings.Contains(mp.matchChars, string((line)[current]))
	fmt.Fprintf(os.Stderr, "matchHere(line='%s', current=%d) with mp = '%v+'\n", string(line), current, mp)
	if mp.inverted {
		matches = !matches
	}
	if matches {
		// fmt.Fprintln("matches")
		return true
	} else {
		// fmt.Fprintln("fails")
		return false
	}
}

func parseSetPattern(inverted bool, patternIn *string, index *int) (matchPoint, error) {
	retval := matchPoint{}
	retval.inverted = inverted
	chars := []byte{}
	for *index < len(*patternIn) {
		switch (*patternIn)[*index] {
		case ']':
			retval.matchChars = string(chars[:])
			return retval, nil
		default:
			chars = append(chars, (*patternIn)[*index])
		}
		(*index)++
	}
	return retval, errors.New("parse pattern not closed")
}

func ParsePattern(patternIn string) (MyRegExp, error) {
	index := 0
	regex := MyRegExp{}
	if patternIn[index] == '^' {
		regex.matchStart = true
		index++
	}

	for index < len(patternIn) {
		var err error
		var p matchPoint
		switch patternIn[index] {
		case '[':
			index++
			if index < len(patternIn) && patternIn[index] == '^' {
				index++
				p, err = parseSetPattern(true, &patternIn, &index)
			} else {
				p, err = parseSetPattern(false, &patternIn, &index)
			}
			if err != nil {
				os.Exit(3)
			}
		case '\\':
			index++
			if index < len(patternIn) {
				switch patternIn[index] {
				case 'w':
					p = matchPoint{matchChars: wordChars}
				case 'd':
					p = matchPoint{matchChars: digits}
				default:
					p = matchPoint{matchChars: string(patternIn[index])}
				}
			} else {
				// last character was a backslash....
				// I guess append a backslash character?
				p = matchPoint{matchChars: "\\"}
			}
		default:
			p = matchPoint{matchChars: string(patternIn[index])}
		}
		regex.mps = append(regex.mps, p)
		index++
	}
	return regex, nil
}

func (re *MyRegExp) MatchLine(line []byte) (bool, error) {
	limit := len(line)
	if re.matchStart {
		limit = 1
	}
	for current := 0; current < limit; current++ {
		matchesHere := re.matchHere(line, current)
		if matchesHere {
			fmt.Fprintln(os.Stderr, "ml whole matched")
			return true, nil
		}
	}

	fmt.Fprintln(os.Stderr, "ml whole fails")
	return false, nil
}
