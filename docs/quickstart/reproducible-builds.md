<!--
order: 3
-->

# Deterministic Builds

Build the `evmosd` binary deterministically using Docker. {synopsis}

## Pre-requisites

- [Install Docker](https://docs.docker.com/get-docker/) {prereq}

## Introduction

The [Tendermint rbuilder Docker image](https://github.com/tendermint/images/tree/master/rbuilder) provides a deterministic build environment that is used to build Cosmos SDK applications. It provides a way to be reasonably sure that the executables are really built from the git source. It also makes sure that the same, tested dependencies are used and statically built into the executable.

::: tip
All the following instructions have been tested on *Ubuntu 18.04.2 LTS* with *Docker 20.10.2*.
:::

## Build with Docker

Clone `evmos`:

``` bash
git clone git@github.com:tharsis/evmos.git
```

Checkout the commit, branch, or release tag you want to build (eg `v0.4.0`):

```bash
cd evmos/
git checkout v0.4.0
```

The buildsystem supports and produces binaries for the following architectures:

* **linux/amd64**

Run the following command to launch a build for all supported architectures:

```bash
make distclean build-reproducible
```

The build system generates both the binaries and deterministic build report in the `artifacts` directory.
The `artifacts/build_report` file contains the list of the build artifacts and their respective checksums, and can be used to verify
build sanity. An example of its contents follows:

```
App: evmosd
Version: 0.4.0
Commit: b7e46982d1dc2d4c34fcd3b52f1edfd2e589d370
Files:
 7594279acff34ff18ea9d896d217a6db  evmosd-0.4.0-linux-amd64
 c083e812acbfa7d6e02583386b371b93  evmosd-0.4.0.tar.gz
Checksums-Sha256:
 d087053050ce888c21d26e40869105163c5521cb5b291443710961ac0c892e81  evmosd-0.4.0-linux-amd64
 6ca3e5e40240f5e433088fd9b7370440f3f94116803934c21257e1c78fb9653d  evmosd-0.4.0.tar.gz
```
