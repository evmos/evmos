local config = import 'default.jsonnet';

config {
  'evmos_9000-1'+: {
    key_name: 'signer1',
    'app-config'+: {
      'index-events': super['index-events'] + ['message.action'],
     grpc: {
        'enable': true,
      },
    },
  },
  'stride-1': {
    cmd: 'strided',
    'account-prefix': 'stride',
    'app-config': {
      'minimum-gas-prices': '0ustrd',
    },
    validators: [
      {
        coins: '2234240000000000000ustrd',
        staked: '10000000000000ustrd',
        mnemonic: '${VALIDATOR1_MNEMONIC}',
        base_port: 26800,
      },
      {
        coins: '987870000000000000ustrd',
        staked: '20000000000000ustrd',
        mnemonic: '${VALIDATOR2_MNEMONIC}',
        base_port: 26810,
      },
    ],
    accounts: [
      {
        name: 'community',
        coins: '10000000000000ustrd',
        mnemonic: '${COMMUNITY_MNEMONIC}',
      },
      {
        name: 'relayer',
        coins: '10000000000000ustrd',
        mnemonic: '${SIGNER1_MNEMONIC}',
      },
      {
        name: 'signer2',
        coins: '10000000000000ustrd',
        mnemonic: '${SIGNER2_MNEMONIC}',
      },
    ],
    genesis: {
      app_state: {
        staking: {
          params: {
            unbonding_time: '1814400s',
            bond_denom: 'ustrd',
          },
        },
        gov: {
          voting_params: {
            voting_period: '10s',
          },
          params: {
            max_deposit_period: '10s',
            voting_period: '10s',
            min_deposit: [
              {
                denom: 'ustrd',
                amount: '1',
              },
            ],
          },
        },
        mint: {
          params: {
            mint_denom: 'ustrd',
          },
        },      
        transfer: {
          params: {
            receive_enabled: true,
            send_enabled: true,
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
        id: 'evmos_9000-1',
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
          price: 40000000000,
          denom: 'aevmos',
        },
        extension_options: [{
          type: 'ethermint_dynamic_fee',
          value: '1000000',
        }],
      },
      {
        id: 'stride-1',
        max_gas: 3000000,
        default_gas: 100000,
        address_type: {
          derivation: 'cosmos',
        },
        gas_price: {
          price: 1000000,
          denom: 'ustrd',
        },
      },
    ],
  },
}
