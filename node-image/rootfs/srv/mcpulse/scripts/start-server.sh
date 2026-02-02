#!/bin/sh
set -exu
. .mcpulse.env

JAVA_BIN=$(command -v java)
if [ -n "$JAVA_VERSION" ]; then
  BIN="/usr/lib/jvm/java-${JAVA_VERSION}-openjdk/bin/java"
  [ -x "$BIN" ] && JAVA_BIN="$BIN" || echo "Java ${JAVA_VERSION} not found. Falling back to: ${JAVA_BIN}" >&2
fi

TOTAL_MEM_KB=$(awk '/MemTotal/ {print $2}' /proc/meminfo)
AVAIL_MB=$(( (TOTAL_MEM_KB / 1024) * 75 / 100 ))
FLAGS="-Xms${AVAIL_MB}M -Xmx${AVAIL_MB}M"
set -f
set -- $JVM_FLAGS
while [ $# -gt 0 ]; do
  case "$1" in
    java)
      shift
      ;;
    -jar)
      shift
      [ $# -gt 0 ] && shift
      ;;
    -Xms*|-Xmx*|--nogui)
      shift
      ;;
    *)
      FLAGS="$FLAGS $1"
      shift
      ;;
  esac
done

BCKP_CODE=$(curl -s -w "%{response_code}" -o server.zip \
                -H "Authorization: Bearer $BACKUP_TOKEN" \
                "$GATEWAY_URL/api/backup")
if [ "$BCKP_CODE" -eq 404 ]; then
  echo "!!! Generating a new world !!!"
elif [ "$BCKP_CODE" -eq 200 ]; then
  unzip -o -q server.zip
else
  echo "Error checking backup status. Gateway returned ${BCKP_CODE}"
  exit 1
fi

exec "$JAVA_BIN" $FLAGS -jar "./$SERVER_JAR" $SERVER_ARGS --nogui
