<!--
order: 1
-->

# Concepts

## Inflation

In a Proof of Stake (PoS) blockchain, inflation is used as a tool to incentivize
participation in the network. Inflation creates and distributes new tokens to
participants who can use their tokens to either interact with the protocol or
stake their assets to earn rewards and vote for governance proposals.

Especially in an early stage of a network, where staking rewards are high and
there are fewer possibilities to interact with the network, inflation can be
used as the major tool to incentivize staking and thereby securing the network.

With more stakers, the network becomes increasingly stable and decentralized. It
becomes *stable*, because assets are locked up instead of causing price changes
through trading. And it becomes *decentralized,* because the power to vote for
governance proposals is distributed amongst more people.

## Evmos Token Model

The Evmos Token Model outlines how the Evmos network is secured through a
balanced incentivized interest from users, developers and validators. In this
model, inflation plays a major role in sustaining this balance. With an initial
supply of 200 million and over 300 million tokens being issued through inflation
during the first year, the model suggests an exponential decline in inflation to
issue 1 billion Evmos tokens within the first 4 years.

We implement two different inflation mechanisms to support the token model:

1. linear inflation for team vesting and
2. exponential inflation for staking rewards, usage incentives and community
   pool.

### Linear Inflation - Team Vesting

The Team Vesting distribution in the Token Model is implemented in a way that
minimized the amount of taxable events. An initial supply of 200M allocated to
`vesting accounts` at genesis. This amount is equal to the total inflation
allocated for team vesting after 4 years (`20% * 1B = 200M`). Over time,
`unvested` tokens on these accounts are converted into `vested` tokens at a
linear rate. Team members cannot delegate, transfer or execute Ethereum
transaction with `unvested` tokens until they are unlocked represented as
`vested` tokens.

### Exponential Inflation - **The Half Life**

The inflation distribution for staking, usage incentives and community pool is
implemented through an exponential formula, a.k.a. the Half Life.

Inflation is minted in daily epochs. During a period of 365 epochs (one year), a
daily provision (`epochProvison`) of Evmos tokens is minted and allocated to staking rewards,
usage incentives and the community pool. The epoch provision depends on module parameters and is recalculated at the end of every epoch.

The calculation of the epoch provision is done according to the following formula:

```latex
periodProvision = exponentialDecay       *  bondingIncentive
f(x)            = (a * (1 - r) ^ x + c)  *  (1 + maxVariance * (1 - bondedRatio / bondingTarget))


epochProvision = periodProvision / epochsPerPeriod

where (with default values):
x = variable    = year
a = 300,000,000 = initial value
r = 0.5         = decay factor
c = 9,375,000   = long term supply

bondedRatio   = variable  = fraction of the staking tokens which are currently bonded
maxVariance   = 0.0       = the max amount to increase inflation
bondingTarget = 0.66      = our optimal bonded ratio
```

```latex
Example with bondedRatio = bondingTarget:

period  periodProvision  cumulated      epochProvision
f(0)    309 375 000      309 375 000	 847 602
f(1)    159 375 000      468 750 000	 436 643
f(2)     84 375 000      553 125 000	 231 164
f(3)     46 875 000      600 000 000	 128 424
```
