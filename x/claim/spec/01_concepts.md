<!--
order: 1
-->

# Concepts

Osmosis' initial supply was given as a quadratic fairdrop to ATOM users.
Details of the fairdrop are in the 'fairdrop' section of the following blog post: 
https://medium.com/osmosis/osmo-token-distribution-ae27ea2bb4db

To briefly summarize, a quadratic fairdrop is an airdrop that rewards you an amount proportional to the square root of the amount of tokens you had at the snapshot date.
The particular quadratic fairdrop Osmosis used had a bonus going towards stakers, as detailed in the blog post.

In Osmosis, users are required to claim their airdrop by participating in core network activities. An Airdrop recipient is given 20% of the airdrop amount up front, and then they have to perform the following activities to get the rest:

* 20% is claimed by staking
* 20% is claimed by voting in governance
* 20% is claimed by swapping into a pool
* 20% is claimed by providing liquidity to a pool

At initial, module stores all airdrop users with amounts from genesis inside KVStore.

Furthermore, to incentivize users to claim in a timely manner, the amount of claimable airdrop reduces over time. Users can claim the full airdrop amount for two months (`DurationUntilDecay`).
After two months, the claimable amount linearly decays until 6 months after launch. (At which point none of it is claimable) This is controlled by the parameter `DurationOfDecay` in the code, which is set to 4 months. (6 months - 2 months).

After 6 months from launch, all unclaimed airdrop tokens are sent to the community pool.
