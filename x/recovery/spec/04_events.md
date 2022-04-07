# Events

The `x/recovery` module emits the following event:

### Recovery
| Type       | Attribute Key    | Attribute Value     |
| ---------- | :---------------:| -----------------:  |
| `recovery` | `sender`         | `senderBech32`      |
| `recovery` | `receiver`       | `recipientBech32`   |

| `recovery` | `amount`       | `amtStr`   |
| `recovery` | `SrcChannel`       | `packet.SourceChannel`   |
| `recovery` | `srcPort`       | `packet.SourcePort`   |
| `recovery` | `DstPort`       | `packet.DestinationPort`   |
| `recovery` | `DstChannel`       | `packet.DestinationChannel`   |

