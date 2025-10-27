#!/usr/bin/env bash
set -euo pipefail

# Safe defaults for running sdexmon
export HORIZON_URL="${HORIZON_URL:-https://horizon.stellar.org}"
export DEBUG="${DEBUG:-true}"

# Set terminal window title
printf '\033]0;sdexmon\007'

# Set fixed terminal size (140 columns x 60 rows)
if command -v tput >/dev/null 2>&1; then
  printf '\e[8;60;140t'
fi

# Run the actual binary
exec sdexmon "$@"
