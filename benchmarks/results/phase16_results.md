# CloudScale — Phase 16 Benchmarking Results

## 1. Objective

Measure the performance and fault-tolerance characteristics of the CloudScale
distributed storage system.

Metrics evaluated:

- Throughput
- Latency
- Replication overhead
- CPU usage
- Memory usage
- Leader recovery time
- Replica synchronization
- Availability during leader failure

---

## 2. Experimental Environment

Architecture:

- 3-node Raft cluster
- 1 API Gateway
- Docker Compose
- Sequential HTTP benchmarks
- Local development environment

Distributed configuration:

Gateway
  |
  +-- Node 1
  +-- Node 2
  +-- Node 3

Raft configuration:

- 1 leader
- 2 followers
- Majority quorum = 2/3

---

# 3. Distributed Cluster Baseline

## PUT Benchmark

| Requests | Average Latency | Throughput |
|----------|-----------------|------------|
| 100 | 22.411 ms | 44.62 req/s |
| 500 | 30.402 ms | 32.89 req/s |
| 1000 | 32.405 ms | 30.86 req/s |

## GET Benchmark

| Requests | Average Latency | Throughput |
|----------|-----------------|------------|
| 100 | 16.793 ms | 59.55 req/s |
| 500 | 17.884 ms | 55.92 req/s |
| 1000 | 27.006 ms | 37.03 req/s |

---

# 4. Single-Node Baseline

## PUT Benchmark

| Requests | Average Latency | Throughput |
|----------|-----------------|------------|
| 100 | 22.969 ms | 43.54 req/s |
| 500 | 20.709 ms | 48.29 req/s |
| 1000 | 26.068 ms | 38.36 req/s |

## GET Benchmark

| Requests | Average Latency | Throughput |
|----------|-----------------|------------|
| 100 | 14.412 ms | 69.38 req/s |
| 500 | 19.744 ms | 50.65 req/s |
| 1000 | 16.026 ms | 62.40 req/s |

---

# 5. Replication Overhead

The distributed system introduces additional latency for writes because
a successful Raft write requires replication to a majority of nodes.

At 1000 requests:

Distributed PUT:

- Latency: 32.405 ms
- Throughput: 30.86 req/s

Single-node PUT:

- Latency: 26.068 ms
- Throughput: 38.36 req/s

Approximate distributed PUT latency increase:

24.3%

Approximate throughput reduction:

19.6%

This overhead represents the cost of distributed consensus and replication.

---

# 6. GET Performance

At 1000 requests:

Distributed GET:

- Latency: 27.006 ms
- Throughput: 37.03 req/s

Single-node GET:

- Latency: 16.026 ms
- Throughput: 62.40 req/s

The distributed deployment shows additional overhead caused by the
gateway and distributed architecture.

---

# 7. Fault-Tolerance Experiment

The original leader (node3) was deliberately stopped.

Before failure:

- Leader: node3
- Followers: node1, node2

After failure:

- New leader: node1
- Remaining quorum: node1 + node2

Measured leader recovery time:

13,432 ms

The remaining two-node majority successfully elected a new leader.

---

# 8. Availability During Failure

After node3 was stopped:

Existing data remained readable.

A new object was successfully written:

failure-write-test

The write was committed through Raft on node1.

The object was subsequently readable through the gateway.

This demonstrates that the system remained operational after losing
one of three nodes.

---

# 9. Replica Recovery

Node3 was restarted after the failure.

Final Raft state:

Leader: node2

Followers:

- node1
- node3

Final commit index:

1603

Final last log index:

1603

All three nodes reported:

commit_index = 1603
last_log_index = 1603

Therefore:

Replica synchronization: PASS

---

# 10. Resource Usage

Three-node deployment sample:

| Component | CPU | Memory |
|-----------|-----|--------|
| Node 1 | 11.88% | 7.45 MiB |
| Node 2 | 20.66% | 7.29 MiB |
| Node 3 | 8.90% | 7.08 MiB |

Approximate total node memory:

21.82 MiB

Approximate sampled total node CPU:

41.44%

Single-node deployment sample:

| Component | CPU | Memory |
|-----------|-----|--------|
| Node | 1.68% | 9.34 MiB |

The distributed system consumes more resources because three independent
Raft nodes must maintain replicated state and communicate with one another.

---

# 11. Results Summary

| Metric | Result |
|--------|--------|
| Maximum tested workload | 1000 requests |
| Best distributed PUT throughput | 44.62 req/s |
| Best distributed GET throughput | 59.55 req/s |
| Distributed PUT latency @1000 | 32.405 ms |
| Distributed GET latency @1000 | 27.006 ms |
| Leader recovery time | 13.432 s |
| Replica synchronization | PASS |
| Write availability after leader failure | PASS |
| Data availability after failure | PASS |

---

# 12. Findings

### Finding 1 — Replication introduces measurable overhead

Distributed writes are slower than single-node writes because data must be
committed through the Raft quorum.

### Finding 2 — The system remains available after losing one node

With three nodes, two nodes are sufficient to maintain quorum.

### Finding 3 — Raft successfully recovers leadership

When the original leader failed, another node became leader and continued
serving writes.

### Finding 4 — Replicas converge

After recovery, all three nodes reached the same commit index and log index.

### Finding 5 — Distributed operation consumes additional resources

Running three Raft nodes increases CPU and memory usage compared with a
single-node deployment.

---

# 13. Limitations

These experiments were performed using sequential requests on a local Docker
environment.

The measurements therefore should not be interpreted as production-scale
performance numbers.

Future experiments should evaluate:

- Concurrent clients
- Larger objects
- Higher request rates
- Network latency
- Network partitions
- Node crashes during writes
- Larger cluster sizes
- Sustained workloads

---

# 14. Conclusion

The Phase 16 experiments demonstrate that CloudScale provides a functioning
distributed storage architecture with measurable replication overhead and
fault-tolerant behavior.

The system successfully:

- Replicates data across three nodes
- Commits operations using Raft
- Elects a replacement leader after failure
- Continues accepting writes after leader failure
- Recovers failed nodes
- Synchronizes replicas

These measurements provide the experimental foundation for Phase 17:
Research and analysis.
