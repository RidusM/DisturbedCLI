package cli

import (
	"flag"
	"fmt"
	"os"
)

const (
	_exitUsage            = 2
	_defaultQuorumDivisor = 2
)

type Flags struct {
	Pattern     string
	Files       []string
	IgnoreCase  bool
	InvertMatch bool
	Count       bool
	LineNumber  bool
	OnlyMatch   bool
	Quiet       bool
	Workers     int
	Quorum      int
	Addr        string
	Peers       string
}

func Parse() *Flags {
	f := &Flags{}

	flag.BoolVar(&f.IgnoreCase, "i", false, "ignore case distinctions")
	flag.BoolVar(&f.InvertMatch, "v", false, "select non-matching lines")
	flag.BoolVar(&f.Count, "c", false, "print count of matching lines")
	flag.BoolVar(&f.LineNumber, "n", false, "prefix output with line number")
	flag.BoolVar(&f.OnlyMatch, "o", false, "print only matching part of the line")
	flag.BoolVar(&f.Quiet, "q", false, "quiet mode: exit 0 if any match, 1 otherwise")
	flag.IntVar(&f.Workers, "workers", 1, "number of parallel worker nodes")
	flag.IntVar(&f.Quorum, "quorum", 0, "quorum size (default: workers/2+1)")
	flag.StringVar(&f.Addr, "addr", "", "worker listen address (worker mode, e.g. :5001)")
	flag.StringVar(&f.Peers, "peers", "", "comma-separated worker addresses (coordinator mode)")

	flag.Usage = usage
	flag.Parse()

	args := flag.Args()

	if f.Addr != "" {
		return f
	}

	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "error: pattern required")
		flag.Usage()
		os.Exit(_exitUsage)
	}

	f.Pattern = args[0]
	f.Files = args[1:]

	if f.Workers < 1 {
		fmt.Fprintln(os.Stderr, "error: -workers must be ≥ 1")
		os.Exit(_exitUsage)
	}
	if f.Quorum == 0 {
		f.Quorum = f.Workers/_defaultQuorumDivisor + 1
	}
	if f.Quorum > f.Workers {
		fmt.Fprintf(os.Stderr, "error: -quorum (%d) cannot exceed -workers (%d)\n", f.Quorum, f.Workers)
		os.Exit(_exitUsage)
	}

	return f
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage:
  grep [flags] PATTERN [FILE...]        # coordinator / local mode
  grep -addr :5001                      # run as worker node

Flags:
`)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, `
Examples:
  grep -workers 4 "error" server.log
  grep -addr :5001
  grep -peers :5001,:5002,:5003 -quorum 2 "panic" app.log
`)
}
