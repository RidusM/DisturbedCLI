package benchmark_test

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"testing"

	"disturbedcli/internal/coordinator"
	"disturbedcli/internal/worker"
)

const benchLineCount = 500_000

func generateLines(n int) []string {
	words := []string{"INFO", "WARN", "DEBUG", "TRACE", "FATAL"}
	lines := make([]string, n)
	for i := range lines {
		if rand.Intn(10) == 0 {
			lines[i] = fmt.Sprintf("2026-01-01 ERROR service crashed at line %d", i)
		} else {
			lines[i] = fmt.Sprintf("2026-01-01 %s all good line %d", words[rand.Intn(len(words))], i)
		}
	}
	return lines
}

func TestMain(m *testing.M) {
	lines := generateLines(benchLineCount)
	benchLines = lines

	f, err := os.CreateTemp("", "grep-bench-*.txt")
	if err != nil {
		panic(err)
	}
	if _, err = f.WriteString(strings.Join(lines, "\n")); err != nil {
		panic(err)
	}
	if err = f.Close(); err != nil {
		panic(err)
	}
	benchFile = f.Name()

	code := m.Run()
	os.Remove(benchFile) 
	os.Exit(code)
}

var (
	benchLines []string 
	benchFile  string   
)

func BenchmarkLocal_1Worker(b *testing.B) {
	re, _ := worker.CompilePattern("ERROR", false)
	for range b.N {
		coordinator.RunLocal(benchLines, coordinator.LocalOptions{ 
			Workers: 1, Quorum: 1, Re: re,
		})
	}
}

func BenchmarkLocal_4Workers(b *testing.B) {
	re, _ := worker.CompilePattern("ERROR", false)
	for range b.N {
		coordinator.RunLocal(benchLines, coordinator.LocalOptions{ 
			Workers: 4, Quorum: 3, Re: re,
		})
	}
}

func BenchmarkLocal_8Workers(b *testing.B) {
	re, _ := worker.CompilePattern("ERROR", false)
	for range b.N {
		coordinator.RunLocal(benchLines, coordinator.LocalOptions{ 
			Workers: 8, Quorum: 5, Re: re,
		})
	}
}

func BenchmarkSystemGrep(b *testing.B) {
	grepPath, err := exec.LookPath("grep")
	if err != nil {
		b.Skip("grep not found in PATH")
	}
	for range b.N {
		cmd := exec.Command(grepPath, "ERROR", benchFile)
		cmd.Stdout = nil
		cmd.Run() 
	}
}
