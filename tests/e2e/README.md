# End-to-End Testing Suite

The End-to-End (E2E) testing suite provides an environment for running end-to-end tests on Evmos. It is used for testing chain upgrades, as it allows for initializing multiple Evmos chains with different versions.

## Structure

### `e2e` Package

The `e2e` package defines an integration testing suite used for full end-to-end testing functionality. This package is decoupled from depending on the Evmos codebase. It initializes the chains for testing via Docker.
As a result, the test suite may provide the desired Evmos version to Docker containers during the initialization. This design allows for the opportunity of testing chain upgrades by providing an older Evmos version to the container, performing the chain upgrade, and running the latest test suite. Here's an overview of the files:

* `e2e_suite_test.go`: defines the testing suite and contains the core bootstrapping logic that creates a testing environment via Docker containers. A testing network is created dynamically with 2 test validators.
* `e2e_test.go`: contains the actual end-to-end integration tests that utilize the testing suite.

### `chain` Package

The `chain` package defines the logic necessary for initializing a chain by creating a genesis file and all required configuration files such as the `app.toml`. This package directly depends on the Evmos codebase.

## Chain Upgrades

e2e testing logic utilizes four parameters:

```shell
# PRE_UPGRADE_VERSION is the tag of evmos node which will be used to build initial validators containers
# by default it gets previous git tag from current, e.g. if current tag is v9.1.0 it will get v9.0.0 from git
PRE_UPGRADE_VERSION := $(shell git describe --abbrev=0 --tags `git rev-list --tags --skip=1 --max-count=1`)

# current latest tag
POST_UPGRADE_VERSION := $(shell git describe --tags --abbrev=0)

# flag to skip containers cleanup after upgrade
# should be set true with make test-e2e command if you need access to the node after upgrade
E2E_SKIP_CLEANUP := false

# flag for genasis migration
# should be set true manually if its necessary to migrate genesis
MIGRATE_GENESIS := false
```

every flag can be set manually with make command:

```shell
make test-e2e MIGRATE_GENESIS=true E2E_SKIP_CLEANUP=true PRE_UPGRADE_VERSION=v8.2.0
```

Testing a chain upgrade is a three step process:

1. Build a chain initializer docker image with pre-upgrade version (e.g. `v9.0.0`)
2. Build a docker image for the evmos post-upgrade version (e.g. `v9.1.0`)
3. Run tests

### Building `pre-upgrade` image

This logic included into `test-e2e` makefile command and runs before the testing.

Download evmos node of provided version tag and builds `chain_init` binary:

```docker
. . .
RUN git clone --depth 1 --branch $PRE_UPGRADE_VERSION https://github.com/evmos/evmos.git

WORKDIR /go/evmos/

RUN GO111MODULE=on go build -o ./build/chain_init ./tests/e2e/chain_init
. . .

```

Specific separated command also included into Makefile:

```shell
make docker-build-e2e-chain-init
```

### Create`post-upgrade` image

Builds container with current evmos node version.

This logic included into `test-e2e` makefile command and runs before the testing.

### Run upgrade tests

The e2e test will first execute the `chain_initializer` and create the necessary files to run a node. Then it will run two validators via docker with the tag provided (`PRE_UPGRADE_VERSION`).

The node will submit a proposal for upgrading to the `POST_UPGRADE_VERSION`.

After block `50` is reached, the test suite destroys the previously used docker images and runs the docker images with the `debug` tag. This will execute the upgrade, and check that it was successful.

```shell
make test-e2e
```

### Testing Results

Running the e2e test make script, will output the test results for each testing file. In case of an sucessfull upgrade the script will output like `ok  	github.com/evmos/evmos/v5/tests/e2e	174.137s`.

In case of test failure, the container wont be deleted. To analyze the error, run

```shell
# check containter id
docker ps -a

# get logs
docker logs <cointainer_id>
```

To rerun the tests, make sure to remove all docker containers first with(if you skiped cleanup or tests failed):

```
docker kill $(docker ps -aq)
docker rm $(docker ps -aq)
```