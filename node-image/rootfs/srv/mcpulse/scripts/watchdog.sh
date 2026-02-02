#!/bin/sh
set -exu
. /etc/mcpulse/gateway.env

call_gateway() {
  curl -X POST "$GATEWAY_URL/api/callback" \
       -H "Authorization: Bearer $SERVER_TOKEN" \
       -d "{\"status\": \"$1\"}"
}

IDLE_MINUTES=0
SHUTDOWN_THRESHOLD=10
while true; do
  call_gateway ping
  sleep 60

  ! rc-service mcpulse-server status && break;

  CONNECTED_PLAYERS=$(mcprobe -host localhost -port 25565)
  IDLE_MINUTES=$(( CONNECTED_PLAYERS == 0 ? IDLE_MINUTES + 1 : 0 ))

  if [ "$IDLE_MINUTES" -ge "$SHUTDOWN_THRESHOLD" ]; then
    rc-service mcpulse-server stop

    # TODO: zip and backup:
    break
  fi
done

while true; do
  call_gateway killme
  sleep 120
done
