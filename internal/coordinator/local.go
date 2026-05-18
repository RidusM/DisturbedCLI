package coordinator

import (
	"fmt"
	"regexp"

	"disturbedcli/internal/quorum"
	"disturbedcli/internal/splitter"
	"disturbedcli/internal/worker"
)

type LocalOptions struct {
	Workers     int
	Quorum      int
	Re          *regexp.Regexp
	IgnoreCase  bool
	InvertMatch bool
	OnlyMatch   bool
}

func RunLocal(lines []string, opts LocalOptions) ([]worker.Match, error) {
	chunks := splitter.Split(lines, opts.Workers)

	resultsCh := make(chan worker.Result, len(chunks))

	for _, chunk := range chunks {
		go func(ch splitter.Chunk) {
			resultsCh <- worker.Run(ch, opts.Re, worker.Options{
				IgnoreCase:  opts.IgnoreCase,
				InvertMatch: opts.InvertMatch,
				OnlyMatch:   opts.OnlyMatch,
			})
		}(chunk)
	}

	matches, err := quorum.Collect(resultsCh, len(chunks), opts.Quorum)
	if err != nil {
		return nil, fmt.Errorf("local run: %w", err)
	}
	return matches, nil
}
