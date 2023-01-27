<!--
order: 3
-->

# Simple Arrangement for Funding Upload (SAFU)

Learn about the Simple Arrangement for Funding Upload (SAFU) on Evmos {synopsis}

## Overview

The [Simple Arrangement for Funding Upload (the "SAFU")](https://docs.google.com/document/d/1kyPn-uRQnCOjeCHjH6IRGGDOTMVHMLeQXeJql9MXnkw/edit#heading=h.vcatw1yk8om7) outlines the post-exploit policy for active vulnerabilities in the Evmos blockchain.
The SAFU is intended for white hat hackers and outlines the process for returning funds and calculating rewards for vulnerabilities found in the network.
In summary, the SAFU states the following:

* Hackers are not at risk of legal action if they act in accordance with the SAFU.
* Hackers have 48 hours to return any exploited funds to a specified Dropbox address and can claim a reward of 5% of the total funds secured up to a maximum of 750,000 Tokens.
* The rewards are distributed during the next upgrade of the network.
* If the reward is valued above $100,000, white hat hackers should go through a Know Your Clients/Know Your Business (KYC/KYB) process.
* Exploiting vulnerabilities for malicious purposes will make a hacker ineligible for any rewards.
* White hat hackers are not entitled to any rewards from the Team or Network for funds from "Out of Scope Projects" (other projects that were exploited by hackers but do not have their own SAFU program).

For more information, visit [the SAFU agreement](https://docs.google.com/document/d/1kyPn-uRQnCOjeCHjH6IRGGDOTMVHMLeQXeJql9MXnkw/edit#heading=h.vcatw1yk8om7).

## Dropbox address

The Dropbox address is an address to which funds taken from the protocol should be deposited.
In the event of a bounty distribution, the bounty for white hat hackers will be paid out from the account balance of this address.

::: tip
The Dropbox address is not controlled by the Team or any individual, it is controlled by the Evmos protocol.
:::

The following Dropbox address is available on the Evmos blockchain:

**Dropbox Address in Bech32 Format**: 

```shell
evmos1c6jdy4gy86s69auueqwfjs86vse7kz3grxm9h2
```

**Dropbox Address in Hex Format**: 

```shell
0xc6A4d255043ea1A2F79CC81c9940FA6433eb0A28
```

### Address Derivation

The Dropbox address provided above is derived cryptographically from the first 20 bytes of the SHA256 sum for the “safu” string, using the following algorithm:

```shell
address = shaSum256([]byte("safu"))[:20])
```

## How to secure vulnerable funds

Within the first 48 hours of a hack, hackers should secure the funds by transferring them to the Dropbox address. 

## How to claim the reward

Rewards distribution will be done manually on the next chain upgrade.
If the reward is valued above $100,000, white hat hackers should go through a Know Your Clients/Know Your Business (KYC/KYB) process.

## Security recommendations for dApps

As previously stated, rewards for secured funds from hacked dApps are not included in the protocol's SAFU.
For such a case, we encourage all dApps on Evmos to have their own SAFU implementation.
We recommend taking the [SAFU.sol](https://github.com/JumpCrypto/Safu/) contract implementation from Jump Crypto as a reference.
