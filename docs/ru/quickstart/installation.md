<!--
order: 1
-->

# Installation

Build and install the Evmos binaries from source or using Docker. {synopsis}

## Pre-requisites

- [Install Go 1.17+](https://golang.org/dl/) {prereq}
- [Install jq](https://stedolan.github.io/jq/download/) {prereq}

## Install Go

::: warning
Evmos is built using [Go](https://golang.org/dl/) version `1.17+`
:::

```bash
go version
```

:::tip
If the `evmosd: command not found` error message is returned, confirm that your [`GOPATH`](https://golang.org/doc/gopath_code#GOPATH) is correctly configured by running the following command:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

:::

## Install Binaries

::: tip
The latest {{ $themeConfig.project.name }} [version](https://github.com/tharsis/evmos/releases) is `{{ $themeConfig.project.binary }} {{ $themeConfig.project.latest_version }}`
:::

### GitHub

Clone and build {{ $themeConfig.project.name }} using `git`:

```bash
git clone https://github.com/tharsis/evmos.git
cd evmos
make install
```

Check that the `{{ $themeConfig.project.binary }}` binaries have been successfully installed:

```bash
evmosd version
```

### Docker

You can build {{ $themeConfig.project.name }} using Docker by running:

```bash
make docker-build
```

This will install the binaries on the `./build` directory. Now, check that the binaries have been
successfully installed:

```bash
evmosd version
```

### Releases

You can also download a specific release available on the {{ $themeConfig.project.name }} [repository](https://github.com/tharsis/evmos/releases) or via command line:

```bash
go install github.com/tharsis/evmos@latest
```
