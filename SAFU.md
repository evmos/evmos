# Simple Arrangement for Funding Upload

## Overview

This Simple Arrangement for Funding Upload (the
“[SAFU](https://jumpcrypto.com/safu-creating-a-standard-for-whitehats/)” or
“Arrangement”) is intended as a simple yet extensible way to specify a
post-exploit policy for whitehats, particularly rewards and distributions. It is
based on the SAFU Framework designed by [Jump Crypto](https://jumpcrypto.com/).

This Arrangement attempts to address the following issues during active
vulnerabilities and post-exploits:

- **Legal uncertainty**: No clear grace period during which the hacker can
  declare themselves as a white hat. No formal guarantees as the team can decide
  at any time to take legal action or not.
- **Lack of clarity**: What to do with the funds secured from an exploit? Where
  should the white hat transfer the secured funds? Is there a
  compensation/reward promised for securing affected users’ funds?
- **Execution risk**: Conflicting proposals and stressful negotiation may arise
  during and after the exploit of a vulnerability, leading to additional
  confusion and uncertainty between the parties involved.

## Concepts

- **Dropbox Address (”Dropbox”)**: An address or contract to which funds taken
  from the protocol should be deposited. In the case of contracts, a Dropbox can
  be *automatic*, handling claims and rewards on a per-depositor basis without
  human input, or *conditional*, requiring additional input such as governance
  approvals or identity verification (KYC).
- **Deposit Interval:** the grace period in which a sender must deposit funds in
  the Dropbox after removing them from the protocol.
- **Claim Delay:** a minimum waiting period before a sender may claim rewards.
  We recommend at least 24 hours, during which the extent of an exploit will
  become clear.
- **Sender Claim Interval:** a maximum waiting period after which the protocol
  may reclaim the sender’s reward, to avoid leaving funds stranded in the
  contract.
- **Bounty Percent**: Pro-rata share of funds secured that are claimable by the
  whitehats.
- **Bounty Cap**: Maximum amount of tokens that a whitehat can claim after
  securing the vulnerable funds.

## Statement for Whitehats

Tharsis Labs Ltd. (the ”Team”) commits to not pursue legal action against white
hats who act in accordance with the Arrangement for vulnerabilities found in the
Evmos blockchain (chain ID: 9001).

### Timeline

The Team gives 48 hours to hackers (”Grace Period”) to deposit the funds to the
Dropbox from the moment they obtain the tokens from the exploited vulnerability.
After this time, the Team will assume that the hacker is acting maliciously and
against this Arrangement if they haven’t transferred the full amount of tokens.

Evmos guarantees that the claiming for the funds process to begin not after 30
days (Claim Delay) of the transfer or during the next upgrade. This is due to
the fact that transfers will need to be executed during an upgrade. We expect
this process to become automatic after a dedicated trustless Cosmos module is
incorporated for this purpose.

If the whitehat doesn’t reclaim the tokens transferred before the 30th day from
the transfer day (Sender Claim Interval), the tokens will be reclaimed and
transferred to the Evmos community pool (aka. community treasury).

### Reward Policy

Whitehats that secure vulnerable funds are able to claim 5% of the total funds
secured (Bounty Percent) up to a total of 250,000 EVMOS (Bounty Cap).

There is no minimum to the amount that can be secured. The reward white hack
hacker can secure from 1 atto EVMOS (1e-18 EVMOS or the equivalent unit of 1 wei
on Ethereum).

We encourage whitehat hackers to report undisclosed vulnerabilities using the
<security@evmos.org> email.

## Dropbox for Protocol Funds

The following Dropbox address is available on the Evmos blockchain for
transferring secured funds by whitehats:

|         | Bech32 Format                                | Hex Format                                 |
| ------- | -------------------------------------------- | ------------------------------------------ |
| Dropbox | evmos1c6jdy4gy86s69auueqwfjs86vse7kz3grxm9h2 | 0xc6A4d255043ea1A2F79CC81c9940FA6433eb0A28 |

While the original purpose for the Dropbox address is to primarily help secure
vulnerable EVMOS tokens, it can also be used as a general purpose escrow account
for whitehackers to help secure other tokens (Native or ERC20) that have been
exploited due to a vulnerability and thus declare behaviour in accordance to
this Agreement.

The Team offers to serve as mediator between the project exploited and the
whitehat hackers that have transferred secured funds to the Dropbox. However,
the Team is not responsible or liable for any reward payout result of the
negotiation between these two parties.

### Address Derivation

The Dropbox address corresponds a `ModuleAccount` address that is not controlled
by the team nor any individual. The module `ModuleAccount` address provided is
derived from the first 20 bytes of the SHA256 sum for the `“safu”` string, using
the following algorithm:

```bash
address = address(shaSum256([]byte("safu"))[:20]))
```

### Source Addresses

In the event of a vulnerability, the rewards for whitehats will be taken out
from the Dropbox account. The total amount claimable by each whitehat is defined
by the Bounty Percent and Bounty Cap using the following formula:

```bash
amount_claimable = min(bounty_cap, amount_secured * bounty_percent)
```

### Conditions for Claiming

**KYC/KYB Requirements**

The Agreement requires KYC/KYB to be done for all whitehats wanting a reward
valued above US$ 1,000. The information required (photographic ID, utility bill)
is assessed by Provenance (the “KYC Provider”). Whitehats that are business
entities will have to provide additional information (e.g., directors, owners).
Please anticipate that the KYC Provider might require documentation in English,
or in certified translations to it. The collection and assessment of this
information will be done by the KYC Provider.

## References

- [Jump Crypto: SAFU - Creating a Standard for Whitehats](https://jumpcrypto.com/safu-creating-a-standard-for-whitehats/)
