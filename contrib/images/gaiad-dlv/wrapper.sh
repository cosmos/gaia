#!/usr/bin/env sh
set -euo pipefail
set -x

DEBUG=${DEBUG:-0}
BINARY=/gaiad/${BINARY:-gaiad}
ID=${ID:-0}
LOG=${LOG:-gaiad.log}

if ! [ -f "${BINARY}" ]; then
	echo "The binary $(basename "${BINARY}") cannot be found. Please add the binary to the shared folder. Please use the BINARY environment variable if the name of the binary is not 'gaiad'"
	exit 1
fi

export GAIADHOME="/data/node${ID}/gaiad"

if [ "$DEBUG" -eq 1 ]; then
  dlv --listen=:2345 --continue --headless=true --api-version=2 --accept-multiclient exec "${BINARY}" -- --home "${GAIADHOME}" "$@"
elif [ "$DEBUG" -eq 1 ] && [ -d "$(dirname "${GAIADHOME}"/"${LOG}")" ]; then
  dlv --listen=:2345 --continue --headless=true --api-version=2 --accept-multiclient exec "${BINARY}" -- --home "${GAIADHOME}" "$@" | tee "${GAIADHOME}/${LOG}"
elif [ -d "$(dirname "${GAIADHOME}"/"${LOG}")" ]; then
  "${BINARY}" --home "${GAIADHOME}" "$@" | tee "${GAIADHOME}/${LOG}"
else
  "${BINARY}" --home "${GAIADHOME}" "$@"
fi