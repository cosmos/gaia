#!/usr/bin/env sh
set -euo pipefail
set -x

BINARY=/gaiad/${BINARY:-gaiad}
ID=${ID:-0}
LOG=${LOG:-gaiad.log}

if ! [ -f "${BINARY}" ]; then
	echo "The binary $(basename "${BINARY}") cannot be found. Please add the binary to the shared folder. Please use the BINARY environment variable if the name of the binary is not 'gaiad'"
	exit 1
fi

export GAIADHOME="/data/node${ID}/gaiad"

if [ -d "$(dirname "${GAIADHOME}"/"${LOG}")" ]; then
  "${BINARY}" --home "${GAIADHOME}" "$@" | tee "${GAIADHOME}/${LOG}"
else
  "${BINARY}" --home "${GAIADHOME}" "$@"
fi
