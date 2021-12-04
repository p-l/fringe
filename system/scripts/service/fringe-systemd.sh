#!/bin/bash -e

/usr/bin/fringe &
PID=$!
echo $PID > /var/lib/fringe/finge.pid

set +e
#attempts=0
#url="$PROTOCOL://127.0.0.1:9990/status"
#result=$(curl -k -s -o /dev/null "$url" -w %{http_code})
#while [ "${result:0:2}" != "20" ] && [ "${result:0:2}" != "30" ] && [ "${result:0:2}" != "40" ]; do
#  attempts=$(($attempts+1))
#  echo "Fringe web service at $url unavailable after $attempts attempts..."
#  sleep 1
#  result=$(curl -k -s -o /dev/null "$url" -w %{http_code})
#done
echo "Fringe service started"
set -e