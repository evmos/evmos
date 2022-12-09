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

if target node version failed to start, caintainers `[error stream]` and `[output stream]` will be printed from container to local terminal:

```log
            	Error:      	Received unexpected error:
            	            	can't start evmos node, container exit code: 2

            	            	[error stream]:

            	            	7:03AM INF Unlocking keyring
            	            	7:03AM INF starting ABCI with Tendermint
            	            	panic: invalid minimum gas prices: invalid decimal coin expression: 0aevmos

            	            	goroutine 1 [running]:
            	            	github.com/cosmos/cosmos-sdk/baseapp.SetMinGasPrices({0xc0013563e7?, 0xc00163a3c0?})
            	            		github.com/cosmos/cosmos-sdk@v0.46.5/baseapp/options.go:29 +0xd9
            	            	main.appCreator.newApp({{{0x3399b40, 0xc000ec1db8}, {0x33ac0f8, 0xc0011314e0}, {0x33a2920, 0xc000ed2b80}, 0xc0000155f0}}, {0x3394520, 0xc001633bc0}, {0x33a5cc0, ...}, ...)
            	            		github.com/evmos/evmos/v10/cmd/evmosd/root.go:243 +0x2ca
            	            	github.com/evmos/ethermint/server.startInProcess(_, {{0x0, 0x0, 0x0}, {0x33b7490, 0xc001784c30}, 0x0, {0x7fff50b37f3d, 0xc}, {0x33ac0f8, ...}, ...}, ...)
            	            		github.com/evmos/ethermint@v0.20.0-rc2/server/start.go:304 +0x9c5
            	            	github.com/evmos/ethermint/server.StartCmd.func2(0xc001620600?, {0xc001745bd0?, 0x0?, 0x1?})
            	            		github.com/evmos/ethermint@v0.20.0-rc2/server/start.go:123 +0x1ec
            	            	github.com/spf13/cobra.(*Command).execute(0xc001620600, {0xc001745bb0, 0x1, 0x1})
            	            		github.com/spf13/cobra@v1.6.1/command.go:916 +0x862
            	            	github.com/spf13/cobra.(*Command).ExecuteC(0xc00160e000)
            	            		github.com/spf13/cobra@v1.6.1/command.go:1044 +0x3bd
            	            	github.com/spf13/cobra.(*Command).Execute(...)
            	            		github.com/spf13/cobra@v1.6.1/command.go:968
            	            	github.com/spf13/cobra.(*Command).ExecuteContext(...)
            	            		github.com/spf13/cobra@v1.6.1/command.go:961
            	            	github.com/cosmos/cosmos-sdk/server/cmd.Execute(0x2170d50?, {0x26d961f, 0x6}, {0xc00112c490, 0xd})
            	            		github.com/cosmos/cosmos-sdk@v0.46.5/server/cmd/execute.go:36 +0x20f
            	            	main.main()
            	            		github.com/evmos/evmos/v10/cmd/evmosd/main.go:20 +0x45


            	            	[output stream]:

            	Test:       	TestIntegrationTestSuite/TestUpgrade
            	Messages:   	can't mount and run upgraded node container
```

To get all containers run:

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

To access containers logs directly, run:

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
