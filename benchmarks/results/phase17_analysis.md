# CloudScale — Phase 17 Research Analysis

## Research Questions

### RQ1 — Replication overhead

How does Raft-based replication affect latency and throughput compared
with a single-node configuration?

### RQ2 — Workload scaling

How does increasing the number of sequential requests affect CloudScale
performance?

### RQ3 — Fault tolerance

How quickly can CloudScale recover from leader failure while maintaining
data availability?

### RQ4 — Resource overhead

What additional CPU and memory resources are required by the distributed
configuration?

---

# Hypotheses

## H1

Raft replication will increase write latency compared with a single-node
configuration.

## H2

Increasing workload size will reduce throughput and increase latency.

## H3

A three-node Raft cluster will remain available after failure of one node
because two nodes retain majority quorum.

## H4

The distributed configuration will consume more system resources than the
single-node configuration.

---

# Quantitative Findings

## PUT

### 100 requests

Distributed latency: 22.411 ms

Single-node latency: 22.969 ms

Difference: -2.43%

Distributed throughput: 44.62 req/s

Single-node throughput: 43.54 req/s

Difference: +2.48%

At this workload, the distributed configuration showed slightly better
measured performance.

### 500 requests

Distributed latency: 30.402 ms

Single-node latency: 20.709 ms

Latency increase: 46.81%

Distributed throughput: 32.89 req/s

Single-node throughput: 48.29 req/s

Throughput reduction: 31.89%

### 1000 requests

Distributed latency: 32.405 ms

Single-node latency: 26.068 ms

Latency increase: 24.31%

Distributed throughput: 30.86 req/s

Single-node throughput: 38.36 req/s

Throughput reduction: 19.55%

---

# GET

## 100 requests

Distributed latency: 16.793 ms

Single-node latency: 14.412 ms

Latency increase: 16.52%

Distributed throughput: 59.55 req/s

Single-node throughput: 69.38 req/s

Throughput reduction: 14.17%

## 500 requests

Distributed latency: 17.884 ms

Single-node latency: 19.744 ms

Latency difference: -9.42%

Distributed throughput: 55.92 req/s

Single-node throughput: 50.65 req/s

Throughput difference: +10.40%

At this workload, the distributed configuration slightly outperformed
the single-node configuration.

## 1000 requests

Distributed latency: 27.006 ms

Single-node latency: 16.026 ms

Latency increase: 68.51%

Distributed throughput: 37.03 req/s

Single-node throughput: 62.40 req/s

Throughput reduction: 40.66%

---

# Fault-Tolerance Findings

The original leader node3 was deliberately stopped.

A replacement leader was elected from the remaining two nodes.

Measured recovery time:

13.432 seconds

The remaining two-node majority continued to accept writes.

Existing data remained readable.

The failed node was subsequently restarted and synchronized with the
cluster.

Final commit index:

1603

All three nodes reached:

commit_index = 1603

last_log_index = 1603

---

# Resource Findings

Three-node distributed configuration:

Node 1:
CPU = 11.88%
Memory = 7.45 MiB

Node 2:
CPU = 20.66%
Memory = 7.29 MiB

Node 3:
CPU = 8.90%
Memory = 7.08 MiB

Approximate total node memory:

21.82 MiB

Single-node configuration:

CPU = 1.68%

Memory = 9.34 MiB

The distributed configuration requires additional resources because
multiple Raft nodes maintain replicated state and communicate with
each other.

---

# Interpretation

The measurements provide partial support for H1.

Distributed PUT latency was higher at 500 and 1000 requests, although
the 100-request measurement showed slightly lower latency than the
single-node configuration.

H2 is supported by the observed reduction in throughput at larger
workloads, particularly in the distributed configuration. However,
the measurements are not perfectly monotonic and should not be
interpreted as a linear scaling relationship.

H3 is supported by the fault-tolerance experiment. The three-node
cluster continued operating after loss of one node because the
remaining two nodes maintained majority quorum.

H4 is supported by the resource measurements. Running three Raft
nodes requires more aggregate CPU and memory than running a single
node.

---

# Research Limitations

The experiments were conducted using:

- Local Docker containers
- Sequential requests
- A three-node cluster
- Relatively small objects
- A local network
- Limited workload sizes

Therefore, the results should not be interpreted as production-scale
performance measurements.

Future experiments should evaluate:

- Concurrent clients
- Larger objects
- Sustained workloads
- Network latency
- Network partitions
- Node crashes during writes
- Larger cluster sizes
- Higher request rates
- Multiple benchmark repetitions
