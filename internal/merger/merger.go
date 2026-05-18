package merger

import (
	"fmt"
	"io"

	"disturbedcli/internal/worker"
)

type Options struct {
	LineNumber bool
	OnlyMatch  bool
	Count      bool
	Quiet      bool
	Filename   string
}

func Print(w io.Writer, matches []worker.Match, opts Options) (int, bool) {
	if opts.Quiet {
		return len(matches), len(matches) > 0
	}

	if opts.Count {
		if opts.Filename != "" {
			fmt.Fprintf(w, "%s:%d\n", opts.Filename, len(matches))
		} else {
			fmt.Fprintf(w, "%d\n", len(matches))
		}
		return len(matches), len(matches) > 0
	}

	for _, m := range matches {
		var prefix string
		if opts.Filename != "" {
			prefix += opts.Filename + ":"
		}
		if opts.LineNumber {
			prefix += fmt.Sprintf("%d:", m.LineNum)
		}

		text := m.Line
		if opts.OnlyMatch {
			text = m.Excerpt
		}

		fmt.Fprintf(w, "%s%s\n", prefix, text)
	}

	return len(matches), len(matches) > 0
}
