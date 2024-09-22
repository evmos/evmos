local default = import 'default.jsonnet';

default {
  'evmos_9002-1'+: {
    config+: {
      consensus+: {
        timeout_commit: '5s',
      },
    },
  },
}
