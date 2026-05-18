package quorum

import (
	"errors"
	"fmt"

	"disturbedcli/internal/worker"
)

var ErrQuorumNotReached = errors.New("quorum not reached: too many worker failures")

func Collect(resultsCh <-chan worker.Result, total, quorum int) ([]worker.Match, error) {
	all := make([]worker.Result, 0, total)
	failures := 0

	for range total {
		r := <-resultsCh
		if r.Err != nil {
			failures++
		}
		all = append(all, r)
	}

	successes := total - failures
	if successes < quorum {
		return nil, fmt.Errorf("%w: %d/%d succeeded, need %d",
			ErrQuorumNotReached, successes, total, quorum)
	}

	var merged []worker.Match
	for _, r := range all {
		if r.Err == nil {
			merged = append(merged, r.Matches...)
		}
	}

	sortMatches(merged)
	return merged, nil
}

func sortMatches(m []worker.Match) {
	for i := 1; i < len(m); i++ {
		key := m[i]
		j := i - 1
		for j >= 0 && m[j].LineNum > key.LineNum {
			m[j+1] = m[j]
			j--
		}
		m[j+1] = key
	}
}
