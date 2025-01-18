package regexp

import (
	"fmt"
)

type RegExp struct {
	mps        matchPoint
	matchStart bool
}

func (re RegExp) String() string {
	return fmt.Sprintf("#[RegExp: '%s' %v]", re.mps, re.matchStart)
}

func (re *RegExp) MatchLine(line []byte) bool {
	limit := len(line)
	if re.matchStart {
		limit = 1
	}
	for ldx := 0; ldx < limit; ldx++ {
		matchesHere := re.mps.matchHere(line, ldx)
		if matchesHere {
			debugf("ml whole matched\n")
			return true
		}
	}

	debugf("ml whole fails\n")
	return false
}
