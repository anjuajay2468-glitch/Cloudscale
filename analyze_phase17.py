import csv

rows = []

with open("phase16_benchmarks.csv") as f:
    rows = list(csv.DictReader(f))


def find(configuration, operation, requests):
    for r in rows:
        if (
            r["configuration"] == configuration
            and r["operation"] == operation
            and int(r["requests"]) == requests
        ):
            return r


def pct_change(distributed, single):
    return ((distributed - single) / single) * 100


print("CloudScale Phase 17 - Quantitative Analysis")
print("=" * 55)

for operation in ["PUT", "GET"]:
    print("\n" + operation)
    print("-" * 55)

    for requests in [100, 500, 1000]:
        d = find("distributed", operation, requests)
        s = find("single-node", operation, requests)

        d_latency = float(d["latency_ms"])
        s_latency = float(s["latency_ms"])

        d_throughput = float(d["throughput_rps"])
        s_throughput = float(s["throughput_rps"])

        latency_change = pct_change(d_latency, s_latency)
        throughput_change = pct_change(d_throughput, s_throughput)

        print("\nRequests:", requests)

        print(
            "Latency:",
            "distributed =", d_latency,
            "ms | single-node =", s_latency,
            "ms"
        )

        print(
            "Latency difference:",
            "{:.2f}%".format(latency_change)
        )

        print(
            "Throughput:",
            "distributed =", d_throughput,
            "req/s | single-node =", s_throughput,
            "req/s"
        )

        print(
            "Throughput difference:",
            "{:.2f}%".format(throughput_change)
        )


print("\n" + "=" * 55)
print("Analysis complete.")
