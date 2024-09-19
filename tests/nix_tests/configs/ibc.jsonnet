local config = import 'default.jsonnet';

config {
  'evmos_9002-1'+: {
    key_name: 'signer1',
    accounts: super.accounts[:std.length(super.accounts) - 1] + [super.accounts[std.length(super.accounts) - 1] {
      coins: super.coins + ',100000000000ibcfee',
    }],
    'app-config'+: {
      'index-events': super['index-events'] + ['message.action'],
     'grpc'+: {
        'enable': true,
      },
    },
    genesis+: {
      app_state+: {
        feemarket+: {
          params+: {
            no_base_fee: true,
            base_fee: '0',
          },
        },
      },
    },
  },
  'chainmain-1': {
    cmd: 'chain-maind',
    'start-flags': '--trace',
    'account-prefix': 'cro',
    'app-config': {
      'minimum-gas-prices': '500basecro',
    },
    validators: [
      {
        coins: '2234240000000000000basecro',
        staked: '10000000000000basecro',
        mnemonic: '${VALIDATOR1_MNEMONIC}',
        base_port: 26800,
      },
      {
        coins: '987870000000000000basecro',
        staked: '20000000000000basecro',
        mnemonic: '${VALIDATOR2_MNEMONIC}',
        base_port: 26810,
      },
    ],
    accounts: [
      {
        name: 'community',
        coins: '10000000000000basecro',
        mnemonic: '${COMMUNITY_MNEMONIC}',
      },
      {
        name: 'relayer',
        coins: '10000000000000basecro',
        mnemonic: '${SIGNER1_MNEMONIC}',
      },
      {
        name: 'signer2',
        coins: '10000000000000basecro',
        mnemonic: '${SIGNER2_MNEMONIC}',
      },
    ],
    genesis: {
      app_state: {
        staking: {
          params: {
            unbonding_time: '1814400s',
          },
        },
        gov: {
          voting_params: {
            voting_period: '1814400s',
          },
          deposit_params: {
            max_deposit_period: '1814400s',
            min_deposit: [
              {
                denom: 'basecro',
                amount: '10000000',
              },
            ],
          },
        },
        transfer: {
          params: {
            receive_enabled: true,
            send_enabled: true,
          },
        },
        interchainaccounts: {
          host_genesis_state: {
            params: {
              allow_messages: [
                '/cosmos.bank.v1beta1.MsgSend',
              ],
            },
          },
        },
      },
    },
  },
  relayer: {
    mode: {
      clients: {
        enabled: true,
        refresh: true,
        misbehaviour: true,
      },
      connections: {
        enabled: true,
      },
      channels: {
        enabled: true,
      },
      packets: {
        enabled: true,
        clear_interval: 100,
        clear_on_start: true,
        tx_confirmation: true,
      },
    },
    rest: {
      enabled: true,
      host: '127.0.0.1',
      port: 3000,
    },
    chains: [
      {
        id: 'evmos_9002-1',
        max_gas: 3000000,
        default_gas: 100000,
        gas_multiplier: 1.2,
        address_type: {
          derivation: 'ethermint',
          proto_type: {
            pk_type: '/ethermint.crypto.v1.ethsecp256k1.PubKey',
          },
        },
        gas_price: {
          price: 80000000000,
          denom: 'aevmos',
        },
        extension_options: [{
          type: 'ethermint_dynamic_fee',
          value: '1000000',
        }],
      },
      {
        id: 'chainmain-1',
        max_gas: 3000000,
        default_gas: 100000,
        gas_price: {
          price: 1000000,
          denom: 'basecro',
        },
      },
    ],
  },
}
