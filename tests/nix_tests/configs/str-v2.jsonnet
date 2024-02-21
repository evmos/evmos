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
  'cosmoshub-1': {
      cmd: 'gaiad',
      'account-prefix': 'cosmos',
      'app-config': {
      'minimum-gas-prices': '0.0025uatom',
    },
    validators: [
      {
        coins: '2234240000000000000uatom',
        staked: '10000000000000uatom',
        mnemonic: '${VALIDATOR1_MNEMONIC}',
        base_port: 26900,
      },
      {
        coins: '987870000000000000uatom',
        staked: '20000000000000uatom',
        mnemonic: '${VALIDATOR2_MNEMONIC}',
        base_port: 26910,
      },
    ],
    accounts: [
      {
        name: 'community',
        coins: '10000000000000uatom',
        mnemonic: '${COMMUNITY_MNEMONIC}',
      },
      {
        name: 'relayer',
        coins: '10000000000000uatom',
        mnemonic: '${SIGNER1_MNEMONIC}',
      },
      {
        name: 'signer2',
        coins: '10000000000000uatom',
        mnemonic: '${SIGNER2_MNEMONIC}',
      },
    ],
    genesis: {
      app_state: {
        staking: {
          params: {
            unbonding_time: '1814400s',
            bond_denom: 'uatom',
          },
        },
        crisis: {
          constant_fee: {
            denom: 'uatom'
          }
        },
        txfees: {
          basedenom: 'uatom',
        },
        gov: {
          voting_params: {
            voting_period: '1814400s',
          },
          deposit_params: {
            max_deposit_period: '1814400s',
            min_deposit: [
              {
                denom: 'uatom',
                amount: '10000000',
              },
            ],
          },
        },
        poolincentives: {
          params: {
            minted_denom: 'uatom'
          }
        },
        mint: {
          params: {
            mint_denom: 'uatom',
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
  'cosmoshub-2': {
        cmd: 'gaiad',
        'account-prefix': 'cosmos',
        'app-config': {
        'minimum-gas-prices': '0.0025uatom',
      },
      validators: [
        {
          coins: '2234240000000000000uatom',
          staked: '10000000000000uatom',
          mnemonic: '${VALIDATOR1_MNEMONIC}',
          base_port: 26920,
        },
        {
          coins: '987870000000000000uatom',
          staked: '20000000000000uatom',
          mnemonic: '${VALIDATOR2_MNEMONIC}',
          base_port: 26930,
        },
      ],
      accounts: [
        {
          name: 'community',
          coins: '10000000000000uatom',
          mnemonic: '${COMMUNITY_MNEMONIC}',
        },
        {
          name: 'relayer',
          coins: '10000000000000uatom',
          mnemonic: '${SIGNER1_MNEMONIC}',
        },
        {
          name: 'signer2',
          coins: '10000000000000uatom',
          mnemonic: '${SIGNER2_MNEMONIC}',
        },
      ],
      genesis: {
        app_state: {
          staking: {
            params: {
              unbonding_time: '1814400s',
              bond_denom: 'uatom',
            },
          },
          crisis: {
            constant_fee: {
              denom: 'uatom'
            }
          },
          txfees: {
            basedenom: 'uatom',
          },
          gov: {
            voting_params: {
              voting_period: '1814400s',
            },
            deposit_params: {
              max_deposit_period: '1814400s',
              min_deposit: [
                {
                  denom: 'uatom',
                  amount: '10000000',
                },
              ],
            },
          },
          poolincentives: {
            params: {
              minted_denom: 'uatom'
            }
          },
          mint: {
            params: {
              mint_denom: 'uatom',
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
        id: 'cosmoshub-1',
        max_gas: 3000000,
        default_gas: 1000000,
        gas_multiplier: 1.2,
        address_type: {
          derivation: 'cosmos',
        },
        gas_price: {
          price: 1000000,
          denom: 'uatom',
        },
      },
      {
        id: 'cosmoshub-2',
        max_gas: 3000000,
        default_gas: 1000000,
        gas_multiplier: 1.2,
        address_type: {
          derivation: 'cosmos',
        },
        gas_price: {
          price: 1000000,
          denom: 'uatom',
        },
      },
    ],
  },
}
