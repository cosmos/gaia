#!/bin/sh

# This builds docs using docusaurus.
# COMMIT=$(git rev-parse HEAD)
echo "building docusaurus main docs"
# (git clean -fdx && git reset --hard && git checkout $COMMIT)
npm ci && npm run build
mv build ~/output
echo "done building docusaurus main docs"
# echo $DOCS_DOMAIN > ~/output/CNAME