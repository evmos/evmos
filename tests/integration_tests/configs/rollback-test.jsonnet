local config = import 'default.jsonnet';

config {
  'evmos_9000-1'+: {
    validators: super.validators[0:1] + [{
      name: 'fullnode',
    }],
  },
}
