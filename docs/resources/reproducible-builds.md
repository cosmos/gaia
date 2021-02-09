<!--
order: 6
-->

# Build Gaia Deterministically

The [Tendermint rbuilder Docker image](https://github.com/tendermint/images/tree/master/rbuilder) provides a deterministic build environment that is used to build Cosmos SDK applications. It provides a way to be reasonably sure that the executables are really built from the git source. It also makes sure that the same, tested dependencies are used and statically built into the executable.

## Prerequisites

Make sure you have [Docker installed on your system](https://docs.docker.com/get-docker/).

All the following instructions have been tested on *Ubuntu 18.04.2 LTS* with *docker 20.10.2*.

## Build

Clone `gaia`:

```
git clone https://github.com/cosmos/gaia.git
```

Checkout the commit, branch, or release tag you want to build:

```
cd gaia/
git checkout v3.0.0
```

The buildsystem supports and produces binaries for the following architectures:
* **darwin/amd64**
* **linux/amd64**
* **linux/arm64**
* **windows/amd64**

Run the following command to launch a build for all supported architectures:

```
make distclean build-reproducible
```

The build system generates both the binaries and deterministic build report in the `artifacts` directory.
The `artifacts/build_report` file contains the list of the build artifacts and their respective checksums, and can be used to verify
build sanity. An example of its contents follows:

```
App: gaiad
Version: 2.0.12-20-gfc0171b
Commit: fc0171b00662fb43df12955378ed8b0c5db85229
Files:
 bee04f003adcc2a1848bcc4ec6dc6731  gaiad-2.0.12-20-gfc0171b-darwin-amd64
 ff1edcd44f6ff7b3746d211eceb8f3f5  gaiad-2.0.12-20-gfc0171b-linux-amd64
 8d6109dc1e1c59b2ffa9660a49bb54e1  gaiad-2.0.12-20-gfc0171b-linux-arm64
 3183eec0ae71da9d8b68e0ba2986b885  gaiad-2.0.12-20-gfc0171b-windows-amd64.exe
 8f26db0add97a3ac1e038b0a8dc3ffb3  gaiad-2.0.12-20-gfc0171b.tar.gz
Checksums-Sha256:
 c08d6bf03ca71254b24e8eda54dfcbf82ef671891b283ac194b6633292792324  gaiad-2.0.12-20-gfc0171b-darwin-amd64
 8c85b5ab2f3c4d50a53d97448d6eab28a3b3e9da1b92616cb478418bc8096f5a  gaiad-2.0.12-20-gfc0171b-linux-amd64
 66606c5cc82794a7713d50364ce9f0b3e582774b8fc8fb5851db933f98f661c2  gaiad-2.0.12-20-gfc0171b-linux-arm64
 1633968dbd987f1a1e2f90820ec78a39984eeae1fb97e0240d1f04909761bdb5  gaiad-2.0.12-20-gfc0171b-windows-amd64.exe
 8b7e34474185d83c6333e1f08f4c35459e95ec06bb8a246fd14bab595427ae9d  gaiad-2.0.12-20-gfc0171b.tar.gz
```
