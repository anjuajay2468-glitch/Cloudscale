import statistics

data = {
    "distributed_put": {
        "latency": [63.919, 59.891, 63.837],
        "throughput": [15.64, 16.70, 15.66],
    },
    "distributed_get": {
        "latency": [20.295, 19.522, 25.697],
        "throughput": [49.27, 51.22, 38.92],
    },
    "single_put": {
        "latency": [24.273, 26.756, 31.029],
        "throughput": [41.20, 37.37, 32.23],
    },
    "single_get": {
        "latency": [22.297, 19.570, 19.731],
        "throughput": [44.85, 51.10, 50.68],
    },
}

print("=" * 65)
print("CloudScale Phase 17 - Repeated Run Statistical Analysis")
print("=" * 65)

results = {}

for name, metrics in data.items():
    print(f"\n{name.upper().replace('_', ' ')}")
    print("-" * 65)

    results[name] = {}

    for metric, values in metrics.items():
        mean = statistics.mean(values)
        stdev = statistics.stdev(values)
        minimum = min(values)
        maximum = max(values)

        results[name][metric] = {
            "mean": mean,
            "stdev": stdev,
            "min": minimum,
            "max": maximum,
        }

        unit = "ms" if metric == "latency" else "req/s"

        print(
            f"{metric.capitalize():12}: "
            f"mean={mean:.3f} {unit}, "
            f"std={stdev:.3f}, "
            f"min={minimum:.3f}, "
            f"max={maximum:.3f}"
        )

print("\n" + "=" * 65)
print("DISTRIBUTED VS SINGLE-NODE")
print("=" * 65)

comparisons = [
    ("PUT", "distributed_put", "single_put"),
    ("GET", "distributed_get", "single_get"),
]

for operation, dist, single in comparisons:
    print(f"\n{operation}")

    dist_lat = results[dist]["latency"]["mean"]
    single_lat = results[single]["latency"]["mean"]

    dist_tp = results[dist]["throughput"]["mean"]
    single_tp = results[single]["throughput"]["mean"]

    latency_overhead = ((dist_lat - single_lat) / single_lat) * 100
    throughput_change = ((dist_tp - single_tp) / single_tp) * 100

    print(f"Distributed mean latency : {dist_lat:.3f} ms")
    print(f"Single-node mean latency : {single_lat:.3f} ms")
    print(f"Latency difference       : {latency_overhead:.2f}%")

    print(f"Distributed mean throughput : {dist_tp:.2f} req/s")
    print(f"Single-node mean throughput : {single_tp:.2f} req/s")
    print(f"Throughput difference       : {throughput_change:.2f}%")

print("\n" + "=" * 65)
print("Analysis complete.")
print("=" * 65)
