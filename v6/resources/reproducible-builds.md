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
git checkout v4.2.1
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
Version: v4.2.1
Commit: dbd8a6fb522c571debf958837f9113c56d418f6b
Files:
 29d219b0b120b3188bd7cd7249fc96b9  gaiad-v4.2.1-darwin-amd64
 80338d9f0e55ea8f6c93f2ec7d4e18d6  gaiad-v4.2.1-linux-amd64
 9bc77a512acca673ca1769ae67b4d6c7  gaiad-v4.2.1-linux-arm64
 c84387860f52178e2bffee08897564bb  gaiad-v4.2.1-windows-amd64.exe
 c25cca8ccceec06a6fabae90f671fab1  gaiad-v4.2.1.tar.gz
Checksums-Sha256:
 05e5b9064bac4e71f0162c4c3c3bff55def22ca016d34205a5520fef89fd2776  gaiad-v4.2.1-darwin-amd64
 ccda422cbda29c723aaf27653bcf0f6412e138eec33fba2b49de131f9ffbe2d2  gaiad-v4.2.1-linux-amd64
 95f89e8213cb758d12e1b0b631285938de822d04d2e25f399e99c0b798173cfd  gaiad-v4.2.1-linux-arm64
 7ef98f0041f1573f0a8601abad4a14b1c163f47481c7ba1954fd81ed423a6408  gaiad-v4.2.1-windows-amd64.exe
 422883ba43c96a6ea5ef9512d39321dd1356633c6a9505517b9c651788df4a7f  gaiad-v4.2.1.tar.gz
```
