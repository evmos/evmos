local config = import 'default.jsonnet';

config {
  'eidon-chain_9002-1'+: {
    config+: {
      storage: {
        discard_abci_responses: true,
      },
    },
  },
}
