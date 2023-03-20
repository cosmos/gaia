#!/bin/bash

./contrib/scripts/run-gaia-v8.sh > v8.out 2>&1 &
./contrib/scripts/run-upgrade-commands.sh 15
./contrib/scripts/test_upgrade.sh 20 5 16 localhost