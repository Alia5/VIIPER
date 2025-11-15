#!/usr/bin/env sh
# Simple VIIPER API client for bash/sh/zsh using netcat
# Usage examples:
#   ./viiper-api.sh "bus/list"
#   ./viiper-api.sh -p 3242 -h 192.168.1.10 "bus/1/add xbox360"
# Environment overrides:
#   VIIPER_HOST, VIIPER_PORT, VIIPER_TIMEOUT

set -eu

HOST=${VIIPER_HOST:-localhost}
PORT=${VIIPER_PORT:-3242}
TIMEOUT=${VIIPER_TIMEOUT:-2}

usage() {
  cat <<EOF
Usage: $(basename "$0") [-h host] [-p port] [-t timeout] "command"

Send a single command to the VIIPER API TCP server and print the response.

Options:
  -h host      API server hostname or IP (default: $HOST)
  -p port      API server port (default: $PORT)
  -t timeout   Seconds for connect/idle timeout (default: $TIMEOUT)

Environment:
  VIIPER_HOST, VIIPER_PORT, VIIPER_TIMEOUT override the respective defaults.

Examples:
  $(basename "$0") "bus/list"
  $(basename "$0") -h 127.0.0.1 -p 3242 "bus/1/add xbox360"
EOF
}

while getopts "h:p:t:" opt; do
  case "$opt" in
    h) HOST="$OPTARG" ;;
    p) PORT="$OPTARG" ;;
    t) TIMEOUT="$OPTARG" ;;
    *) usage; exit 2 ;;
  esac
done
shift $((OPTIND - 1))

if [ "$#" -eq 0 ]; then
  usage
  exit 2
fi
CMD=$*

# Require netcat
if ! command -v nc >/dev/null 2>&1; then
  echo "Error: netcat (nc) is required but not found in PATH" >&2
  exit 127
fi

NC_QUIT=""
if nc -h 2>&1 | grep -q -- "-q "; then
  NC_QUIT="-q 1"
fi


if ! OUTPUT=$(printf '%s\n' "$CMD" | nc $NC_QUIT -w "$TIMEOUT" "$HOST" "$PORT"); then
  echo "Error: failed to connect to ${HOST}:${PORT} (is the VIIPER API running?)" >&2
  exit 1
fi

if [ -n "$OUTPUT" ]; then
  printf '%s\n' "$OUTPUT"
fi
