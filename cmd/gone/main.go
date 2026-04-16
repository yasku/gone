package main

import (
	"fmt"
	"os"
	"time"

	"gone/internal/scanner"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: gone <search-term>")
		os.Exit(1)
	}
	term := os.Args[1]
	fmt.Printf("Scanning for %q...\n", term)

	start := time.Now()
	results, err := scanner.Search(term, scanner.ScanPaths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	elapsed := time.Since(start)

	for _, m := range results {
		fmt.Printf("  [%s] %s  (%d bytes, %s)\n", m.Kind, m.Path, m.Size, m.ModTime.Format("2006-01-02"))
	}

	// After file results loop:
	rcMatches := scanner.SearchRC(term)
	for _, rc := range rcMatches {
		fmt.Printf("  [rc-line] %s:%d  %s\n", rc.File, rc.LineNum, rc.Line)
	}
	fmt.Printf("\n%d files + %d rc lines in %s\n", len(results), len(rcMatches), elapsed)
}
