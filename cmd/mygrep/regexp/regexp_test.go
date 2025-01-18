package regexp

import "testing"

func RegexTester(lineStr string, pattern string) bool {
	line := []byte(lineStr)
	regex := ParseRegExp(pattern)
	return regex.MatchLine(line)
}

type RegexInput struct {
	name     string
	line     string
	pattern  string
	expected bool
}

var tests = []RegexInput{
	{"dot_t", "cat", "c.t", true},
	{"dot_f", "car", "c.t", false},
	{"dotplus_t", "goøö0Ogol", "g.+gol", true},
	{"questionmark_t", "act", "ca?t", true},
	{"questionmark_f", "dog", "ca?t", false},
	{"questionmark_f", "cag", "ca?t", false},
	{"plus_t", "caaats", "ca+t", true},
	{"plus_t", "cat", "ca+t", true},
	{"plus_f", "act", "ca+t", false},
	{"plus_f", "ca", "ca+t", false},
	{"dollarsign_t", "cat", "cat$", true},
	{"dollarsign_t", "cats", "cat$", false},
	{"caret_t", "log", "^log", true},
	{"caret_f", "slog", "^log", false},
	{"digits_t", "sally has 3 apples", "\\d apple", true},
	{"digits_f", "sally has 1 orange", "\\d apple", false},
	{"digits_t", "sally has 124 apples", "\\d\\d\\d apples", true},
	{"digits_f", "sally has 12 apples", "\\d\\\\d\\\\d apples", false},
	{"alphanum_t", "sally has 3 dogs", "\\d \\w\\w\\ws", true},
	{"alphanum_t", "sally has 4 dogs", "\\d \\w\\w\\ws", true},
	{"alphanum_f", "sally has 1 dog", "\\d \\w\\w\\ws", false},
	{"brackets_t", "apple", "[^xyz]", true},
	{"brackets_f", "banana", "[^anb]", false},
	{"brackets_t", "a", "[abcd]", true},
	{"brackets_f", "efgh", "[abcd]", false},
	{"alphanum_t", "word", "\\w", true},
	{"alphanum_f", "$!?", "\\w", false},
	{"digits_t", "123", "\\d", true},
	{"digits_f", "apple", "\\d", false},
	{"simple_t", "dog", "d", true},
	{"simple_f", "dog", "f", false},
}

func TestRegexTableDriven(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RegexTester(tt.line, tt.pattern); got != tt.expected {
				t.Errorf("%s ~ /%s/ = %v; want %v", tt.line, tt.pattern, got, tt.expected)
			}
		})
	}
}
