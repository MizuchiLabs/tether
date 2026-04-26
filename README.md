<p align="center">
<img src="./.github/logo.svg" width="80">
<br><br>
<img alt="GitHub Tag" src="https://img.shields.io/github/v/tag/MizuchiLabs/tether?label=Version">
<img alt="GitHub License" src="https://img.shields.io/github/license/MizuchiLabs/tether">
<img alt="GitHub Issues or Pull Requests" src="https://img.shields.io/github/issues/MizuchiLabs/tether">
</p>

# Tether

Tether is a centralized Traefik configuration server and agent designed to simplify dynamic routing for distributed Docker nodes. It removes the need for complex orchestration (like Docker Swarm or Kubernetes) or a full KV store (like Consul/Redis).

## How it Works

Tether consists of two main concepts: the **Central Server** and the **Agent**.

1. **The Agent:** Runs on your individual servers or nodes. It watches Docker labels, generates a Traefik dynamic configuration, and sends it to the central Tether server via a simple HTTP heartbeat.
2. **The Server:** Keeps all the incoming agent configurations in memory. It merges these disparate configs into a single, unified Traefik dynamic configuration.
3. **Traefik:** Your actual Traefik instance(s) point to the central Tether server's HTTP endpoint (`GET /config`). Traefik periodically fetches this merged configuration and updates its routing rules automatically.

### Environments

Tether supports multiple **Environments**. If you have multiple distinct Traefik instances (e.g., a "production" load balancer and a "staging" load balancer, or an "internal" vs. "external" proxy), you can group agents and configurations by environment.

- An agent specifies its environment when sending a heartbeat (e.g., `env: "production"`).
- Traefik fetches the configuration for its specific environment by querying the endpoint with an environment parameter (e.g., `GET /config?env=production`).
- The central server merges and serves configurations isolated to each environment.

### Central Traefik Configuration

Configure your central Traefik instance to pull configuration from the Tether server. You can specify the `env` query parameter to only pull configs for a specific environment.

**traefik.yml:**

```yaml
providers:
  http:
    # Traefik queries the Tether server's endpoint for the specific environment
    endpoint: "http://<TETHER_SERVER_IP>:3000/config?env=production"
    pollInterval: "5s"
    headers:
      Authorization: "Bearer <TETHER_SECRET>"
```

You can also just use the `http://<TETHER_SERVER_IP>:3000/config` endpoint, it will use env=default.

**or using cli arguments:**

```bash
command:
  - --providers.http.endpoint=http://tether:3000/config
  - --providers.http.headers.Authorization=Bearer your-super-secret-key
```

## Setup

### Docker

The Tether server is a simple HTTP server that accepts agent heartbeats and returns a unified Traefik dynamic configuration. You can run it as a Docker container, or the raw binary.

```yaml
services:
  tether:
    image: ghcr.io/mizuchilabs/tether:latest
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
    environment:
      - TETHER_SECRET=your-super-secret-key
      # - TETHER_CONFIG=/path/to/dynamic.yml # Optional: defaults to /data/dynamic.yml
      # - TETHER_PORT=1234 # Optional: default is 3000
    ports:
      - 3000:3000
    restart: unless-stopped
```

### Binary

Download the binary from [releases](https://github.com/mizuchilabs/tether/releases).

## License

Apache 2.0 License - see [LICENSE](LICENSE) for details.
