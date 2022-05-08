# Cluster

Maybe you need a lot of evmos.  Fire yourself up a large server, the more iops the better.  Adjust replicas for taste, and note that you'll want to restart in a rolling way that leaves 10m between restarts, about every 24 hours for ideal performance.  This will not give you archive data.

```yaml
version: "3.9"
services:
  evmos:
    image: ghcr.io/faddat/evmos
    deploy:
      replicas: 5
      endpoint_mode: vip
      mode: replicated
    labels:
      proxied: "proxied=true"
    networks:
      - overlay
    ports:
      - "26656"
      - "1317"
      - "26657"
      - "8545"
      - "9090"

networks:
  overlay:
```


I have this config-- without a swarm -- running on a single hetzner ax-101.


Now we've got to do some load balancing.  Since we have grpc in our lives, we're going to use gobetween.io, a very simple layer 4 load balancer that can interface with Docker directly.

config for gobetween:

```toml
#
# gobetween.toml - sample config file
#
# Website: http://gobetween.io
# Documentation: https://github.com/yyyar/gobetween/wiki/Configuration
#


#
# Logging configuration
#
[logging]
level = "info"    # "debug" | "info" | "warn" | "error"
output = "stdout" # "stdout" | "stderr" | "/path/to/gobetween.log"
format = "text"   # (optional) "text" | "json"

#
# Pprof profiler configuration
#
[profiler]
enabled = false # false | true
bind = ":6060"  # "host:port"

#
# REST API server configuration
#
[api]
enabled = false  # true | false
bind = ":8888"  # "host:port"
cors = false    # cross-origin resource sharing

#  [api.basic_auth]   # (optional) Enable HTTP Basic Auth
#  login = "admin"    # HTTP Auth Login
#  password = "1111"  # HTTP Auth Password

#  [api.tls]                        # (optional) Enable HTTPS
#  cert_path = "/path/to/cert.pem"  # Path to certificate
#  key_path = "/path/to/key.pem"    # Path to key

#
# Metrics server configuration
#
[metrics]
enabled = true # false | true
bind = ":9284"  # "host:port"

#
# Default values for server configuration, may be overridden in [servers] sections.
# All "duration" fields (for example, postfixed with '_timeout') have the following format:
# <int><duration> where duration can be one of 'ms', 's', 'm', 'h'.
# Examples: "5s", "1m", "500ms", etc. "0" value means no limit
#
[defaults]
max_connections = 0              # Maximum simultaneous connections to the server
client_idle_timeout = "0"        # Client inactivity duration before forced connection drop
backend_idle_timeout = "0"       # Backend inactivity duration before forced connection drop
backend_connection_timeout = "0" # Backend connection timeout (ignored in udp)

#
## Acme (letsencrypt) configuration.
## Letsencrypt allows server obtain free TLS certificates automagically.
## See https://letsencrypt.org for details.
##
## Each server that requires acme certificates should have acme_hosts configured in tls section.
#
#[acme]                           # (optional)
#challenge = "http"               # (optional) http | sni | dns
#http_bind = "0.0.0.0:80"         # (optional) It is possible to bind to other port, but letsencrypt will send requests to http(80) anyway
#cache_dir = "/tmp"               # (optional) directory to put acme certificates

#
# Servers contains as many [server.<name>] sections as needed.
#
[servers]

 [servers.rpc]
  protocol = "tcp"
  bind = "0.0.0.0:26657"
  [servers.rpc.discovery]
  kind = "docker"
  docker_endpoint = "http://localhost:2375" # (required) Endpoint to docker API
  docker_container_label = "proxied=true"   # (optional) Label to filter containers
  docker_container_private_port = 26657        # (required) Private port of container to use


 [servers.api]
  bind = "0.0.0.0:1317"
  protocol = "tcp"
  [servers.api.discovery]
  kind = "docker"
  bind = "0.0.0.0:1317"
  docker_endpoint = "http://localhost:2375" # (required) Endpoint to docker API
  docker_container_label = "proxied=true"   # (optional) Label to filter containers
  docker_container_private_port = 1317        # (required) Private port of container to use


  [servers.grpc]
  protocol = "tcp"
  bind = "0.0.0.0:9090"
  [servers.grpc.discovery]
  kind = "docker"
  docker_endpoint = "http://localhost:2375" # (required) Endpoint to docker API
  docker_container_label = "proxied=true"   # (optional) Label to filter containers
  docker_container_private_port = 9090        # (required) Private port of container to use

  [servers.eth]
  protocol = "tcp"
  bind = "0.0.0.0:8545"
  [servers.eth.discovery]
  kind = "docker"
  docker_endpoint = "http://localhost:2375" # (required) Endpoint to docker API
  docker_container_label = "proxied=true"   # (optional) Label to filter containers
  docker_container_private_port = 8545        # (required) Private port of container to use
```
