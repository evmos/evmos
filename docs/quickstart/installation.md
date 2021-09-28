<!--
order: 1
-->

# Installation

Build and install the Evmos binaries from source or using Docker. {synopsis}

## Pre-requisites

- [Install Go 1.17+](https://golang.org/dl/) {prereq}
- [Install jq](https://stedolan.github.io/jq/download/) {prereq}

## Install Binaries

### GitHub

Clone and build Evmos using `git`:

```bash
git clone https://github.com/tharsis/evmos.git
cd evmos
make install
```

Check that the binaries have been successfully installed:

```bash
evmosd version
```

### Docker

You can build Evmos using Docker by running:

```bash
make docker-build
```

This will install the binaries on the `./build` directory. Now, check that the binaries have been
successfully installed:

```bash
evmosd version
```

### Releases

You can also download a specific release available on the Evmos [repository](https://github.com/tharsis/evmos/releases) or via command line:

```bash
go install github.com/tharsis/evmos@latest
```
