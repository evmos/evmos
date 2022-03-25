#!/bin/bash
import subprocess
import yaml
import json

# Set chain id of Evmos mainnet
chain_id = "evmos_9001-2"

# Set address of node we'll be querying. You can find other nodes at atlas.cosmos.network or https://github.com/cosmos/registry
node = "http://159.138.10.224:26657"

# Output file for parameters in json format
output_file = "parameters.json"

params = {
    'auth': {
        'MaxMemoCharacters': '',
        'SigVerifyCostED25519': '',
        'SigVerifyCostSecp256k1': '',
        'TxSigLimit': '',
        'TxSizeCostPerByte': ''
    },
    'bank': {
        'DefaultSendEnabled': '',
        'SendEnabled': ''
    },
    'crisis': {
        'ConstantFee': ''
    },
    'distribution': {
        'baseproposerreward': '',
        'bonusproposerreward': '',
        'communitytax': '',
        'withdrawaddrenabled': ''
    },
    'gov': {
        'depositparams': '',
        'tallyparams': '',
        'votingparams': ''
    },
    'slashing': {
        'DowntimeJailDuration': '',
        'MinSignedPerWindow': '',
        'SignedBlocksWindow': '',
        'SlashFractionDoubleSign': '',
        'SlashFractionDowntime': ''
    },
    'staking': {
        'BondDenom': '',
        'HistoricalEntries': '',
        'MaxEntries': '',
        'MaxValidators': '',
        'UnbondingTime': ''
    },
    'transfer': {
        'SendEnabled': '',
        'ReceiveEnabled': ''
    },
    'baseapp': {
        'BlockParams': '',
        'EvidenceParams': '',
        'ValidatorParams': ''
    }
}

for subspace, keys in params.items():
    for key, value in keys.items(): 
        query_result = subprocess.check_output(['evmosd query params subspace' + ' ' + str(subspace) + ' ' + str(key) + ' ' + '--node ' + node + ' --chain-id ' + chain_id], shell=True)
        yaml_result = yaml.safe_load(query_result)['value']
        print(yaml_result)
        params[subspace][key] = json.loads(yaml_result)

with open(output_file, 'w') as outfile:
    json.dump(params, outfile, indent=4, sort_keys=True)