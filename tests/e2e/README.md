# End-to-End Testing Suite

The End-to-End (E2E) testing suite provides an environment for running end-to-end tests on Evmos. It is used for testing chain upgrades, as it allows for initializing multiple Evmos chains with different versions.

### Quick Start

To run a chain upgrade test, execute:

```shell
make test-upgrade
```

This logic utilizes parameters that can be set manually(if necessary):

```shell
# flag to skip containers cleanup after upgrade
# should be set true with make test-e2e command if you need access to the node after upgrade
E2E_SKIP_CLEANUP := false

# version of initial evmos node that will be upgraded, tag e.g. 'v9.0.0'
INITIAL_VERSION

# version of upgraded evmos node that will replace the initial node, tag e.g. 'v9.1.0'
TARGET_VERSION

# mount point for upgraded node container, to mount new node version to previous node state folder
# by defaullt './build/.evmosd:/root/.evmosd'
# more info https://docs.docker.com/engine/reference/builder/#volume
MOUNT_PATH

# '--chain-id' evmos cli parameter, used to start nodes with specific chain-id and submit proposals
# by default 'evmos_9000-1'
CHAIN_ID
```

To test an upgrade to explicit target version and continue to run the upgraded node, run:

```shell
make test-e2e E2E_SKIP_CLEANUP=true INITIAL_VERSION=<tag> TARGET_VERSION=<tag>
```

### Upgrade Process

Testing a chain upgrade is a multi-step process:


1. Build a docker image for the evmos target version (local repo by default, if no explicit `TARGET_VERSION` provided as argument) (e.g. `v9.1.0`)
2. Run tests
3. The e2e test will first run an `INITIAL_VERSION` node container.
4. The node will submit, deposit and vote for an upgrade proposal for upgrading to the `TARGET_VERSION`.
5. After block `50` is reached, the test suite exports `/.evmosd` folder from docker container to local `build/` and than purge the container.
6. Suite will mount `TARGET_VERSION` node to local `build/` dir and start the node. Node will get upgrade information from `upgrade-info.json` and will execute the upgrade.

## Structure

### `e2e` Package

The `e2e` package defines an integration testing suite used for full end-to-end testing functionality. This package is decoupled from depending on the Evmos codebase. It initializes the chains for testing via Docker.
As a result, the test suite may provide the desired Evmos version to Docker containers during the initialization. This design allows for the opportunity of testing chain upgrades by providing an older Evmos version to the container, performing the chain upgrade, and running the latest test suite. Here's an overview of the files:

* `e2e_suite_test.go`: defines the testing suite and contains the core bootstrapping logic that creates a testing environment via Docker containers. A testing network is created dynamically with 2 test validators.

* `e2e_test.go`: contains the actual end-to-end integration tests that utilize the testing suite.

* `e2e_utils_test.go`: contains suite upgrade params loading logic.

### `upgrade` Package

The `e2e` package defines an upgrade `Manager` abstraction. Suite will utilize `Manager`'s functions to run different versions of evmos containers, propose, vote, delegate and query nodes.

* `manager.go`: defines core manager logic for running containers, export state and create networks.

* `govexec.go`: defines `gov-specific` exec commands to submit/delegate/vote through nodes `gov` module.

* `node.go`: defines `Node` strcuture responsible for setting node container parameters before run.

### Version retrieve

`TARGET_VERSION` by default retieved latest upgrade version from local codebase `evmos/app/upgrades` folder according to sevmver scheme.
If explicit `TARGET_VERSION` provided as argument, corresponding node container will be pulled from [dockerhub](https://hub.docker.com/r/tharsishq/evmos/tags).

`INITIAL_VERSION` retrieved as one version before the latest upgrade in `evmos/app/upgrades` correspondingly.

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

For interaction with upgraded set `SKIP_CLEANUP=true` flag on make command agruments and enter the container after upgrade finished:

```shell
docker exec -it <container-id> bash
```

If cleanup was skipped upgraded node container should be removed manually:

```shell
docker kill <container-id>
docker rm <container-id>
```
