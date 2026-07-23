#!/usr/bin/env bash
set -euo pipefail

# Safe defaults for running sdexmon
export HORIZON_URL="${HORIZON_URL:-https://horizon.stellar.org}"
export DEBUG="${DEBUG:-false}"

# Set terminal window title
printf '\033]0;sdexmon\007'

# Run the actual binary
exec sdexmon "$@"
