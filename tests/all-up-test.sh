#!/bin/bash
set -eux
# the directory of this script, useful for allowing this script
# to be run with any PWD
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# builds the container containing various system deps
# also builds althea-chain once in order to cache Go deps, this container
# must be rebuilt every time you run this test because it pulls in the
# current repo state by reading the latest git commit during image creation
# if you want a faster solution use start chains and then run tests
# if you are running many tests on the same code set the NO_IMAGE_BUILD=1 env var
set +u
if [[ -z ${NO_IMAGE_BUILD} ]]; then
bash $DIR/build-container.sh
fi
set -u

# Remove existing container instance
set +e
docker rm -f althea_all_up_test_instance
set -e

NODES=4
set +u
TEST_TYPE=$1
set -u

# setup for Mac M1 Compatibility
PLATFORM_CMD=""
if [[ "$OSTYPE" == "darwin"* ]]; then
    if [[ -n $(sysctl -a | grep brand | grep "M1") ]]; then
       echo "Setting --platform=linux/amd64 for Mac M1 compatibility"
       PLATFORM_CMD="--platform=linux/amd64"; fi
fi

docker run --name althea_all_up_test_instance $PLATFORM_CMD --cap-add=NET_ADMIN -t althea-base /bin/bash /althea/tests/container-scripts/all-up-test-internal.sh $NODES $TEST_TYPE
