# CloudScale — Phase 17 Repeated-Run Research Analysis

## 1. Objective

Phase 17 analyzes the experimental results collected during Phase 16 and
extends the evaluation using repeated 1000-request workloads.

Three runs were performed for each of:

- Distributed PUT
- Distributed GET
- Single-node PUT
- Single-node GET

The purpose was to measure average behavior and run-to-run variability.

---

## 2. Repeated-Run Dataset

| Configuration | Operation | Run | Latency (ms) | Throughput (req/s) |
|---|---|---:|---:|---:|
| Distributed | PUT | 1 | 63.919 | 15.64 |
| Distributed | PUT | 2 | 59.891 | 16.70 |
| Distributed | PUT | 3 | 63.837 | 15.66 |
| Distributed | GET | 1 | 20.295 | 49.27 |
| Distributed | GET | 2 | 19.522 | 51.22 |
| Distributed | GET | 3 | 25.697 | 38.92 |
| Single-node | PUT | 1 | 24.273 | 41.20 |
| Single-node | PUT | 2 | 26.756 | 37.37 |
| Single-node | PUT | 3 | 31.029 | 32.23 |
| Single-node | GET | 1 | 22.297 | 44.85 |
| Single-node | GET | 2 | 19.570 | 51.10 |
| Single-node | GET | 3 | 19.731 | 50.68 |

---

## 3. Statistical Summary

### Distributed PUT

- Mean latency: 62.549 ms
- Standard deviation: 2.302 ms
- Minimum latency: 59.891 ms
- Maximum latency: 63.919 ms
- Mean throughput: 16.00 req/s
- Throughput standard deviation: 0.606 req/s

### Distributed GET

- Mean latency: 21.838 ms
- Standard deviation: 3.364 ms
- Minimum latency: 19.522 ms
- Maximum latency: 25.697 ms
- Mean throughput: 46.47 req/s
- Throughput standard deviation: 6.611 req/s

### Single-node PUT

- Mean latency: 27.353 ms
- Standard deviation: 3.417 ms
- Minimum latency: 24.273 ms
- Maximum latency: 31.029 ms
- Mean throughput: 36.93 req/s
- Throughput standard deviation: 4.501 req/s

### Single-node GET

- Mean latency: 20.533 ms
- Standard deviation: 1.530 ms
- Minimum latency: 19.570 ms
- Maximum latency: 22.297 ms
- Mean throughput: 48.88 req/s
- Throughput standard deviation: 3.494 req/s

---

## 4. Distributed vs Single-Node Comparison

### PUT

Distributed mean latency was 62.549 ms compared with 27.353 ms for the
single-node configuration.

This represents approximately 128.68% higher latency.

Distributed mean throughput was 16.00 req/s compared with 36.93 req/s for
the single-node configuration, representing approximately 56.68% lower
throughput.

The increased write cost is consistent with the distributed architecture:
writes require Raft replication and majority commitment before successful
completion.

### GET

Distributed mean latency was 21.838 ms compared with 20.533 ms for the
single-node configuration.

This represents approximately 6.36% higher latency.

Distributed mean throughput was 46.47 req/s compared with 48.88 req/s,
representing approximately 4.92% lower throughput.

The relatively small difference suggests that read operations experience
much less distributed overhead than writes.

---

## 5. Interpretation

The experiments demonstrate an asymmetric performance cost.

Writes are substantially more expensive in the distributed configuration
because successful writes require consensus and replication across the Raft
cluster.

Reads show considerably smaller performance differences because they do not
require the same majority-commit operation.

This behavior is consistent with the intended design trade-off between
fault tolerance and performance.

---

## 6. Variability

Distributed PUT showed relatively low run-to-run variability, with latency
between 59.891 ms and 63.919 ms.

Distributed GET showed greater variability, primarily due to the third run,
which reached 25.697 ms.

Single-node GET had the lowest latency variability among the four workloads.

These results demonstrate why repeated measurements are preferable to relying
on a single benchmark run.

---

## 7. Comparison With Earlier Phase 16 Baseline

The original Phase 16 distributed PUT benchmark at 1000 requests measured:

- 32.405 ms average latency
- 30.86 req/s throughput

The repeated Phase 17 distributed PUT experiment measured:

- 62.549 ms mean latency
- 16.00 req/s mean throughput

The repeated experiment therefore produced substantially slower results.

The difference should not be interpreted as a regression without controlled
investigation. The experiments were conducted at different points during
the development session and under a local Docker/WSL environment where
filesystem state, cached data, background workloads, and accumulated Raft
logs may influence measurements.

For this reason, the repeated-run experiment is treated as a separate
measurement set rather than being combined numerically with the earlier
baseline.

---

## 8. Key Findings

### Finding 1 — Distributed writes have significant overhead

The repeated measurements show approximately 129% higher write latency and
57% lower write throughput compared with the single-node configuration.

### Finding 2 — Distributed reads remain relatively efficient

GET latency increased by only approximately 6%, while throughput decreased
by approximately 5%.

### Finding 3 — Replication creates a measurable performance trade-off

The system sacrifices write performance in exchange for replicated state,
majority commitment, and fault tolerance.

### Finding 4 — Repeated experiments reveal variability

Single measurements can hide environmental variation. Repeated runs provide
a more reliable estimate of system behavior.

### Finding 5 — Fault tolerance has a measurable cost

The benchmark results complement the earlier Phase 16 fault-tolerance
experiments, which demonstrated leader recovery and replica synchronization.

---

## 9. Limitations

The experiments were conducted:

- sequentially
- on a local development machine
- inside Docker
- using WSL
- with 1000-request workloads
- using relatively small objects

The measurements should therefore not be interpreted as production-scale
performance benchmarks.

Future work should evaluate concurrent clients, larger objects, sustained
load, network latency, network partitions, node failures during writes,
and larger cluster sizes.

---

## 10. Conclusion

The repeated experiments provide quantitative evidence that CloudScale's
distributed architecture introduces a substantial write-performance cost
while maintaining comparatively similar read performance.

The results demonstrate a clear trade-off between distributed consistency,
replication, fault tolerance, and raw write performance.

This establishes an empirical foundation for the technical conclusions and
research discussion surrounding CloudScale.
