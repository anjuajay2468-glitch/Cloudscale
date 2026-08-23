#!/bin/bash

TOTAL=1000
URL="http://localhost:9000/objects"
KEY="cloudscale-admin-123"

echo "Running $TOTAL PUT requests..."
echo

total_time=0

for i in $(seq 1 $TOTAL); do
    start=$(date +%s%N)

    curl -s -o /dev/null \
      -X PUT \
      -H "X-API-Key: $KEY" \
      -H "Content-Type: application/octet-stream" \
      --data-binary "benchmark-payload-$i" \
      "$URL/benchmark-$i"

    end=$(date +%s%N)

    elapsed=$((end-start))
    total_time=$((total_time+elapsed))

    echo "$i $elapsed"
done

echo
echo "===== SUMMARY ====="

avg=$((total_time / TOTAL))

echo "Requests: $TOTAL"
echo "Total time (ns): $total_time"
echo "Average latency (ns): $avg"
echo "Average latency (ms): $(awk "BEGIN {printf \"%.3f\", $avg/1000000}")"
echo "Throughput (requests/sec): $(awk "BEGIN {printf \"%.2f\", $TOTAL/(($total_time)/1000000000)}")"
