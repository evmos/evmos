# End-to-end Tests

# Structure

## `e2e` Package

The `e2e` package defines an integration testing suite used for full end-to-end
testing functionality. This package is decoupled from depending on the Evmos codebase.
It initializes the chains for testing via Docker files. As a result, the test suite may
provide the desired Evmos version to Docker containers during the initialization.
This design allows for the opportunity of testing chain upgrades in the future by providing
an older Evmos version to the container, performing the chain upgrade, and running the latest test suite.

The file e2e_suite_test.go defines the testing suite and contains the core
bootstrapping logic that creates a testing environment via Docker containers.
A testing network is created dynamically with 2 test validators.

The file e2e_test.go contains the actual end-to-end integration tests that
utilize the testing suite.

Currently, there is a single test in `e2e_test.go` to query the balances of a validator.

## `chain` Package

The `chain` package introduces the logic necessary for initializing a chain by creating a genesis
file and all required configuration files such as the `app.toml`. This package directly depends on the Evmos codebase.

# Running Locally

##### To build chain initializer:

```
make build-e2e-chain-init
```
- The produced binary is an entrypoint to the `evmos-e2e-chain-init:debug` image.

```
make docker-build-e2e-chain-init
```

##### To run the chain initialization container locally:

```
mkdir < path >
docker run -v < path >:/tmp/evmos-test evmos-e2e-chain-init:debug --data-dir=/tmp/evmos-test
sudo rm -r < path > # must be root to clean up
```
- runs a container with a volume mounted at < path > where all chain initialization files are placed.
- < path > must be absolute.
- `--data-dir` flag is needed for outputting the files into a directory inside the container

Example:
```
docker run -v /home/rama/test/:/chain evmos-e2e-chain-init:debug --data-dir /chain --chain-id evmos_9001-1
```

##### To build the debug Evmos image:

```
make docker-build-e2e-debug
```

##### Prepare for testing upgrades e2e:

Create the chain initializer docker image on the latest stable version of the software (before the upgrade).
Since we are testing upgrades `v3`-> `v4`, we need to initialize the genesis file with the `v3` version.
```
rm -rf build

git checkout v3.0.2

make build-e2e-chain-init

make docker-build-e2e-chain-init
```

The `v4` version should have an upgrade handler, now build the docker image. This docker image will be tagged as `debug`,
and will represent post upgrade Evmos node.

```
make docker-build-e2e-debug
```

The e2e test will first execute the chain_initializer and create the necessary files to run a node.
Then it will run two validators via docker with the tag provided (pre upgrade tag).


```e2e_setup_test.go#L161
	for i, val := range c.Validators {
		runOpts := &dockertest.RunOptions{
			Name:      val.Name,
			NetworkID: s.dkrNet.Network.ID,
			Mounts: []string{
				fmt.Sprintf("%s/:/evmos/.evmosd", val.ConfigDir),
			},
			Repository: "tharsishq/evmos",
			Tag:        "v3.0.2",  <-------------------- Upgrade this tag to reflect pre upgrade version
			Cmd: []string{
				"/usr/bin/evmosd",
				"start",
				"--home",
				"/evmos/.evmosd",
			},
		}
```

The node will run with the previous version (`v3`), and submit a proposal for upgrading to the post upgrade version.
Modify the proposal to reflect current upgrade.

```e2e_util_test.go#L129
			"tx", "gov", "submit-proposal",
			"software-upgrade", "v4.0.0",  <--- Update the upgrade currently in testing
			"--title=\"v4.0.0\"",
			"--description=\"v4 upgrade proposal\"",
			"--upgrade-height=75",
			"--upgrade-info=\"\"",
```
After block 75 is reached, it will destroy the previously used docker images, and will run the docker images with the `debug` tag.
This will execute the upgrade, and check that it was successful.

##### Run the e2e upgrade test:
Once the testing files have been updated, and the correct docker images have been built, run the testing suite.
```
make test-e2e
```