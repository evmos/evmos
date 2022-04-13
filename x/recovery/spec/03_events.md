<!--
order: 3
-->

# Events

The `x/recovery` module emits the following event:

## Recovery

| Type       |    Attribute Key     |             Attribute Value |
| :--------- | :------------------- | :-------------------------- |
| `recovery` |       `sender`       |              `senderBech32` |
| `recovery` |      `receiver`      |           `recipientBech32` |
| `recovery` |       `amount`       |                    `amtStr` |
| `recovery` | `packet_src_channel` |      `packet.SourceChannel` |
| `recovery` |  `packet_src_port`   |         `packet.SourcePort` |
| `recovery` | `packet_dst_channel` |    `packet.DestinationPort` |
| `recovery` |  `packet_dst_port`   | `packet.DestinationChannel` |
