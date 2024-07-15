# Gaia Docker image

There's a `gaia` docker image built on a nightly basis, as well as for every
release tag, and pushed to `ghcr.io/cosmos/gaia`. It's built from the
[`Dockerfile`](./Dockerfile) in this directory.

The images contain statically compiled `gaiad` binaries running on an `alpine`
container. By default, `gaiad` runs as user `nonroot`, with UID/GUID `1025`.
The image exposes ports `26656,26657,1317,9090`. This is how the `gaiad` is
compiled:

```Dockerfile
RUN LEDGER_ENABLED=false LINK_STATICALLY=true BUILD_TAGS=muslc make build
```

Since the image has an entrypoint of `gaiad start`, you can use it to start a
node by mounting in a `.gaia` config directory. So, for instance, you can start
a `v17.3.0` node running a chain configured at `$HOME/.gaia` by running:

```bash
docker run --rm -it -v "$HOME/.gaia:/opt/gaia" ghcr.io/cosmos/gaia:v17.3.0 --home /opt/gaia
```

Of course, you can also use the images to just run generic gaia commands:

```bash
docker run --rm -it --entrypoint gaiad -v "$HOME/.gaia:/opt/gaia" ghcr.io/cosmos/gaia:v17.3.0 q tendermint-validator-set --home /opt/gaia
```

## Building

The images are built by workflow
[docker-push.yml](./.github/workflows/docker-push.yml). This workflow is
invoked on release as well as every night, and may be invoked manually by
people to build an arbitrary branch. It uses the `docker/metadata-action` to
decide how to tag the image, according to the following rules:

* If invoked via schedule, the image is tagged `nightly` and `main` (since it's a build of the `main` branch)
* If invoked from a release, including an rc, it is tagged with the release tag
* If invoked manually on a branch, it is tagged with the branch name

**NOTE:** To avoid surprising users, there is no `latest` tag generated.
