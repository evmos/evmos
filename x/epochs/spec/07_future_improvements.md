<!--
order: 7
-->

# Future Improvements

## Correct Usage

In the current design, each epoch should be at least two blocks, as the start block should be different from the endblock.
Because of this, the time allocated to each epoch will be `max(block_time x 2, epoch_duration)`.
For example: if the `epoch_duration` is set to `1s`, and `block_time` is `5s`, actual epoch time should be `10s`.

It is recommended to configure `epoch_duration` to be more than two times the `block_time`, to use this module correctly.
If there is a mismatch between the `epoch_duration` and the actual epoch time, as in the example above,
then module logic could become invalid.

## Block-Time Drifts

This implementation of the `x/epochs` module has block-time drifts based on the value of `block_time`.
For example: if we have an epoch of 100 units that ends at `t=100`,
and we have a block at `t=97` and a block at `t=104` and `t=110`, this epoch ends at `t=104`,
and the new epoch will start at `t=110`.

There are time drifts here, varying about 1-2 blocks time, which will slow down epochs.
