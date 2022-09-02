#!/bin/bash

IS_CLI_OK=$(pointd status 2>&1 | grep -o NodeInfo)

echo $IS_CLI_OK

if [ "$IS_CLI_OK" != "NodeInfo" ]; then
    echo "Your pointd cli is not working properly. Run pointd manually to test it"
    exit
fi

set -ue

IS_NOT_SYNC=$(pointd status 2>&1  | jq .SyncInfo | grep catching_up | grep -o 'true\|false')

if [ "$IS_NOT_SYNC" = "false" ]; then
    echo "Your node is synced"
else
    echo "Your node is out of sync"
    CURRENT_HEIGHT=$(curl  -s http://xnet-neptune-1.point.space:8545 -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq .result | tr -d '"')
    LOCAL_HEIGHT=$(curl -s http://127.0.0.1:8545 -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq .result | tr -d '"')
    echo Current blockchain height is: $[$CURRENT_HEIGHT]
    echo Your current height is: $[$LOCAL_HEIGHT]
    echo "Wait for node to fully sync"
    exit
fi

VOTING_POWER=$(pointd status | jq .ValidatorInfo.VotingPower | tr -d '"')

if [[ $VOTING_POWER -gt 0 ]]; then
  echo "Your voting power is $VOTING_POWER, it means your validator is working ok"
else
  echo "Your voting power is $VOTING_POWER, it means you are not a validator. Let's verify why"
  echo "You will have to provide the key name you have used to create the validator"
  echo "We will show you the list of all the keys you have created. We need you to pickup the name from the list (probably it will be validatorkey)"
  pointd keys list | grep name
  read -p "Type the name of your key: " KEYNAME
  echo "We are going to check if you are jailed"
  VALOPER_ADDRESS=$(pointd keys show $KEYNAME -a --bech val)
  JAILED=$(pointd query staking validator $VALOPER_ADDRESS | grep jailed | grep -o 'true\|false')
  if [ "$JAILED" = "true" ]; then
    echo "Your node is jailed"
    echo "We will check if you can unjail"
    MIN_SELF_DELEGATION=$(pointd query staking validator $VALOPER_ADDRESS | grep min_self_delegation | grep -Eo '[0-9]+')
    TOKENS=$(pointd query staking validator $VALOPER_ADDRESS | grep tokens | grep -Eo '[0-9]+')
    ENOUGH_TOKENS=$(bc <<< "$TOKENS > $MIN_SELF_DELEGATION")
    if [ "$ENOUGH_TOKENS" -eq 1 ]; then
      JAILED_UNTIL=$(pointd query slashing signing-info $(pointd tendermint show-validator) | grep jailed_until)
      if [[ "$OSTYPE" == "darwin"* ]]; then
        JAILED_TIMESTAMP=$(gdate --date="$(echo $JAILED_UNTIL | cut -c 16-34)"  +%s)
      else
        JAILED_TIMESTAMP=$(date --date="$(echo $JAILED_UNTIL | cut -c 16-34)"  +%s)
      fi
      NOW=$(date -u +"%s")
      if [ "$NOW" -ge "$JAILED_TIMESTAMP" ]; then
        echo "You are able to unjail. Try to unjail manually running unjail command"
      else
        echo "Your unjail time has not expired yet, you need to wait until $JAILED_UNTIL"
      fi

    else
      echo "Your staked tokens are lower than your min_self_delegation param, to unjail you need to delegate more tokens"
      echo "You currently have set min_self_delegation to: $MIN_SELF_DELEGATION but you have staked only $TOKENS tokens. You need to delegate at least $(bc <<< "$MIN_SELF_DELEGATION - $TOKENS") apoint"
      echo "Check in faq document how to delegate more tokens. After delegating re run this script"
      exit
    fi

  else
    echo "You are not jailed, but you have voting power 0, this should not happen contact support"
    exit
  fi
fi

