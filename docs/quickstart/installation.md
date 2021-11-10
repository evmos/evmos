<!--
order: 1
-->

# Installation

Build and install the Hazlor binaries from source or using Docker. {synopsis}

## Pre-requisites

- [Install Go 1.17+](https://golang.org/dl/) {prereq}
- [Install jq](https://stedolan.github.io/jq/download/) {prereq}

## Install Go

::: warning
Hazlor is built using [Go](https://golang.org/dl/) version `1.17+`
:::

```bash
go version
```

:::tip
If the `hazlord: command not found` error message is returned, confirm that your [`GOPATH`](https://golang.org/doc/gopath_code#GOPATH) is correctly configured by running the following command:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

:::

## Install Binaries

::: tip
The latest {{ $themeConfig.project.name }} [version](https://github.com/hazlorlabs/hsc-chain/releases) is `{{ $themeConfig.project.binary }} {{ $themeConfig.project.latest_version }}`
:::

### GitHub

Clone and build {{ $themeConfig.project.name }} using `git`:

```bash
git clone https://github.com/hazlorlabs/hsc-chain.git
cd evmos
make install
```

Check that the `{{ $themeConfig.project.binary }}` binaries have been successfully installed:

```bash
hazlord version
```

### Docker

You can build {{ $themeConfig.project.name }} using Docker by running:

```bash
make docker-build
```

This will install the binaries on the `./build` directory. Now, check that the binaries have been
successfully installed:

```bash
hazlord version
```

### Releases

You can also download a specific release available on the {{ $themeConfig.project.name }} [repository](https://github.com/hazlorlabs/hsc-chain/releases) or via command line:

```bash
go install github.com/hazlorlabs/hsc@latest
```
