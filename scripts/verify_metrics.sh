#!/bin/bash
# Quick metrics verification script
# Samples metrics twice and shows the difference

ENDPOINT="${1:-http://localhost:9090/metrics}"

echo "Verifying Prometheus metrics are updating dynamically..."
echo ""

# Sample 1
echo "📊 Sample 1 (now):"
SAMPLE1=$(curl -s $ENDPOINT | grep "pingheat_ping_sent_total" | grep -v "^#")
echo "  $SAMPLE1"

SENT1=$(echo "$SAMPLE1" | grep -oE '[0-9]+$')
UPTIME1=$(curl -s $ENDPOINT | grep "pingheat_uptime_seconds" | grep -oE '[0-9.]+$')
RTT1=$(curl -s $ENDPOINT | grep "pingheat_ping_last_rtt_ms" | grep -oE '[0-9.]+$')

echo "  Uptime: ${UPTIME1}s"
echo "  Last RTT: ${RTT1}ms"

# Wait
echo ""
echo "⏳ Waiting 5 seconds..."
sleep 5

# Sample 2
echo ""
echo "📊 Sample 2 (after 5 seconds):"
SAMPLE2=$(curl -s $ENDPOINT | grep "pingheat_ping_sent_total" | grep -v "^#")
echo "  $SAMPLE2"

SENT2=$(echo "$SAMPLE2" | grep -oE '[0-9]+$')
UPTIME2=$(curl -s $ENDPOINT | grep "pingheat_uptime_seconds" | grep -oE '[0-9.]+$')
RTT2=$(curl -s $ENDPOINT | grep "pingheat_ping_last_rtt_ms" | grep -oE '[0-9.]+$')

echo "  Uptime: ${UPTIME2}s"
echo "  Last RTT: ${RTT2}ms"

# Calculate changes
echo ""
echo "📈 Changes:"
DIFF_SENT=$((SENT2 - SENT1))
DIFF_UPTIME=$(echo "$UPTIME2 - $UPTIME1" | bc)

echo "  Pings sent: +${DIFF_SENT} (${SENT1} → ${SENT2})"
echo "  Uptime: +${DIFF_UPTIME}s (${UPTIME1}s → ${UPTIME2}s)"
echo "  Last RTT changed: ${RTT1}ms → ${RTT2}ms"

# Verdict
echo ""
if [ "$DIFF_SENT" -gt 0 ]; then
    echo "✅ SUCCESS: Metrics are updating dynamically!"
    echo "   Rate: ~$(echo "scale=1; $DIFF_SENT / $DIFF_UPTIME" | bc) pings/second"
else
    echo "❌ FAIL: Metrics are NOT updating"
fi
