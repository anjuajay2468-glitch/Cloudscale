#!/bin/bash

TOTAL=100
URL="http://localhost:9000/objects"
KEY="cloudscale-reader-123"

echo "Running $TOTAL GET requests..."
echo

total_time=0

for i in $(seq 1 $TOTAL); do
    start=$(date +%s%N)

    curl -s -o /dev/null \
      -H "X-API-Key: $KEY" \
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
