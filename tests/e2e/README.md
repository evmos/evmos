# End-to-End Testing Suite

The End-to-End (E2E) testing suite provides an environment for running end-to-end tests on Evmos. It is used for testing chain upgrades, as it allows for initializing multiple Evmos chains with different versions.

## Structure

### `e2e` Package

The `e2e` package defines an integration testing suite used for full end-to-end testing functionality. This package is decoupled from depending on the Evmos codebase. It initializes the chains for testing via Docker files. As a result, the test suite may provide the desired Evmos version to Docker containers during the initialization. This design allows for the opportunity of testing chain upgrades by providing an older Evmos version to the container, performing the chain upgrade, and running the latest test suite. Here's an overview of the files:

* `e2e_suite_test.go`: defines the testing suite and contains the core bootstrapping logic that creates a testing environment via Docker containers. A testing network is created dynamically with 2 test validators.
* `e2e_test.go`: contains the actual end-to-end integration tests that utilize the testing suite.

### `chain` Package

The `chain` package defines the logic necessary for initializing a chain by creating a genesis file and all required configuration files such as the `app.toml`. This package directly depends on the Evmos codebase.

## Chain Upgrades

Testing a chain upgrade is a three step process:

1. Build a chain initializer docker image with pre-upgrade version (e.g. `v3`)
2. Build a chain initializer docker image with post-upgrade version (e.g. `v4`)
3. Run tests on pre-upgrade version

### Create `pre-upgrade` image

Create the chain initializer docker image on the latest stable version of the software (before the upgrade). Since in this example we are testing an upgrade from `v3` to `v4`, we need to initialize the genesis file with the `v3` version.

```shell
# checkout pre-upgrade release branch
git checkout <pre-upgrade version>

# build chain initializer image
make build-e2e-chain-init # The produced binary is an entrypoint to the `evmos-e2e-chain-init:debug` image.
make docker-build-e2e-chain-init
```

### Create`post-upgrade` image

The `v4` version should have an upgrade handler, now build the docker image. This docker image will be tagged as `debug`,
and will represent post upgrade Evmos node.

```shell
# checkout post-upgrade release branch
git checkout <post-upgrade version>

make docker-build-debug
```

### Run upgrade tests

The e2e test will first execute the chain_initializer and create the necessary files to run a node. Then it will run two validators via docker with the tag provided (pre upgrade tag). Before running the test suite, you need to update the version tags that are relevant for the upgrade (e.g. `v3` -> `v4`):

In `e2e_suite_test.go#L114` update the validator version tag:

```go
	for i, val := range c.Validators {
		runOpts := &dockertest.RunOptions{
			Name:      val.Name,
			NetworkID: s.dkrNet.Network.ID,
			Mounts: []string{
				fmt.Sprintf("%s/:/evmos/.evmosd", val.ConfigDir),
			},
			Repository: "tharsishq/evmos",
			Tag:        "v3.0.2",  // <-------------------- Upgrade this tag to reflect pre upgrade version
			Cmd: []string{
				"/usr/bin/evmosd",
				"start",
				"--home",
				"/evmos/.evmosd",
			},
		}
```

The node will run with the previous version (`v3`), and submit a proposal for upgrading to the post upgrade version. Modify the proposal to reflect current upgrade in `e2e_util_test.go#L129`.

```go
			"tx", "gov", "submit-proposal",
			"software-upgrade", "v4.0.0", // <--- Update the upgrade currently in testing
			"--title=\"v4.0.0\"",
			"--description=\"v4 upgrade proposal\"",
			"--upgrade-height=75",
			"--upgrade-info=\"\"",
```
After block 75 is reached, the test suite destroys the previously used docker images and runs the docker images with the `debug` tag. This will execute the upgrade, and check that it was successful.

If the upgrade needs to migrate a genesis file to the new version, change the tag in `e2e_util_test.go#L317`:

```go
		Cmd: []string{
			"/usr/bin/evmosd",
			"--home",
			"/evmos/.evmosd",
			"migrate",
			"v4",           // <------ Update the migration version
			"/evmos/.evmosd/config/genesis.json",
			"--chain-id=evmos_9001-1",
		},
```

Once the testing files have been updated, and the correct docker images have been built, run the testing suite :

```shell
make test-e2e
```


### Testing Results

Running the e2e test make script, will output the test results for each testing file. In case of an sucessfull upgrade the script will output `ok  	github.com/tharsis/evmos/v4/tests/e2e	174.137s`.

In case of test failure, the container wont be deleted. To analyze the error, run

```shell
# check containter id
docker ps -a

# get logs
docker logs cointainerid
```

To rerun the tests, make sure to remove all docker containers first with:

```
docker kill $(docker ps -a -q)
docker rm $(docker ps -a -q)
```