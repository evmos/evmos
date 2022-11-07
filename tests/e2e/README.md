# End-to-End Testing Suite

The End-to-End (E2E) testing suite provides an environment for running end-to-end tests on Evmos. It is used for testing chain upgrades, as it allows for initializing multiple Evmos chains with different versions.

## Structure

### `e2e` Package

The `e2e` package defines an integration testing suite used for full end-to-end testing functionality. This package is decoupled from depending on the Evmos codebase. It initializes the chains for testing via Docker.
As a result, the test suite may provide the desired Evmos version to Docker containers during the initialization. This design allows for the opportunity of testing chain upgrades by providing an older Evmos version to the container, performing the chain upgrade, and running the latest test suite. Here's an overview of the files:

* `e2e_suite_test.go`: defines the testing suite and contains the core bootstrapping logic that creates a testing environment via Docker containers. A testing network is created dynamically with 2 test validators.
* `e2e_test.go`: contains the actual end-to-end integration tests that utilize the testing suite.

### `upgrade` Package

The `e2e` package defines an upgrade `Manager` abstraction. Suite will utilize `Manager`'s functions to run different versions of evmos containers, propose, vote and delegate.

* `manager.go`: defines core manager logic for running containers, export state and create networks.

* `govexec.go`: defines `gov-specific` exec commands to submit/delegate/vote through nodes `gov` module.

## Chain Upgrades

e2e testing logic utilizes three parameters:

```shell
# INITIAL_VERSION is the tag of the evmos node which will be used to build the initial validators container.
# By default the previous git tag is retrieved from `current`, e.g. if the current tag is `v9.1.0` it will get `v9.0.0` from git
INITIAL_VERSION := $(shell git describe --abbrev=0 --tags `git rev-list --tags --skip=1 --max-count=1`)

# TARGET_VERSION  is the tag to upgrade to. By default, this is the  current latest tag
TARGET_VERSION := $(shell git describe --abbrev=0 --tags `git rev-list --tags --max-count=1`)

# E2E_SKIP_CLEANUP is a flag to skip the container cleanup after an upgrade. It should be set to `true` if you need access to the node after the upgrade.
# should be set true with make test-e2e command if you need access to the node after upgrade
E2E_SKIP_CLEANUP := false
```

every flag can be set manually with make command:

```shell
make test-e2e E2E_SKIP_CLEANUP=true INITIAL_VERSION=v8.2.0
```

Testing a chain upgrade is a three step process:

1. Build a initial node version docker container (e.g. `v9.0.0`)
2. Build a docker image for the evmos target version(local repo by default) (e.g. `v9.1.0`)
3. Run tests

### Run upgrade tests

The e2e test will first run a `INITIAL_VERSION` node.

The node will submit, deposit and vote for an upgrade proposal for upgrading to the `TARGET_VERSION`.

After block `50` is reached, the test suite exports `/.evmosd` folder from docker container to local `build/` and than purge the container.

Suite will mount `TARGET_VERSION` node to local `build/` dir and start the node. Node will get upgrade information from `upgrade-info.json` and will execute the upgrade.

### Version retrieve

`INITIAL_VERSION` and `TARGET_VERSION` are retrieved from git tags by default with the following commands:

```shell
# INITIAL_VERSION
git describe --abbrev=0 --tags `git rev-list --tags --skip=1 --max-count=1`

# TARGET_VERSION
git describe --abbrev=0 --tags `git rev-list --tags --max-count=1`
```

If `Makefile` command cannot get the tags for some reason (i.e. you have no tag for the local branch and want to upgrade from a specific version to a local node etc), versions should be specified manually:

```shell
make test-e2e INITIAL_VERSION=<version> TARGET_VERSION=<version>
```

`TARGET_VERSION` used as a software upgrade version in proposal and must match the version in `upgrades` package.

### Testing Results

The `make test-upgrade` script will output the test results for each testing file. In case of a successful upgrade, the script will print the following output (example):

```log
ok  	github.com/evmos/evmos/v9/tests/e2e	174.137s.
```

To get containers logs run:

```shell
# check containters
docker ps -a
```

Container names will be listed as follows:

```log
CONTAINER ID   IMAGE
9307f5485323   evmos:local    <-- upgraded node
f41c97d6ca21   evmos:v9.0.0   <-- initial node
```

To get containers logs, run:

```shell
docker logs <container-id>
```

For interaction with upgraded node container/cli, set `SKIP_CLEANUP=true` on make command agruments and enter the container after upgrade finished:

```shell
docker exec -it <container-id> bash
```

To rerun the tests, make sure to remove all docker containers first with(if you skiped cleanup or tests failed):

```shell
docker kill $(docker ps -aq)
docker rm $(docker ps -aq)
```