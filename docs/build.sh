#!/bin/sh

# This builds the docs.cosmos.network docs using docusaurus.
# Old documentation, is not migrated, but is still available at the appropriate release tag.

# Get the current commit hash, usually should be main.
COMMIT=$(git rev-parse HEAD)
mkdir -p ~/versioned_docs  ~/versioned_sidebars

# Build docs for each version tag in versions.json.
for version in $(jq -r .[] versions.json); do
    echo ">> Building docusaurus $version docs"
    (git clean -fdx && git reset --hard && git checkout $version && npm install && npm run docusaurus docs:version $version)
    mv ./versioned_docs/* ~/versioned_docs/
    mv ./versioned_sidebars/* ~/versioned_sidebars/
    echo ">> Finished building docusaurus $version docs"
done

# Build docs for $COMMIT that we started on.
echo ">> Building docusaurus main docs"
(git clean -fdx && git reset --hard && git checkout $COMMIT)
mv ~/versioned_docs ~/versioned_sidebars .
npm ci && npm run build
mv build ~/output

echo ">> Finished building docusaurus main docs"
exit 0