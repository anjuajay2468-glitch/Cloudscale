import csv
import matplotlib.pyplot as plt

data = []

with open("phase16_benchmarks.csv") as f:
    for row in csv.DictReader(f):
        data.append({
            "configuration": row["configuration"],
            "operation": row["operation"],
            "requests": int(row["requests"]),
            "latency": float(row["latency_ms"]),
            "throughput": float(row["throughput_rps"])
        })


def get_rows(operation, configuration):
    return [
        r for r in data
        if r["operation"] == operation
        and r["configuration"] == configuration
    ]


# --------------------------------------------------
# 1. PUT Latency
# --------------------------------------------------

plt.figure()

for configuration in ["distributed", "single-node"]:
    rows = get_rows("PUT", configuration)

    x = [r["requests"] for r in rows]
    y = [r["latency"] for r in rows]

    plt.plot(x, y, marker="o", label=configuration)

plt.xlabel("Number of Requests")
plt.ylabel("Average Latency (ms)")
plt.title("PUT Latency: Distributed vs Single-Node")
plt.legend()
plt.grid(True)
plt.tight_layout()
plt.savefig("phase17_put_latency.png", dpi=300)
plt.close()


# --------------------------------------------------
# 2. GET Latency
# --------------------------------------------------

plt.figure()

for configuration in ["distributed", "single-node"]:
    rows = get_rows("GET", configuration)

    x = [r["requests"] for r in rows]
    y = [r["latency"] for r in rows]

    plt.plot(x, y, marker="o", label=configuration)

plt.xlabel("Number of Requests")
plt.ylabel("Average Latency (ms)")
plt.title("GET Latency: Distributed vs Single-Node")
plt.legend()
plt.grid(True)
plt.tight_layout()
plt.savefig("phase17_get_latency.png", dpi=300)
plt.close()


# --------------------------------------------------
# 3. PUT Throughput
# --------------------------------------------------

plt.figure()

for configuration in ["distributed", "single-node"]:
    rows = get_rows("PUT", configuration)

    x = [r["requests"] for r in rows]
    y = [r["throughput"] for r in rows]

    plt.plot(x, y, marker="o", label=configuration)

plt.xlabel("Number of Requests")
plt.ylabel("Throughput (requests/sec)")
plt.title("PUT Throughput: Distributed vs Single-Node")
plt.legend()
plt.grid(True)
plt.tight_layout()
plt.savefig("phase17_put_throughput.png", dpi=300)
plt.close()


# --------------------------------------------------
# 4. GET Throughput
# --------------------------------------------------

plt.figure()

for configuration in ["distributed", "single-node"]:
    rows = get_rows("GET", configuration)

    x = [r["requests"] for r in rows]
    y = [r["throughput"] for r in rows]

    plt.plot(x, y, marker="o", label=configuration)

plt.xlabel("Number of Requests")
plt.ylabel("Throughput (requests/sec)")
plt.title("GET Throughput: Distributed vs Single-Node")
plt.legend()
plt.grid(True)
plt.tight_layout()
plt.savefig("phase17_get_throughput.png", dpi=300)
plt.close()


print("Generated:")
print("  phase17_put_latency.png")
print("  phase17_get_latency.png")
print("  phase17_put_throughput.png")
print("  phase17_get_throughput.png")
