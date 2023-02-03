<!--
order: 3
-->

# Simple Arrangement for Funding Upload (SAFU)

Learn about the Simple Arrangement for Funding Upload (SAFU)
on Evmos {synopsis}

## Overview

<!-- markdown-link-check-disable-next-line -->
The [Simple Arrangement for Funding Upload (the "SAFU")](https://github.com/evmos/evmos/tree/main/SAFU.pdf)
outlines the post-exploit policy for active vulnerabilities in the Evmos blockchain.
The SAFU is intended for white hat hackers
and outlines the process for returning funds and calculating rewards
for vulnerabilities found in the network.
In summary, the SAFU states the following:

* Hackers are not at risk of legal action if they act in accordance
  with the SAFU.
* Hackers have a Grace Period to return any exploited funds
  to a specified dropbox address and can claim a reward of
  a Bounty Percent of the total funds secured up to the Bounty Cap.
* The rewards are distributed during the next upgrade of the network.
* If the reward is valued above a specified threshold amount,
  white hat hackers should go through
  a Know Your Clients/Know Your Business (KYC/KYB) process.
* Exploiting vulnerabilities for malicious purposes
  will make a hacker ineligible for any rewards.
* White hat hackers are not entitled to any rewards from the team or network
  for funds from "Out of Scope Projects" (other projects that were exploited
  by hackers but do not have their own SAFU program).

For more information,
visit [the SAFU agreement](https://github.com/evmos/evmos/tree/main/SAFU.pdf).<!-- markdown-link-check-disable-line -->

## Dropbox Address

The Dropbox Address is an address to which funds are taken from
the protocol should be deposited.
In the event of a bounty distribution,
the bounty for white hat hackers will be paid out
from the account balance of this address.

::: tip
The dropbox address is not controlled by the team
or any individual, it is controlled by the Evmos protocol.
:::

The following dropbox address is available on the Evmos blockchain:

**Dropbox Address in Bech32 Format**:

```shell
evmos1c6jdy4gy86s69auueqwfjs86vse7kz3grxm9h2
```

**Dropbox Address in Hex Format**:

```shell
0xc6A4d255043ea1A2F79CC81c9940FA6433eb0A28
```

### Address Derivation

The dropbox address provided above is derived cryptographically from the
first 20 bytes of the SHA256 sum for the `“safu”` string,
using the following algorithm:

```shell
address = shaSum256([]byte("safu"))[:20])
```

## How To Secure Vulnerable Funds

Within the Grace Period of a hack,
white hats should secure the funds by transferring them to the dropbox address.

## How To Claim The Reward

Rewards distribution will be done manually on the next chain upgrade.
If the reward is valued above a certain threshold amount,
white hat hackers should go through a
Know Your Clients/Know Your Business (KYC/KYB) process.

## Security recommendations for dApps

As previously stated, rewards for secured funds from hacked dApps
are not included in the protocol's SAFU.
For such a case, we encourage all dApps on Evmos
to have their own SAFU implementation.
We recommend taking the [SAFU.sol](https://github.com/JumpCrypto/Safu/)
contract implementation from Jump Crypto as a reference.
