# CloudScale

## Fault-Tolerant Distributed Object Storage System

CloudScale is a distributed object-storage system implemented in **Go** to study the practical trade-offs between **replication, consensus, fault tolerance, consistency, and performance**.

The project evolves from a single-node storage engine into a multi-node system using **Raft consensus, replicated state, leader election, failure detection, automatic recovery, authenticated API access, Docker, Kubernetes, observability, chaos experiments, cloud deployment, and performance evaluation**.

---

## Research Question

> **How does Raft-based replication and consistency enforcement affect the performance of a distributed object storage system as workload size increases?**

CloudScale combines systems engineering with experimental evaluation. The goal is not only to build a distributed storage service, but also to measure the performance and recovery costs introduced by distributed coordination.

---

## Key Results

Repeated 1,000-request experiments comparing a single-node configuration with a three-node Raft cluster produced:

| Metric | Single Node | 3-Node Raft | Change |
|---|---:|---:|---:|
| PUT throughput | 36.93 req/s | 16.00 req/s | -56.68% |
| PUT latency | 27.35 ms | 62.55 ms | +128.68% |
| GET throughput | 48.88 req/s | 46.47 req/s | -4.92% |
| GET latency | 20.53 ms | 21.84 ms | +6.36% |

Additional fault-tolerance experiments measured:

- **Leader recovery time:** 13.432 seconds
- Successful leader election after leader failure
- Successful writes after leader failure
- Successful reads after recovery
- Replica synchronization after a failed node rejoined
- Replica convergence to the same committed log state

### Main finding

**Distributed writes incur significant coordination overhead, while read performance remains comparatively close to the single-node baseline under the tested workload.**

This demonstrates a fundamental distributed-systems trade-off:

```text
More coordination
       |
       v
Higher write overhead
       |
       +
       |
       v
Replicated state + fault tolerance
Architecture
                              Client
                                |
                                v
                       +----------------+
                       |  API Gateway   |
                       +----------------+
                                |
                     Authentication / Routing
                                |
                    +-----------+-----------+
                    |           |           |
                    v           v           v
               +---------+ +---------+ +---------+
               | Node 1  | | Node 2  | | Node 3  |
               | Leader  | |Follower | |Follower |
               +---------+ +---------+ +---------+
                    |           |           |
                    +-----------+-----------+
                                |
                         Raft Replicated Log
                                |
                         Commit / Apply
                                |
                    +-----------+-----------+
                    |                       |
                    v                       v
              Persistent State       Object Storage
Write Path
Client
  |
  v
API Gateway
  |
  v
Raft Leader
  |
  +----> Replicate log entry
  |              |
  |              +----> Follower
  |              |
  |              +----> Follower
  |
  v
Commit
  |
  v
Apply to storage
  |
  v
Response
Failure Path
Leader Failure
      |
      v
Heartbeat timeout
      |
      v
Leader election
      |
      v
New leader
      |
      v
Writes continue
      |
      v
Failed node rejoins
      |
      v
Replica synchronization
      |
      v
Cluster convergence
Core Components
ComponentResponsibility
Storage EngineObject persistence and retrieval
Distributed NodesIndependent storage processes
ReplicationReplicated state across nodes
RaftConsensus and replicated log
Failure DetectionHeartbeats and failure detection
Fault ToleranceLeader failover and recovery
API GatewayExternal request routing
AuthenticationProtected API access
SchedulerAI-oriented workload scheduling
Dockerfile Local containerized deployment
KubernetesCluster orchestration
PrometheusMetrics collection
GrafanaMonitoring dashboards
AWS / EKSCloud deployment experiments
BenchmarkingPerformance measurement
Research AnalysisExperimental interpretation
## Raft Consensus

CloudScale uses Raft to coordinate replicated cluster state.

The implementation tracks core consensus state including:

Current term
Voted-for node
Raft log
Commit index
Last applied index
Node state
Leader identity
Replication progress

The cluster size is derived from the configured peer set, allowing the same implementation to operate in both single-node and multi-node configurations.

A replicated write follows this general process:

Client sends a request to the gateway.
Gateway identifies the active leader.
Leader appends the operation to its Raft log.
The log entry is replicated to followers.
The entry reaches the required commit condition.
The committed operation is applied to storage.
The result is returned to the client.
## Replication

CloudScale maintains replicated state across multiple nodes.

For a three-node cluster:

                   +---------+
                   | Leader  |
                   +---------+
                    /       \
                   /         \
                  v           v
           +---------+   +---------+
           | Node 2  |   | Node 3  |
           |Follower |   |Follower |
           +---------+   +---------+

Replication provides redundancy and allows the cluster to continue operating when a leader fails, provided enough nodes remain available for the implemented recovery path.

## Fault Tolerance

CloudScale was explicitly tested under leader-failure conditions.

The experiment:

Started a three-node cluster.
Established a leader.
Stopped the active leader.
Allowed the remaining nodes to detect the failure.
Observed leader election.
Sent a new write through the gateway.
Verified data availability.
Restarted the failed node.
Allowed replica synchronization.
Verified convergence.
Result

Leader recovery time: 13.432 seconds

The test successfully demonstrated:

Failure detection
Leader election
Continued write availability
Data availability
Replica recovery
Cluster convergence
## Consistency

CloudScale uses a leader-oriented consistency model for API operations.

Gateway reads are routed to the current Raft leader.

Direct follower reads can return a non-leader response rather than serving independently from a follower.

This design simplifies the consistency model while keeping state-changing operations coordinated through the Raft replicated log.

The implementation should therefore be understood as a research prototype with leader-oriented reads and Raft-coordinated replicated writes, rather than as a claim of full production-grade consistency semantics.

## API Gateway

The API gateway provides the external entry point to the distributed system.

Default ports
Gateway     : 9000
Node 1      : 8001
Node 2      : 8002
Node 3      : 8003
Health
GET /health

Provides gateway/cluster health information including leader information.

Raft status
GET /raft/status

Exposes information such as:

Current term
Raft state
Voted-for node
Commit index
Last applied index
Last log index
Leader information
Replication progress

Protected operations use authentication and authorization mechanisms implemented by the gateway.

Security

CloudScale includes security-related infrastructure including:

Authentication
Authorization
Protected API operations
Audit logging
Encryption-related functionality

Secrets and runtime credentials are intentionally excluded from the repository through .gitignore.

For production deployment, additional security hardening would still be required.

## Performance Evaluation

CloudScale includes a reproducible benchmark workflow comparing:

Single-node storage
        vs
Three-node Raft cluster

The benchmark evaluates:

PUT latency
PUT throughput
GET latency
GET throughput
Distributed coordination overhead
Resource utilization
Failure recovery

The repeated benchmark dataset uses three runs of a 1,000-request workload for each configuration and operation.

## Benchmark Results
PUT Latency

PUT Throughput

GET Latency

GET Throughput

Detailed experimental results are available in:

benchmarks/results/phase16_results.md
benchmarks/results/phase17_analysis.md
benchmarks/results/phase17_repeated_results.md

Raw benchmark data is stored under:

benchmarks/data/
## Experimental Findings
PUT operations

Mean single-node PUT latency:

27.353 ms

Mean distributed PUT latency:

62.549 ms

The distributed configuration therefore experienced approximately:

+128.68% latency

Mean PUT throughput changed from:

36.933 req/s

to:

16.000 req/s

or approximately:

-56.68%

This reflects the cost of replicated coordination.

GET operations

Mean single-node GET latency:

20.533 ms

Mean distributed GET latency:

21.838 ms

The difference was approximately:

+6.36%

GET throughput changed from:

48.877 req/s

to:

46.470 req/s

or approximately:

-4.92%

Under the tested workload, reads therefore experienced substantially less overhead than writes.

## Replica Convergence

During recovery testing, all three replicas converged to the same committed state.

Observed final indices:

commit_index   = 1603
last_log_index = 1603
last_applied   = 1603

Matching indices across replicas provide evidence that the cluster successfully synchronized its committed state after recovery.

## Observability

CloudScale integrates with:

Prometheus for metrics collection
Grafana for visualization
Health endpoints
Raft status endpoints

Monitoring manifests are included in:

deployments/

The observability phase is intended to make cluster behavior and runtime performance easier to inspect during experiments.

## Docker

CloudScale supports local containerized execution using Docker Compose.

Distributed cluster
docker compose up --build
Single-node baseline
docker compose -f docker-compose.single.yml up --build

The two configurations make it possible to compare local storage behavior against replicated distributed operation.

## Kubernetes

Kubernetes manifests are provided under:

deployments/

The deployment includes configuration for:

Gateway
Node 1
Node 2
Node 3
Prometheus
Grafana
Namespace

The project also contains AWS-specific deployment configuration under:

deployments/aws/
## AWS / EKS

CloudScale includes configuration for deploying the system to Amazon EKS.

The cloud deployment phase extends the project from local Docker/Kubernetes experimentation toward managed cloud infrastructure.

The repository provides infrastructure configuration rather than claiming production-scale cloud deployment.

Chaos and Failure Experiments

Failure-oriented testing was used to validate distributed-system behavior under node failure.

The experiments include:

Leader failure
      |
      v
Failure detection
      |
      v
Leader election
      |
      v
New leader
      |
      v
Write availability
      |
      v
Node recovery
      |
      v
Replica synchronization

These experiments complement the performance benchmarks by evaluating not only how fast the system operates, but also how it behaves when components fail.

## Project Structure
Cloudscale/
│
├── cmd/
│   ├── gateway/
│   │   └── main.go
│   │
│   └── node/
│       ├── main.go
│       ├── storage_apply.go
│       └── storage_apply_test.go
│
├── internal/
│   ├── auth/
│   ├── raft/
│   ├── replication/
│   ├── scheduler/
│   └── storage/
│
├── deployments/
│   ├── aws/
│   │   └── cluster.yaml
│   ├── gateway.yaml
│   ├── grafana.yaml
│   ├── namespace.yaml
│   ├── node1.yaml
│   ├── node2.yaml
│   ├── node3.yaml
│   ├── prometheus-config.yaml
│   └── prometheus.yaml
│
├── benchmarks/
│   ├── data/
│   ├── results/
│   └── scripts/
│
├── scripts/
├── tests/
├── data/
│
├── Dockerfile
├── docker-compose.yml
├── docker-compose.single.yml
├── go.mod
├── .gitignore
└── README.md
## Reproducibility

The benchmark workflow separates experiment execution from analysis:

Benchmark scripts
       |
       v
Raw measurements
       |
       v
CSV / text datasets
       |
       v
Python analysis
       |
       v
Statistical summaries
       |
       v
Graphs
       |
       v
Research findings

Benchmark scripts and analysis programs are available under:

benchmarks/scripts/

This structure allows future experiments to reuse the same measurement and analysis pipeline.

## Research Contribution

CloudScale is primarily an engineering and experimental systems project.

Its contribution is the integration of several distributed-systems mechanisms into a working prototype and the empirical evaluation of their trade-offs.

The project connects:

Storage Systems
      +
Distributed Systems
      +
Consensus
      +
Fault Tolerance
      +
Cloud Infrastructure
      +
Performance Engineering
      +
Experimental Research

The benchmark results provide a concrete example of the cost of distributed coordination:

Fault tolerance and replicated state are not free; they introduce measurable latency and throughput overhead, particularly for write-heavy workloads.

## Limitations

CloudScale is a research and engineering prototype rather than a production object-storage platform.

Current limitations include:

Relatively small benchmark workloads
Limited cluster size
Environment-dependent performance measurements
No dedicated distributed benchmark hardware
Limited large-object evaluation
Simplified consistency semantics
Recovery time dependent on runtime and environment
Further security hardening required for production use
Cloud deployment configuration intended for experimentation
Additional scalability testing required

These limitations are explicitly documented so that experimental results are interpreted within the scope of the current implementation.

## Future Work

Potential future extensions include:

Larger cluster experiments
Concurrent client workloads
Larger object sizes
Network latency injection
Packet-loss experiments
Disk-failure simulation
Automated cluster membership changes
Improved replica placement
Storage compaction
Snapshotting
Advanced Raft log management
Distributed tracing
Multi-region experiments
Cloud cost analysis
Additional consistency models
More comprehensive security hardening
## Technology Stack
CategoryTechnology
LanguageGo
ConsensusRaft
StoragePersistent local object storage
ContainersDocker
OrchestrationKubernetes
CloudAWS / Amazon EKS
MonitoringPrometheus
DashboardsGrafana
AnalysisPython
BenchmarkingShell + Python
Version ControlGit / GitHub
## Development Roadmap

CloudScale follows an 18-phase development roadmap:

Phase 0  — Environment
Phase 1  — Storage Engine
Phase 2  — Distributed Nodes
Phase 3  — Replication
Phase 4  — Raft
Phase 5  — Failure Detection
Phase 6  — Fault Tolerance
Phase 7  — Consistency
Phase 8  — API Gateway
Phase 9  — Security
Phase 10 — Docker
Phase 11 — Kubernetes
Phase 12 — Observability
Phase 13 — Chaos Engineering
Phase 14 — AI Infrastructure
Phase 15 — Cloud
Phase 16 — Benchmarking
Phase 17 — Research
Phase 18 — Admissions Presentation
### Current status

**Phase 17 — Research: In Progress**

Completed engineering and evaluation phases include:

- Storage engine
- Distributed nodes
- Replication
- Raft consensus
- Leader election
- Failure detection
- Fault tolerance
- Replica recovery
- Consistency behavior
- API gateway
- Security mechanisms
- Docker
- Kubernetes
- Observability
- Chaos/failure testing
- AI infrastructure
- AWS/EKS configuration
- Benchmarking
- Research analysis

---

## Why CloudScale?

CloudScale was intentionally designed to go beyond a conventional CRUD application.

The project explores the complete path from local storage to distributed infrastructure:

```text
Single Node
    |
    v
Replication
    |
    v
Consensus
    |
    v
Fault Tolerance
    |
    v
Cloud Infrastructure
    |
    v
Performance Evaluation
    |
    v
Research
```

The central objective is to understand not only **how** distributed storage works, but also **what it costs** in latency, throughput, coordination, and recovery.

---

## Author

**Anju Ajayakumar**

Independent Systems & Distributed Computing Project  
2026

---

## License

This project is currently intended as an academic and research portfolio project.
