<!--
order: 1
-->

# Genesis

## User rewards

All the users' wallets that will have available rewards must be added to the genesis file on the `claims` section:

```sh
# For this example we are going to just set the rewards for the first address in the node
node_address=$(evmosd keys list | grep  "address: " | cut -c12-)
amount_to_claim=10000
cat $HOME/.evmosd/config/genesis.json | jq -r --arg node_address "$node_address" --arg amount_to_claim "$amount_to_claim" '.app_state["claims"]["claims_records"]=[{"initial_claimable_amount":$amount_to_claim, "actions_completed":[false, false, false, false],"address":$node_address}]' > $HOME/.evmosd/config/tmp_genesis.json && mv $HOME/.evmosd/config/tmp_genesis.json $HOME/.evmosd/config/genesis.json
```

_NOTE: `actions_completed` is a boolean array with length 4, that represents if the mission was already completed or not._

## Module Account

In order to use the `Claims` module coins must be allocated in the claims' module account (`0xA61808Fe40fEb8B3433778BBC2ecECCAA47c8c47`, `evmos15cvq3ljql6utxseh0zau9m8ve2j8erz89m5wkz`).

The amount of coins sent to the module account MUST be equal to the total amount that is reserved for all the users' rewards.

- Allocate coins to the module account:

```sh
# Amount_to_claim must be equals to the sum of all the rewards
amount_to_claim=10000
cat $HOME/.evmosd/config/genesis.json | jq -r --arg amount_to_claim "$amount_to_claim" '.app_state["bank"]["balances"] += [{"address":"evmos15cvq3ljql6utxseh0zau9m8ve2j8erz89m5wkz","coins":[{"denom":"aevmos", "amount":$amount_to_claim}]}]' > $HOME/.evmosd/config/tmp_genesis.json && mv $HOME/.evmosd/config/tmp_genesis.json $HOME/.evmosd/config/genesis.json
```

- Update the total supply after adding all the genesis transactions:

```sh
validators_supply=$(cat $HOME/.evmosd/config/genesis.json | jq -r '.app_state["bank"]["supply"][0]["amount"]')
total_supply=$(bc <<< "$amount_to_claim+$validators_supply")
cat $HOME/.evmosd/config/genesis.json | jq -r --arg total_supply "$total_supply" '.app_state["bank"]["supply"][0]["amount"]=$total_supply' > $HOME/.evmosd/config/tmp_genesis.json && mv $HOME/.evmosd/config/tmp_genesis.json $HOME/.evmosd/config/genesis.json
```

## Module params

There are 4 params that can be set to configure the module:

- EnableClaims: this flag activates/deactivates the claims module. (Bool)
- AirdropStartTime: time when the airdrop starts (Timestamp format `%Y-%m-%dT%TZ`)
- DurationUntilDecay: duration until the decay stats (Duration: `1000s`)
- DurationOfDecay: duration until the rewards are no longer claimable after the decay starts (Duration: `1000s`)
- ClaimsDenom: denomination used for the rewards (String `aevmos`)

```sh
current_date=$(date -u +"%Y-%m-%dT%TZ")
cat $HOME/.evmosd/config/genesis.json | jq -r --arg current_date "$current_date" '.app_state["claims"]["params"]["airdrop_start_time"]=$current_date' > $HOME/.evmosd/config/tmp_genesis.json && mv $HOME/.evmosd/config/tmp_genesis.json $HOME/.evmosd/config/genesis.json
```
