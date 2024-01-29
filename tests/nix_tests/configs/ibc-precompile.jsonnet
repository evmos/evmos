local config = import 'ibc.jsonnet';

config {
  'evmos_9000-1'+: {
    genesis+: {
      app_state+: {
        feemarket+: {
          params+: {
            no_base_fee: false,
            base_fee: '100000000000',
          },
        },
      },
    },
  },
  'chainmain-1'+: {
    validators: [
      {
        coins: '2234240000000000000cro',
        staked: '10000000000000cro',
        mnemonic: '${VALIDATOR1_MNEMONIC}',
        base_port: 28040,
      },
      {
        coins: '987870000000000000cro',
        staked: '20000000000000cro',
        mnemonic: '${VALIDATOR2_MNEMONIC}',
        base_port: 28060,
      },
    ],
  },
  'relayer'+: {
      'rest'+: {
        port: 3010,
    },
  }
}
