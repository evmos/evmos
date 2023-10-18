local default = import 'default.jsonnet';

default {
  'evmos_9000-1'+: {
    'app-config'+: {
      'app-db-backend': 'rocksdb',
      memiavl: {
        enable: true,
      },
      store: {
        streamers: ['versiondb'],
      },
    },
    config+: {
       'db_backend': 'rocksdb',
    },
  },
}
