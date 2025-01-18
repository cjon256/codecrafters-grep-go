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
	{"gol", "gol", "gol", true},
	{"gol", "gol", "gorl", false},
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

var _ = `
		{"gol", "gol", "gol", true},
		{"gol", "gol", "gorl", false},
`
