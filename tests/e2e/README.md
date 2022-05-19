# End-to-end Tests

# Structure

## `e2e` Package

The `e2e` package defines an integration testing suite used for full end-to-end
testing functionality. This package is decoupled from depending on the Osmosis codebase.
It initializes the chains for testing via Docker files. As a result, the test suite may
provide the desired Osmosis version to Docker containers during the initialization.
This design allows for the opportunity of testing chain upgrades in the future by providing
an older Osmosis version to the container, performing the chain upgrade, and running the latest test suite.

The file e2e_suite_test.go defines the testing suite and contains the core
bootstrapping logic that creates a testing environment via Docker containers.
A testing network is created dynamically with 2 test validators.

The file e2e_test.go contains the actual end-to-end integration tests that
utilize the testing suite.

Currently, there is a single test in `e2e_test.go` to query the balances of a validator.

## `chain` Package

The `chain` package introduces the logic necessary for initializing a chain by creating a genesis
file and all required configuration files such as the `app.toml`. This package directly depends on the Osmosis codebase.

## `upgrade` Package

The `upgrade` package starts chain initialization. In addition, there is a Dockerfile `init-e2e.Dockerfile`. 
When executed, its container produces all files necessary for starting up a new chain. 
These resulting files can be mounted on a volume and propagated to our production osmosis container to start the `osmosisd` service.

The decoupling between chain initialization and start-up allows to minimize the differences between our test suite and the production environment.

# Running Locally

##### To build the binary that initializes the chain:

```
make build-e2e-chain-init
```
- The produced binary is an entrypoint to the `osmosis-e2e-chain-init:debug` image.

##### To build the image for initializing the chain (`osmosis-e2e-chain-init:debug`):

```
make docker-build-e2e-chain-init
```

##### To run the chain initialization container locally:

```
mkdir < path >
docker run -v < path >:/tmp/osmo-test osmosis-e2e-chain-init:debug --data-dir=/tmp/osmo-test
sudo rm -r < path > # must be root to clean up
```
- runs a container with a volume mounted at < path > where all chain initialization files are placed.
- < path > must be absolute.
- `--data-dir` flag is needed for outputting the files into a directory inside the container

Example:
```
docker run -v /home/roman/cosmos/osmosis/tmp:/tmp/osmo-test osmosis-e2e-chain-init:debug --data-dir=/tmp/osmo-test
```

##### To build the debug Osmosis image:

```
make docker-build-e2e-debug
```
