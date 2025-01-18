package regexp

import (
	"fmt"
	"os"
)

// DEBUG flag, set to true to enable debug mode
const DEBUG bool = false

// debugf convenience function
func debugf(format string, args ...interface{}) {
	if DEBUG {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}
