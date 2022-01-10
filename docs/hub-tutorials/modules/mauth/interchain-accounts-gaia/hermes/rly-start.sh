#!/bin/bash
# Start the hermes relayer in multi-paths mode
echo "Starting hermes relayer..."
$HERMES_BINARY -c $CONFIG_DIR start

hermes -c rly-config.toml start
