<!--
order: 2
-->

# Hardhat: Deploying a Smart Contract

Learn how to deploy a simple Solidity-based smart contract to Evmos using the Hardhat environment {synopsis}

[Hardhat](https://hardhat.org/) is a flexible development environment for building Ethereum-based smart contracts.
It is designed with integrations and extensibility in mind

## Install Dependencies

Before proceeding, you need to install Node.js (we'll use v16.x) and the npm package manager.
You can download directly from [Node.js](https://nodejs.org/en/download/) or in your terminal:

<CodeGroup>
<CodeGroupItem title="Ubuntu">

```bash
curl -sL https://deb.nodesource.com/setup_16.x | sudo -E bash -

sudo apt install -y nodejs
```

</CodeGroupItem>
<CodeGroupItem title="MacOS">

```bash
# You can use homebrew (https://docs.brew.sh/Installation)
$ brew install node

# Or you can use nvm (https://github.com/nvm-sh/nvm)
$ nvm install node
```

</CodeGroupItem>
</CodeGroup>

You can verify that everything is installed correctly by querying the version for each package:

```bash
$ node -v
...

$ npm -v
...
```

## Create Hardhat Project

To create a new project, navigate to your project directory and run:

```bash
$ npx hardhat

888    888                      888 888               888
888    888                      888 888               888
888    888                      888 888               888
8888888888  8888b.  888d888 .d88888 88888b.   8888b.  888888
888    888     "88b 888P"  d88" 888 888 "88b     "88b 888
888    888 .d888888 888    888  888 888  888 .d888888 888
888    888 888  888 888    Y88b 888 888  888 888  888 Y88b.
888    888 "Y888888 888     "Y88888 888  888 "Y888888  "Y888

👷 Welcome to Hardhat v2.9.3 👷‍

? What do you want to do? …
  Create a basic sample project
❯ Create an advanced sample project
  Create an advanced sample project that uses TypeScript
  Create an empty hardhat.config.js
  Quit
```

Following the prompts should create a new project structure in your directory.
Consult the [Hardhat config page](https://hardhat.org/config/) for a list of configuration options
to specify in `hardhat.config.js`.
Most importantly, you should set the `defaultNetwork` entry to point to your desired JSON-RPC network:

<CodeGroup>
<CodeGroupItem title="Local Node">

```javascript
module.exports = {
  defaultNetwork: "local",
  networks: {
    hardhat: {
    },
    local: {
      url: "http://localhost:8545/",
      accounts: [privateKey1, privateKey2, ...]
    }
  },
  ...
}
```

</CodeGroupItem>
<CodeGroupItem title="Testnet">

```javascript
module.exports = {
  defaultNetwork: "testnet",
  networks: {
    hardhat: {
    },
    testnet: {
      url: "https://eth.bd.evmos.dev:8545",
      accounts: [privateKey1, privateKey2, ...]
    }
  },
  ...
}
```

</CodeGroupItem>
</CodeGroup>

* To get the value for privateKey:
    * MetaMask -> Account Details -> Export Private Key -> add '0x' as prefix -> `privateKey1`

To ensure you are targeting the correct network,
you can query for a list of accounts available to you from your default network provider:

```bash
$ npx hardhat accounts
0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
0x70997970C51812dc3A010C7d01b50e0d17dc79C8
0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC
0x90F79bf6EB2c4f870365E785982E1f101E93b906
...
```

* To make `accounts` command work in recent hardhat (`npx hardhat --version // 2.12.5`)
add this to `hardhat.config.js`

```javascript
task("accounts", "Prints the list of accounts", async () => {
  const accounts = await ethers.getSigners();

  for (const account of accounts) {
    console.log(account.address);
  }
});
```

## Deploying a Smart Contract

You will see that a default smart contract, written in Solidity,
has already been provided under `contracts/Greeter.sol`:

```javascript
pragma solidity ^0.8.0;

import "hardhat/console.sol";

contract Greeter {
    string private greeting;

    constructor(string memory _greeting) {
        console.log("Deploying a Greeter with greeting:", _greeting);
        greeting = _greeting;
    }

    function greet() public view returns (string memory) {
        return greeting;
    }

    function setGreeting(string memory _greeting) public {
        console.log("Changing greeting from '%s' to '%s'", greeting, _greeting);
        greeting = _greeting;
    }
}
```

This contract allows you to set and query a string `greeting`.
Hardhat also provides a script to deploy smart contracts to a target network;
this can be invoked via the following command, targeting your default network:

```bash
npx hardhat run scripts/deploy.js
```

Hardhat also lets you manually specify a target network via the `--network <your-network>` flag:

<CodeGroup>
<CodeGroupItem title="Local Node">

```bash
npx hardhat run --network {{ $themeConfig.project.rpc_url_local }} scripts/deploy.js
```

</CodeGroupItem>
<CodeGroupItem title="Testnet">

```bash
npx hardhat run --network {{ $themeConfig.project.rpc_url_testnet }} scripts/deploy.js
```

</CodeGroupItem>
</CodeGroup>

Finally, try running a Hardhat test:

```bash
$ npx hardhat test
Compiling 1 file with 0.8.4
Compilation finished successfully


  Greeter
Deploying a Greeter with greeting: Hello, world!
Changing greeting from 'Hello, world!' to 'Hola, mundo!'
    ✓ Should return the new greeting once it's changed (803ms)


  1 passing (805ms)
```
