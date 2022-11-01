<!--
order: 1
-->

# Concepts

The `epochs` module defines on-chain timers that execute at fixed time intervals. Other Evoblock modules can then register logic to be executed at the timer ticks. We refer to the period in between two timer ticks as an "epoch".

Every timer has a unique identifier, and every epoch will have a start time and an end time, where `end time = start time + timer interval`.
