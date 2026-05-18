package worker

import (
	"fmt"
	"regexp"
	"strings"

	"disturbedcli/internal/splitter"
)

type Match struct {
	LineNum int
	Line    string
	Excerpt string
}

type Result struct {
	Matches []Match
	Count   int
	Err     error
}

type Options struct {
	IgnoreCase  bool
	InvertMatch bool
	OnlyMatch   bool
	Count       bool
}

func Run(chunk splitter.Chunk, re *regexp.Regexp, opts Options) Result {
	var matches []Match

	for i, line := range chunk.Lines {
		lineNum := chunk.StartLine + i
		matched, excerpt := matchLine(line, re, opts)
		if matched {
			matches = append(matches, Match{
				LineNum: lineNum,
				Line:    line,
				Excerpt: excerpt,
			})
		}
	}

	return Result{
		Matches: matches,
		Count:   len(matches),
	}
}

func matchLine(line string, re *regexp.Regexp, opts Options) (bool, string) {
	target := line
	if opts.IgnoreCase {
		target = strings.ToLower(line)
	}

	loc := re.FindStringIndex(target)
	matched := loc != nil

	if opts.InvertMatch {
		return !matched, ""
	}

	if !matched {
		return false, ""
	}

	return true, line[loc[0]:loc[1]]
}

func CompilePattern(pattern string, ignoreCase bool) (*regexp.Regexp, error) {
	if ignoreCase {
		pattern = strings.ToLower(pattern)
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("compile pattern %q: %w", pattern, err)
	}
	return re, nil
}
