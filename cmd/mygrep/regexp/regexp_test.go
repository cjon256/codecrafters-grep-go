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
	{
		name:     "dot_t",
		line:     "cat",
		pattern:  "c.t",
		expected: true,
	},
	{
		name:     "dot_f",
		line:     "car",
		pattern:  "c.t",
		expected: false,
	},
	{
		name:     "dotplus_t",
		line:     "goøö0Ogol",
		pattern:  "g.+gol",
		expected: true,
	},
	{
		name:     "questionmark_t",
		line:     "act",
		pattern:  "ca?t",
		expected: true,
	},
	{
		name:     "questionmark_f",
		line:     "dog",
		pattern:  "ca?t",
		expected: false,
	},
	{
		name:     "questionmark_f",
		line:     "cag",
		pattern:  "ca?t",
		expected: false,
	},
	{
		name:     "plus_t",
		line:     "caaats",
		pattern:  "ca+t",
		expected: true,
	},
	{
		name:     "plus_t",
		line:     "cat",
		pattern:  "ca+t",
		expected: true,
	},
	{
		name:     "plus_f",
		line:     "act",
		pattern:  "ca+t",
		expected: false,
	},
	{
		name:     "plus_f",
		line:     "ca",
		pattern:  "ca+t",
		expected: false,
	},
	{
		name:     "dollarsign_t",
		line:     "cat",
		pattern:  "cat$",
		expected: true,
	},
	{
		name:     "dollarsign_t",
		line:     "cats",
		pattern:  "cat$",
		expected: false,
	},
	{
		name:     "caret_t",
		line:     "log",
		pattern:  "^log",
		expected: true,
	},
	{
		name:     "caret_f",
		line:     "slog",
		pattern:  "^log",
		expected: false,
	},
	{
		name:     "digits_t",
		line:     "sally has 3 apples",
		pattern:  "\\d apple",
		expected: true,
	},
	{
		name:     "digits_f",
		line:     "sally has 1 orange",
		pattern:  "\\d apple",
		expected: false,
	},
	{
		name:     "digits_t",
		line:     "sally has 124 apples",
		pattern:  "\\d\\d\\d apples",
		expected: true,
	},
	{
		name:     "digits_f",
		line:     "sally has 12 apples",
		pattern:  "\\d\\\\d\\\\d apples",
		expected: false,
	},
	{
		name:     "alphanum_t",
		line:     "sally has 3 dogs",
		pattern:  "\\d \\w\\w\\ws",
		expected: true,
	},
	{
		name:     "alphanum_t",
		line:     "sally has 4 dogs",
		pattern:  "\\d \\w\\w\\ws",
		expected: true,
	},
	{
		name:     "alphanum_f",
		line:     "sally has 1 dog",
		pattern:  "\\d \\w\\w\\ws",
		expected: false,
	},
	{
		name:     "brackets_t",
		line:     "apple",
		pattern:  "[^xyz]",
		expected: true,
	},
	{
		name:     "brackets_f",
		line:     "banana",
		pattern:  "[^anb]",
		expected: false,
	},
	{
		name:     "brackets_t",
		line:     "a",
		pattern:  "[abcd]",
		expected: true,
	},
	{
		name:     "brackets_f",
		line:     "efgh",
		pattern:  "[abcd]",
		expected: false,
	},
	{
		name:     "alphanum_t",
		line:     "word",
		pattern:  "\\w",
		expected: true,
	},
	{
		name:     "alphanum_f",
		line:     "$!?",
		pattern:  "\\w",
		expected: false,
	},
	{
		name:     "digits_t",
		line:     "123",
		pattern:  "\\d",
		expected: true,
	},
	{
		name:     "digits_f",
		line:     "apple",
		pattern:  "\\d",
		expected: false,
	},
	{
		name:     "simple_t",
		line:     "dog",
		pattern:  "d",
		expected: true,
	},
	{
		name:     "simple_f",
		line:     "dog",
		pattern:  "f",
		expected: false,
	},

	{
		name:     "plus_t",
		line:     "cart",
		pattern:  "car+t",
		expected: true,
	},
	{
		name:     "plus_t",
		line:     "carrrrrt",
		pattern:  "car+t",
		expected: true,
	},
	{
		name:     "plus_f",
		line:     "cat",
		pattern:  "car+t",
		expected: false,
	},
	{
		name:     "star_t",
		line:     "cart",
		pattern:  "car*t",
		expected: true,
	},
	{
		name:     "star_t",
		line:     "carrrrrt",
		pattern:  "car*t",
		expected: true,
	},
	{
		name:     "star_t",
		line:     "cat",
		pattern:  "car*t",
		expected: true,
	},
	{
		name:     "star_f",
		line:     "caat",
		pattern:  "car*t",
		expected: false,
	},
	{
		name:     "parens_t",
		line:     "caat",
		pattern:  "ca(a)t",
		expected: true,
	},
	{
		name:     "parens_t",
		line:     "caat",
		pattern:  "c(aa)t",
		expected: true,
	},
	{
		name:     "alter_t",
		line:     "caat",
		pattern:  "c(ol|aa)t",
		expected: true,
	},
	{
		name:     "alter_t",
		line:     "colt",
		pattern:  "c(ol|aa)t",
		expected: true,
	},
	{
		name:     "alter_f",
		line:     "cola",
		pattern:  "c(ol|aa)t",
		expected: false,
	},
	{
		name:     "alter_f",
		line:     "cont",
		pattern:  "c(ol|aa)t",
		expected: false,
	},
	// {
	// 	name:     "backref_err",
	// 	line:     "the cat is a cat",
	// 	pattern:  "the cat is a \\1",
	// 	expected: error, // there is no group
	// },
	{
		name:     "backref_t",
		line:     "the cat is a cat",
		pattern:  "the (cat) is a \\1",
		expected: true,
	},
	{
		name:     "backref_f",
		line:     "the cat is a cutey",
		pattern:  "the (cat) is a \\1",
		expected: false,
	},
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
