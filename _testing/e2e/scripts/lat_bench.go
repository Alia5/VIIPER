package main

// lat_bench.go
// Utility to run (or parse) Xbox360E2E benchmarks and emit enriched latency tables.
// Supports markdown, plain table and JSON output.
// IMPORTANT: The underlying benchmark MUST NOT run in parallel; the benchmark
// itself calls b.SetParallelism(1). This tool does not force parallel execution.
//
// Usage examples:
//   # Run benchmarks (count=5) and emit markdown
//   go run ./_testing/e2e/scripts/lat_bench.go -format markdown -count 5 > latency.md
//
//   # Parse existing benchmark output instead of running (offline mode)
//   go run ./_testing/e2e/scripts/lat_bench.go -format table -input bench.txt
//
//   # Produce JSON for CI consumption
//   go run ./_testing/e2e/scripts/lat_bench.go -format json -count 3
//
//   # Fixed iteration benchtime (5000 operations per sub benchmark)
//   go run ./_testing/e2e/scripts/lat_bench.go -benchtime 5000x -format markdown > latency.md
//
//   # Time based benchtime (2 seconds per benchmark)
//   go run ./_testing/e2e/scripts/lat_bench.go -benchtime 2s -format table
//
//   # Default benchtime when not specified is 1000x (fixed iterations)
//   go run ./_testing/e2e/scripts/lat_bench.go -format table              # implicit -benchtime=1000x
//
//   # Filter benchmarks by encryption status
//   go run ./_testing/e2e/scripts/lat_bench.go -encryption plain          # only unencrypted (default)
//   go run ./_testing/e2e/scripts/lat_bench.go -encryption encrypted       # only encrypted
//   go run ./_testing/e2e/scripts/lat_bench.go -encryption both            # all benchmarks
//
// The tool always:
//   * Groups repeated benchmark cycles when count > 1
//   * Omits memory statistics (B/op, allocs/op)
//   * Uses E2E-InputDelay as 100% baseline for %Full column
//
// If running benchmarks on systems without the required gamepad/server setup
// you can capture output elsewhere and use -input parsing locally.

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/tabwriter"
	"time"
)

var (
	flagFormat     = flag.String("format", "table", "Output format: markdown|table|json")
	flagCount      = flag.Int("count", 1, "Number of benchmark runs when executing go test")
	flagInput      = flag.String("input", "", "Optional file path with pre-recorded benchmark output to parse instead of running")
	flagOutFile    = flag.String("out", "", "Optional output file path. If empty prints to stdout")
	flagBenchtime  = flag.String("benchtime", "", "Optional benchtime argument passed to 'go test' (e.g. 2s or 5000x, defaults to 1000x)")
	flagTestFlags  = flag.String("testflags", "", "Arbitrary additional flags passed verbatim to 'go test' (e.g. -testflags='-benchtime=5000x -timeout=120s'). Overrides -benchtime if it includes a benchtime.")
	flagPkg        = flag.String("pkg", ".", "Package path passed to 'go test'. Default '.' (current directory).")
	flagEncryption = flag.String("encryption", "plain", "Filter benchmarks by encryption: plain (default, unencrypted only), encrypted (encrypted only), or both (no filtering)")
)

func main() {
	flag.Parse()
	var raw string
	if *flagInput != "" {
		data, err := os.ReadFile(*flagInput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read input file: %v\n", err)
			os.Exit(1)
		}
		raw = string(data)
	} else {
		var err error
		raw, err = runBench(context.Background(), *flagPkg, *flagCount)
		if err != nil {
			fmt.Fprintf(os.Stderr, "benchmark execution error: %v\n", err)
			os.Exit(1)
		}
	}
	lines, err := parseLines(raw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(1)
	}
	lines = filterByEncryption(lines, *flagEncryption)
	runs := groupRuns(lines)
	td := tableData{Timestamp: time.Now(), Count: *flagCount}
	for i, r := range runs {
		metrics, notes := deriveRun(r)
		td.Runs = append(td.Runs, runData{Index: i, Lines: metrics, Notes: notes})
	}
	if out, err := exec.Command("go", "version").Output(); err == nil {
		td.GoVersion = strings.TrimSpace(string(out))
	}
	var outStr string
	switch strings.ToLower(*flagFormat) {
	case "markdown", "md":
		outStr = outputMarkdown(td)
	case "table":
		outStr = outputTable(td)
	case "json":
		js, err := json.MarshalIndent(td, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "json marshal error: %v\n", err)
			os.Exit(1)
		}
		outStr = string(js)
	default:
		fmt.Fprintf(os.Stderr, "unknown format: %s\n", *flagFormat)
		os.Exit(1)
	}
	if *flagOutFile != "" {
		if werr := os.WriteFile(*flagOutFile, []byte(outStr), 0644); werr != nil {
			fmt.Fprintf(os.Stderr, "failed to write output: %v\n", werr)
			os.Exit(1)
		}
	} else {
		fmt.Print(outStr)
	}
}

type benchLine struct {
	Name       string  `json:"name"`
	BaseName   string  `json:"base_name"`
	Threads    int     `json:"threads"`
	Iterations int     `json:"iterations"`
	NsPerOp    float64 `json:"ns_per_op"`
}

type derivedMetrics struct {
	benchLine
	PercentOfFull float64 `json:"percent_of_full"`
	ClientShare   float64 `json:"client_share_pct"`
	LatencyShare  float64 `json:"latency_share_pct"`
}

type runData struct {
	Index int              `json:"index"`
	Lines []derivedMetrics `json:"lines"`
	Notes []string         `json:"notes"`
}

type tableData struct {
	Timestamp time.Time `json:"timestamp"`
	GoVersion string    `json:"go_version"`
	Count     int       `json:"run_count"`
	Runs      []runData `json:"runs"`
}

var benchRegexp = regexp.MustCompile(
	`^Benchmark([^\s]+(?:/[^\s]+)*)-(\d+)\s+(\d+)\s+(\d+) ns/op\s+(\d+) B/op\s+(\d+) allocs/op$`,
)

func parseLines(in string) ([]benchLine, error) {
	var results []benchLine
	scanner := bufio.NewScanner(strings.NewReader(in))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		m := benchRegexp.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		parts := strings.Split(m[1], "/")
		bl := benchLine{
			Name:     m[1],
			BaseName: parts[len(parts)-1],
		}
		fmt.Sscanf(m[2], "%d", &bl.Threads)
		fmt.Sscanf(m[3], "%d", &bl.Iterations)
		fmt.Sscanf(m[4], "%f", &bl.NsPerOp)
		results = append(results, bl)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errors.New("no benchmark lines parsed – ensure benchmark ran successfully")
	}
	return results, nil
}

func runBench(ctx context.Context, pkg string, count int) (string, error) {
	args := []string{"test", "-bench=.", "-run", "NONE", "-benchmem", fmt.Sprintf("-count=%d", count)}
	if *flagTestFlags != "" {
		for _, f := range strings.Fields(*flagTestFlags) {
			args = append(args, f)
		}
	} else if *flagBenchtime != "" {
		args = append(args, fmt.Sprintf("-benchtime=%s", *flagBenchtime))
	} else {
		args = append(args, "-benchtime=1000x")
	}
	args = append(args, pkg)
	cmd := exec.CommandContext(ctx, "go", args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return buf.String(), fmt.Errorf("go test failed: %w\nOutput:\n%s", err, buf.String())
	}
	return buf.String(), nil
}

func deriveRun(lines []benchLine) (out []derivedMetrics, notes []string) {
	var client, delay, e2e, press *benchLine
	for i := range lines {
		lb := strings.ToLower(lines[i].BaseName)
		if client == nil && strings.Contains(lb, "client") && strings.Contains(lb, "write") {
			client = &lines[i]
		}
		if delay == nil && strings.Contains(lb, "without-client") {
			delay = &lines[i]
		}
		if e2e == nil && strings.Contains(lb, "e2e") && strings.Contains(lb, "inputdelay") {
			e2e = &lines[i]
		}
		if press == nil && strings.Contains(lb, "e2e") && strings.Contains(lb, "press") {
			press = &lines[i]
		}
	}
	full := 0.0
	if e2e != nil {
		full = e2e.NsPerOp
	} else {
		for i := range lines {
			if lines[i].NsPerOp > full {
				full = lines[i].NsPerOp
			}
		}
	}
	for _, l := range lines {
		dm := derivedMetrics{benchLine: l}
		if full > 0 {
			dm.PercentOfFull = l.NsPerOp / full * 100.0
		}
		if client != nil && l.BaseName == client.BaseName {
			dm.ClientShare = 100.0
			dm.LatencyShare = 0.0
		} else if delay != nil && l.BaseName == delay.BaseName {
			dm.ClientShare = 0.0
			dm.LatencyShare = 100.0
		} else if e2e != nil && client != nil && l.BaseName == e2e.BaseName {
			dm.ClientShare = client.NsPerOp / e2e.NsPerOp * 100.0
			dm.LatencyShare = (e2e.NsPerOp - client.NsPerOp) / e2e.NsPerOp * 100.0
		} else if press != nil && client != nil && l.BaseName == press.BaseName {
			clientTotal := 2 * client.NsPerOp
			dm.ClientShare = clientTotal / press.NsPerOp * 100.0
			dm.LatencyShare = (press.NsPerOp - clientTotal) / press.NsPerOp * 100.0
		}
		out = append(out, dm)
	}
	missing := []string{}
	if client == nil {
		missing = append(missing, "client-write")
	}
	if delay == nil {
		missing = append(missing, "delay-without-client")
	}
	if e2e == nil {
		missing = append(missing, "e2e-inputdelay")
	}
	if press == nil {
		missing = append(missing, "e2e-pressandrelease")
	}
	if len(missing) > 0 {
		notes = append(notes, "Missing roles: "+strings.Join(missing, ", "))
	}
	return
}

func filterByEncryption(lines []benchLine, mode string) []benchLine {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "both" || mode == "all" {
		return lines
	}

	wantEncrypted := mode == "encrypted" || mode == "enc"
	filtered := []benchLine{}
	for _, line := range lines {
		// Check if benchmark name contains (ENC) or (PLAIN)
		isEncrypted := strings.Contains(strings.ToUpper(line.Name), "(ENC)")
		isPlain := strings.Contains(strings.ToUpper(line.Name), "(PLAIN)")

		if wantEncrypted && isEncrypted {
			filtered = append(filtered, line)
		} else if !wantEncrypted && isPlain {
			filtered = append(filtered, line)
		} else if !isEncrypted && !isPlain {
			// If no encryption marker, include in plain mode by default
			if !wantEncrypted {
				filtered = append(filtered, line)
			}
		}
	}
	return filtered
}

func groupRuns(lines []benchLine) [][]benchLine {
	if len(lines) == 0 {
		return nil
	}
	occ := make(map[string][]benchLine)
	order := []string{}
	for _, l := range lines {
		if _, ok := occ[l.BaseName]; !ok {
			order = append(order, l.BaseName)
		}
		occ[l.BaseName] = append(occ[l.BaseName], l)
	}
	minCount := len(occ[order[0]])
	for _, name := range order[1:] {
		if c := len(occ[name]); c < minCount {
			minCount = c
		}
	}
	runs := make([][]benchLine, 0, minCount)
	for i := 0; i < minCount; i++ {
		var run []benchLine
		for _, name := range order {
			if i < len(occ[name]) {
				run = append(run, occ[name][i])
			}
		}
		runs = append(runs, run)
	}
	return runs
}

func outputMarkdown(td tableData) string {
	var b strings.Builder
	actualRuns := len(td.Runs)
	if actualRuns == 1 {
		b.WriteString(fmt.Sprintf("_Run count: %d (requested: %d)_\n\n", actualRuns, td.Count))
	} else {
		b.WriteString(fmt.Sprintf("_Runs parsed: %d (requested: %d)_\n", actualRuns, td.Count))
	}

	for _, run := range td.Runs {
		if actualRuns > 1 {
			b.WriteString(fmt.Sprintf("\n### Run %d\n\n", run.Index+1))
		}
		b.WriteString("| Benchmark | Count | ns/op | % of Full | Client Share % | Latency Share % |\n")
		b.WriteString("|-----------|-------|-------|-----------|----------------|-----------------|\n")
		for _, l := range run.Lines {
			b.WriteString(fmt.Sprintf("| %s | %d | %.0f | %.2f | %.2f | %.2f |\n", l.BaseName, l.Iterations, l.NsPerOp, l.PercentOfFull, l.ClientShare, l.LatencyShare))
		}
		if len(run.Notes) > 0 {
			b.WriteString("\n**Notes**:\n")
			for _, n := range run.Notes {
				b.WriteString("- " + n + "\n")
			}
		}
	}
	return b.String()
}

func outputTable(td tableData) string {
	var b strings.Builder
	actualRuns := len(td.Runs)
	if actualRuns == 1 {
		b.WriteString(fmt.Sprintf("Run count: %d (requested: %d)\n", actualRuns, td.Count))
	} else {
		b.WriteString(fmt.Sprintf("Runs parsed: %d (requested: %d)\n", actualRuns, td.Count))
	}

	for _, run := range td.Runs {
		if actualRuns > 1 {
			b.WriteString(fmt.Sprintf("Run %d\n", run.Index+1))
		}
		w := tabwriter.NewWriter(&b, 0, 2, 2, ' ', 0)
		fmt.Fprintf(w, "Benchmark\tCount\tNs/op\t%%Full\tClientShare%%\tLatencyShare%%\n")
		for _, l := range run.Lines {
			fmt.Fprintf(w, "%s\t%d\t%.0f\t%.2f\t%.2f\t%.2f\n", l.BaseName, l.Iterations, l.NsPerOp, l.PercentOfFull, l.ClientShare, l.LatencyShare)
		}
		w.Flush()
		if len(run.Notes) > 0 {
			b.WriteString("Notes:\n")
			for _, n := range run.Notes {
				b.WriteString("  - " + n + "\n")
			}
		}
		if actualRuns > 1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}
