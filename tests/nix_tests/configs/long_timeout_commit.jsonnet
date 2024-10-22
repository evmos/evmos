local default = import 'default.jsonnet';

default {
  'eidon-chain_9002-1'+: {
    config+: {
      consensus+: {
        timeout_commit: '5s',
      },
    },
  },
}
