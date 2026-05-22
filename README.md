# lanforge

Containerized LAN Party Manager — deploy and manage game servers on a shared Docker host with role-based access control and a web dashboard.

## Architecture

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Keycloak    │◄────│   Backend    │◄────│  Dashboard   │
│  (Auth)      │     │  (Go :8080)  │     │  (Vue :5173) │
└──────────────┘     └──────┬───────┘     └──────────────┘
                            │
                     ┌──────┴───────┐
                     │  Docker Host │
                     │ (game svrs)  │
                     └──────────────┘
```

- **Backend** — Go HTTP server on `:8080`; talks to the Docker daemon to manage containers, volumes, and deployments.
- **Dashboard** — Vue 3 / Vite SPA on `:5173`; OIDC login via Keycloak, real-time logs and stats over WebSocket.
- **Keycloak** — OIDC identity provider on `:8081`; pre-configured `LanParty` realm with role-based access.

---

## Development

### Prerequisites

- [Go](https://go.dev/dl/) 1.26+
- [Node.js](https://nodejs.org/) (includes npm)
- [Docker](https://docs.docker.com/engine/install/) + Docker Compose

### ⚠️ Docker Socket Security Notice
**This project mounts `/var/run/docker.sock` to the backend container for Docker-in-Docker functionality.**

**What this means:**
- The backend has full control over your host's Docker daemon
- Can create, start, stop, delete containers with any permissions your host Docker allows
- **Development OK:** Expected behavior for LAN party management features (deploying game servers)
- **Production Risk:** If the API is compromised, attacker gains Docker-on-host privileges

**Recommended for production deployments:**
- Option 1: Run backend only on the docker host (not in Docker itself) ✓ Easiest
- Option 2: Use read-only volumes where possible + network restrictions
- Option 3: Implement a Docker proxy service that validates commands before forwarding to daemon

See `docker-compose.yml` line 44 and `plan.md` section 1 for details.

### 1. Start Keycloak

```bash
docker compose up -d
```

This starts Keycloak on `http://localhost/auth` with the `LanParty` realm and a `lan-control-plane` client pre-imported from `LanParty-realm.json`.

| Credential          | Value         |
|---------------------|---------------|
| Admin username      | `admin`       |
| Admin password      | `admin`       |

> **Note:** The realm export includes users, roles (`admin`, `operator`), and the OIDC client. You can log into the Keycloak Admin Console at `http://localhost:8081/admin` to inspect or modify them.

### 2. Start the Backend

The backend is written in Go, located in `backend/cmd/server/main.go`. It requires Go 1.26+.

**Build the binary:**

```bash
cd backend/cmd/server
go build -o lanforge-backend .
./lanforge-backend
```

**Run directly (no build step):**

```bash
cd backend/cmd/server
go run main.go
```

**Hot reload during development (recommended):**

Install [air](https://github.com/air-verse/air) for automatic rebuilds on file changes:

```bash
go install github.com/air-verse/air@latest
cd backend
air
```

> If you don't have a `.air.toml` at the backend root, air will use defaults which watch `.go` files in `./cmd/server` and rebuild on save.

**Run tests:**

```bash
cd backend
go test ./...
```

The backend starts on `http://localhost:8080`.  
It reads `OIDC_ISSUER` from `.env` (searched in `backend/cmd/server/.env` then `backend/.env`). If neither file exists, it uses the system environment. If the variable is not set at all, the server will fail to start — see [Environment Variables](#environment-variables).

Available API endpoints (all require a valid access token from Keycloak):

| Method | Path                          | Roles               |
|--------|-------------------------------|---------------------|
| GET    | `/containers`                 | authenticated      |
| POST   | `/containers/{id}/start`      | admin, operator    |
| POST   | `/containers/{id}/stop`       | admin, operator    |
| DELETE | `/containers/{id}`            | admin              |
| GET    | `/containers/{id}/logs`       | authenticated (WS) |
| GET    | `/containers/{id}/stats`      | authenticated (WS) |
| GET    | `/templates`                  | authenticated      |
| GET    | `/templates/{id}`             | authenticated      |
| POST   | `/templates/{id}/deploy`      | admin, operator    |
| GET    | `/deploy/stream/{id}`         | authenticated (WS) |
| GET    | `/volumes`                    | authenticated      |
| POST   | `/volumes/{name}`             | admin, operator    |
| DELETE | `/volumes/{name}`             | admin, operator    |

> **Note:** The backend requires access to the Docker daemon socket (`/var/run/docker.sock` on Linux, or the equivalent on macOS via Docker Desktop / Colima / Orbstack).

#### Development Mode Configuration
The application supports a development/production mode toggle:

```bash
# Set environment variables before starting
export NODE_ENV=development  # or production
export DOCKER_ENABLED=true   # Allow Docker operations (default in dev)

# Or add to backend/.env
NODE_ENV=development
DOCKER_ENABLED=true
```

**Behavior:**
| Environment | Docker Enabled | What Happens |
|-------------|---------------|--------------|
| `NODE_ENV=development` with `DOCKER_ENABLED=true` | ✅ Yes | Full Docker access, all features work |
| `NODE_ENV=development` without flag | ✅ Yes (defaults) | Full Docker access |
| `NODE_DOMAIN=production` with `DOCKER_ENABLED=false` | ❌ Disabled | Container creation commands will fail gracefully, logs only streamed |

**For production deployments:** Set `DOCKER_ENABLED=false` and document that container management features are limited. The API remains functional for viewing stats/logs but cannot create/modify containers without the explicit Docker proxy setup.

### 3. Start the Dashboard

```bash
cd dashboard
npm install
npm run dev
```

The dashboard starts on `http://localhost:5173`.  
It automatically redirects to Keycloak login on first visit. Use a pre-configured user or create one in the Keycloak admin console.

After successful login you will be redirected back to the dashboard where you can:

- View running containers
- Browse deployment templates
- Deploy game servers from templates
- View real-time container logs and resource stats
- Manage Docker volumes

### Environment Variables

| Variable       | Default                                     | Description                        |
|----------------|---------------------------------------------|------------------------------------|
| `NODE_ENV`     | (unset, inferred)                           | Runtime mode (`development` or `production`) |
| `DOCKER_ENABLED` | `true`                                    | Enable Docker socket operations    |
| `OIDC_ISSUER`  | `http://localhost:8081/realms/LanParty`     | Keycloak realm issuer URL          |

**Development:** `NODE_ENV=development, DOCKER_ENABLED=true` (default)  
**Production:** `NODE_ENV=production, DOCKER_ENABLED=false` (recommended) + use Docker proxy in compose

You can override via `.env` files at `backend/cmd/server/.env` or `backend/.env`, or set them in your shell.

---

## Production

### Prerequisites

- Docker host with the Docker socket exposed to the backend container
- A publicly-accessible Keycloak instance (self-hosted or managed)
- A reverse proxy (e.g. Caddy, nginx, Traefik) for TLS termination

### Run with Docker Compose (Recommended)

This setup includes:
- **Caddy** — Reverse proxy on ports 80/443, terminates TLS
- **API Backend** — Go backend at `http://api.home.arpa` 
- **Dashboard** — Vue frontend at `https://chat.home.arpa`
- **Keycloak** — OIDC auth at `http://auth.home.arpa/admin`

```yaml
services:
  # Reverse Proxy (Caddy) — Terminates TLS on ports 80/443
  caddy:
    image: caddy:2.11.3
    container_name: caddy
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./caddy/Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy_data:/data
      - caddy_config:/config
    networks:
      lanforge-net:
        aliases:
          - home.arpa
          - auth.home.arpa
          - chat.home.arpa
    depends_on:
      - keycloak
      - api
      - dashboard
    restart: unless-stopped

  # API Backend (Go) — Internal service, no public port
  api:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: lanforge-api
    environment:
      - DATABASE_URL=file:LanParty-realm.json
      - JWT_SECRET=change-in-production
      - OIDC_ISSUER=https://auth.home.arpa/realms/LanParty
      # NODE_ENV and DOCKER_ENABLED control production behavior (see env section below)
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock  # ⚠️ Security risk in prod
      - caddy_data:/caddy_data:ro  # Read Caddy-ca certs
    networks:
      - lanforge-net
    depends_on:
      - keycloak
    restart: unless-stopped

  # Dashboard (Vue) — Access via Caddy at chat.home.arpa
  dashboard:
    build:
      context: ./dashboard
      dockerfile: Dockerfile
    container_name: lanforge-dashboard
    environment:
      - API_URL=http://api:8080
      - WS_URL=home.arpa
      - OIDC_ISSUER=https://auth.home.arpa/realms/LanParty
    volumes:
      - caddy_data:/caddy_data:ro  # Read Caddy-ca certs
    networks:
      - lanforge-net
    depends_on:
      - api
    restart: unless-stopped

  # Keycloak (OIDC) — Access at auth.home.arpa/admin
  keycloak:
    image: keycloak/keycloak:26.6.2
    container_name: keycloak
    environment:
      - KC_BOOTSTRAP_ADMIN_USERNAME=admin
      - KC_BOOTSTRAP_ADMIN_PASSWORD=<secure-password>
      - KEYCLOAK_IMPORT=/opt/keycloak/data/import/*
      - KC_HEALTH_ENABLED=true
      - KC_HTTP_ENABLED=true
      - KC_PROXY=edge
    volumes:
      - ./keycloak_data:/opt/keycloak/data
      - ./import:/opt/keycloak/data/import
    ports:
      - "8080:8080"
    networks:
      - lanforge-net
    command: start --import-realm

networks:
  lanforge-net:
    driver: bridge

volumes:
  caddy_data:
  caddy_config:
```

### Environment Variables

Set these before running or add to `backend/.env`:

| Variable | Value (Development) | Value (Production) | Description |
|----------|---------------------|-------------------|-------------|
| `NODE_ENV` | development | production | Mode flag (default: unsets) |
| `DOCKER_ENABLED` | true | false OR use Docker proxy | Enable Docker socket operations |
| `OIDC_ISSUER` | http://localhost:8081/realms/LanParty | https://<domain>/realms/LanParty | Keycloak issuer URL |

**Production Command Example:**
```bash
export NODE_ENV=production
export DOCKER_ENABLED=false  # Disable Docker operations without proxy
docker compose up -d
```

### Production Checklist

- [ ] Use a **strong admin password** for Keycloak and disable the `start-dev` flag
- [ ] Set **`KC_HOSTNAME`** to your Keycloak domain (e.g. `auth.yourlanparty.com`)
- [ ] Enable **TLS** on all public endpoints via a reverse proxy (Caddy, nginx, Traefik)
- [ ] Configure Docker socket access: Set `DOCKER_ENABLED=false` in production OR use read-write-only volumes for templates
- [ ] Run backend **directly on host** (not nested Docker), or implement option C environment variable controls: `NODE_ENV=production` + `DOCKER_ENABLED=false`
- [ ] Configure **session timeouts** in the Keycloak realm to match your security posture
- [ ] Set up **Keycloak backup & restore** for user and realm data
- [ ] Limit **OIDC client scopes** — the dashboard only needs `openid`, `profile`, `email`, and `offline_access`
- [ ] Run behind a **WAF or rate limiter** to protect the login endpoint
- [ ] Use Docker **read-only root filesystem** for the backend container where possible
- [ ] Set up **health checks** on all services (backend `/health`, Keycloak `/health/ready`)

### Performance Considerations

- The backend uses WebSocket streaming for logs and stats — ensure your reverse proxy supports WebSocket upgrades
- For large LAN parties (50+ concurrent users), consider increasing Keycloak's JVM heap via `KC_JVM_OPTIONS=-Xms512m -Xmx2g`
- Template YAML files define CPU/memory limits for game servers — adjust per game type

---

## Project Structure

```
lanforge/
├── docker-compose.yml            # Keycloak only (dev)
├── LanParty-realm.json           # Pre-configured Keycloak realm
├──   backend/
│   ├── cmd/server/               # Go entrypoint
│   ├── internal/
│   │   ├── api/                  # HTTP handlers & router
│   │   ├── auth/                 # OIDC middleware & RBAC
│   │   ├── docker/               # Docker client wrapper with security checks
│   │   ├── service/              # Business logic
│   │   ├── templates/            # YAML template store
│   │   └── websocket/            # Log, stats, deploy streams
│   └── templates/                # Game server definitions (YAML)
├── dashboard/
│   ├── src/
│   │   ├── api/                  # Axios HTTP client
│   │   ├── auth/                 # OIDC client (oidc-client-ts)
│   │   ├── components/           # Vue components
│   │   ├── layouts/              # Page layouts
│   │   ├── pages/                # Route pages
│   │   └── stores/               # Pinia state stores
│   └── vite.config.ts
└── LICENSE