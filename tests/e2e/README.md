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

* `govexec.go`: defines `gov-specific` exec commands to submit/delegate/vote thru nodes `gov` module.

## Chain Upgrades

e2e testing logic utilizes four parameters:

```shell
# INITIAL_VERSION is the tag of evmos node which will be used to build initial validators containers
# by default it gets previous git tag from current, e.g. if current tag is v9.1.0 it will get v9.0.0 from git
INITIAL_VERSION := $(shell git describe --abbrev=0 --tags `git rev-list --tags --skip=1 --max-count=1`)

# current latest tag
TARGET_VERSION := $(shell git describe --abbrev=0 --tags `git rev-list --tags --max-count=1`)

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

1. Build a initial node version docker container (e.g. `v9.0.0`)
2. Build a docker image for the evmos target version(local repo by default) (e.g. `v9.1.0`)
3. Run tests

### Run upgrade tests

The e2e test will first run a `INITIAL_VERSION` node.

The node will submit, deposit and vote for an upgrade proposal for upgrading to the `TARGET_VERSION`.

After block `50` is reached, the test suite exports `/.evmosd` folder from docker container to local `build/` and than purge the container.

Suite will mount `TARGET_VERSION` node to local `build/` dir and start the node. Node will get upgrade information from `upgrade-info.json` and will execute the upgrade.

```shell
make test-e2e
```

### Testing Results

Running the e2e test make script, will output the test results for each testing file. In case of an sucessfull upgrade the script will output like `ok  	github.com/evmos/evmos/v9/tests/e2e	174.137s`.

In case of test failure, the container wont be deleted. To analyze the error, run

```shell
# check containter id
docker ps -a

# get logs
docker logs <cointainer_id>
```

For interaction with upgraded node container/cli, set `SKIP_CLEANUP=true` on make command agruments and enter the container after upgrade finished:

```
docker exec -it <container-id> bash
```

To rerun the tests, make sure to remove all docker containers first with(if you skiped cleanup or tests failed):

```
docker kill $(docker ps -aq)
docker rm $(docker ps -aq)
```