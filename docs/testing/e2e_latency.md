# E2E Latency Benchmarks

The script `viiper/testing/e2e/scripts/lat_bench.go` runs (or parses) end‑to‑end input latency benchmarks and produces enriched output (table, markdown, or JSON).

It groups repeated cycles when `-count > 1` and uses the single press E2E measurement (`E2E-InputDelay`) as the 100% baseline.

## Output

| Column | Meaning |
|--------|---------|
| Benchmark | Name of the sub benchmark |
| Count | Iterations performed (from Go bench output; affected by `-benchtime`) |
| ns/op | Nanoseconds per operation (direct Go benchmark figure) |
| % of Full | Relative to `E2E-InputDelay` (single press baseline) |
| Client Share % | Portion attributed to the (go) client write phase (for E2E rows) |
| Latency Share % | Remainder attributed to transport + virtual device/host stack + tight device polling loop |

`E2E-PressAndRelease` includes both press and release cycles, so it is expected to be ~2× the single press and thus can exceed 100% in `% of Full`.

## Scope / Methodology

- All benchmarks included here are executed against a VIIPER server on the same host (localhost).  
  They therefore measure in-process client emission plus local USBIP stack + emulated device processing only.  
  Remote/network USBIP attachment will add network RTT and jitter which is intentionally excluded from these baseline figures.
- Benchmarks use a single emulated Xbox360 controller device.  
  Other devices might produce slightly different results depending on USB report size and VIIPER-InputState size.
- Benchmarks use a single button press, which is enough as clients/VIIPER always produce a full report of the devices state.  

## Benchtime Mode

Runs use a fixed-iteration benchtime (e.g. `-benchtime=1000x`, `-benchtime=10000x`) rather than time-based (e.g. `2s`).  

## Running

From repository root:

```bash
cd viiper/testing/e2e
# Single run, 1000 fixed iterations per sub benchmark
go run ./scripts/lat_bench.go -benchtime=1000x -count=1 -format markdown
```

Example output (Windows / AMD Ryzen 9 3900X / Go 1.25+, 10k iterations):

| Benchmark | Count | ns/op | % of Full | Client Share % | Latency Share % |
|-----------|-------|-------|-----------|----------------|-----------------|
| 1_Go-Client-Write | 10000 | 27933 | 16.60 | 100.00 | 0.00 |
| 2_InputDelay-Without-Client | 10000 | 133724 | 79.45 | 0.00 | 100.00 |
| 3_E2E-InputDelay | 10000 | 168307 | 100.00 | 16.60 | 83.40 |
| 4_E2E-PressAndRelease | 10000 | 331439 | 196.93 | 16.86 | 83.14 |

Variability across repeated measurement runs has been negligible.  
Use a larger `-count` if you want to increase the number of runs.

## Notes

- Memory statistics from Go benchmarks are intentionally omitted.
- `% of Full` falls back to the largest ns/op if the baseline row is missing.
- All benchmarking must run with parallelism 1 in underlying benches.
- Benchmarks use a tight polling loop using SDL3 to detect input state changes on the emulated device.
- Benchmarks must be run without any other game controllers connected and without an already running VIIPER server instance.
