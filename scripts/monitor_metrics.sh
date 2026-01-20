#!/bin/bash
# Prometheus Metrics Monitor
# Shows metrics updating in real-time

ENDPOINT="${1:-http://localhost:9090/metrics}"
INTERVAL="${2:-3}"

echo "Monitoring Prometheus metrics from: $ENDPOINT"
echo "Sampling every ${INTERVAL} seconds"
echo "Press Ctrl+C to stop"
echo ""

while true; do
    TIMESTAMP=$(date '+%H:%M:%S')

    echo "╔═══════════════════════════════════════════════════════════════"
    echo "║ Time: $TIMESTAMP"
    echo "╠═══════════════════════════════════════════════════════════════"

    # Counters
    echo "║ COUNTERS:"
    curl -s $ENDPOINT | grep -E "pingheat_ping_(sent|success|timeout)_total" | grep -v "^#" | sed 's/^/║   /'

    # Status
    echo "║"
    echo "║ STATUS:"
    curl -s $ENDPOINT | grep -E "pingheat_ping_up|pingheat_ping_current_streak|pingheat_uptime_seconds" | grep -v "^#" | sed 's/^/║   /'

    # Latency
    echo "║"
    echo "║ LATENCY:"
    curl -s $ENDPOINT | grep -E "pingheat_ping_latency_ms|pingheat_ping_last_rtt" | grep -v "^#" | sed 's/^/║   /'

    # Statistics
    echo "║"
    echo "║ STATISTICS:"
    curl -s $ENDPOINT | grep -E "pingheat_ping_(stddev|jitter|variance)_ms" | grep -v "^#" | sed 's/^/║   /'

    # Percentiles
    echo "║"
    echo "║ PERCENTILES:"
    curl -s $ENDPOINT | grep -E "pingheat_ping_latency_p[0-9]" | grep -v "^#" | sed 's/^/║   /'

    # Availability
    echo "║"
    echo "║ AVAILABILITY:"
    curl -s $ENDPOINT | grep -E "pingheat_ping_(loss|availability)_percent" | grep -v "^#" | sed 's/^/║   /'

    # Instability
    echo "║"
    echo "║ INSTABILITY:"
    curl -s $ENDPOINT | grep -E "pingheat_ping_(loss_bursts|brownout|longest)" | grep -v "^#" | sed 's/^/║   /'

    echo "╚═══════════════════════════════════════════════════════════════"
    echo ""

    sleep $INTERVAL
done
