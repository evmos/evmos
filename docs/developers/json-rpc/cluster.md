# Cluster

Maybe you need a lot of evmos.  Fire yourself up a large server, the more iops the better.  Adjust replicas for taste, and note that you'll want to restart in a rolling way that leaves 10m between restarts, about every 24 hours for ideal performance.  This will not give you archive data.

```yaml
version: "3.9"
services:
  evmos:
    image: ghcr.io/faddat/evmos
    deploy:
      mode: replicated
      replicas: 5
      endpoint_mode: vip
    networks:
      - overlay
    ports:
      - "26656:26656"
      - "1317:1317"
      - "26657:26657"
      - "8545:8545"
      - "9090:9090"

networks:
  overlay:
```
