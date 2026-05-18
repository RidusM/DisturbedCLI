package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"disturbedcli/internal/cli"
	"disturbedcli/internal/coordinator"
	"disturbedcli/internal/merger"
	"disturbedcli/internal/worker"
)

const (
	_exitNoMatch = 1
	_exitError   = 2
)

func main() {
	flags := cli.Parse()

	if flags.Addr != "" {
		if err := coordinator.ServeWorker(flags.Addr); err != nil {
			fmt.Fprintf(os.Stderr, "worker error: %v\n", err)
			os.Exit(_exitError)
		}
		return
	}

	re, err := worker.CompilePattern(flags.Pattern, flags.IgnoreCase)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid pattern %q: %v\n", flags.Pattern, err)
		os.Exit(_exitError)
	}

	fileLines, readErr := readInput(flags)
	if readErr != nil {
		fmt.Fprintf(os.Stderr, "warning: %v\n", readErr)
	}

	peers := parsePeers(flags.Peers)
	totalMatches, anyMatch := processFiles(fileLines, flags, re, peers)

	if flags.Quiet {
		if anyMatch {
			os.Exit(0)
		}
		os.Exit(_exitNoMatch)
	}
	if totalMatches == 0 {
		os.Exit(_exitNoMatch)
	}
}

func readInput(flags *cli.Flags) (map[string][]string, error) {
	if len(flags.Files) == 0 {
		lines, err := merger.ReadLines(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("read stdin: %w", err)
		}
		return map[string][]string{"": lines}, nil
	}
	return merger.ReadFiles(flags.Files)
}

func processFiles(
	fileLines map[string][]string,
	flags *cli.Flags,
	re *regexp.Regexp,
	peers []string,
) (totalMatches int, anyMatch bool) {
	for filename, lines := range fileLines {
		matches, err := runGrep(lines, flags, re, peers)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(_exitError)
		}

		displayName := ""
		if len(flags.Files) > 1 {
			displayName = filename
		}

		n, found := merger.Print(os.Stdout, matches, merger.Options{
			LineNumber: flags.LineNumber,
			OnlyMatch:  flags.OnlyMatch,
			Count:      flags.Count,
			Quiet:      flags.Quiet,
			Filename:   displayName,
		})
		totalMatches += n
		if found {
			anyMatch = true
		}
	}
	return totalMatches, anyMatch
}

func runGrep(
	lines []string,
	flags *cli.Flags,
	re *regexp.Regexp,
	peers []string,
) ([]worker.Match, error) {
	if len(peers) > 0 {
		return coordinator.RunDistributed(lines, coordinator.RemoteOptions{
			Peers:       peers,
			Quorum:      flags.Quorum,
			Pattern:     flags.Pattern,
			IgnoreCase:  flags.IgnoreCase,
			InvertMatch: flags.InvertMatch,
			OnlyMatch:   flags.OnlyMatch,
		})
	}

	return coordinator.RunLocal(lines, coordinator.LocalOptions{
		Workers:     flags.Workers,
		Quorum:      flags.Quorum,
		Re:          re,
		IgnoreCase:  flags.IgnoreCase,
		InvertMatch: flags.InvertMatch,
		OnlyMatch:   flags.OnlyMatch,
	})
}

func parsePeers(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
