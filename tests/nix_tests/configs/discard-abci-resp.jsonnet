local config = import 'default.jsonnet';

config {
  'evmos_9000-1'+: {
    config+: {
      storage: {
        discard_abci_responses: true,
      },
    },
  },
}
