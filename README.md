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
- (Optional) Docker socket accessible by the backend — the Docker CLI needs access to the host daemon.

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
| `OIDC_ISSUER`  | `http://localhost:8081/realms/LanParty`     | Keycloak realm issuer URL          |

You can override via `.env` files at `backend/cmd/server/.env` or `backend/.env`, or set them in your shell.

---

## Production

### Prerequisites

- Docker host with the Docker socket exposed to the backend container
- A publicly-accessible Keycloak instance (self-hosted or managed)
- A reverse proxy (e.g. Caddy, nginx, Traefik) for TLS termination

### Run with Docker Compose

A production-oriented `docker-compose.yml` can be extended to include all three services:

```yaml
services:
  keycloak:
    image: quay.io/keycloak/keycloak:latest
    environment:
      KC_BOOTSTRAP_ADMIN_USERNAME: admin
      KC_BOOTSTRAP_ADMIN_PASSWORD: <secure-admin-password>
    ports:
      - "127.0.0.1:8081:8080"
    volumes:
      - keycloak_data:/opt/keycloak/data
      - ./import:/opt/keycloak/data/import
    command: start --import-realm

  backend:
    build: ./backend/cmd/server
    ports:
      - "127.0.0.1:8080:8080"
    environment:
      OIDC_ISSUER: https://<your-domain>/realms/LanParty
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./templates:/app/templates
    depends_on:
      keycloak:
        condition: service_healthy

  dashboard:
    build: ./dashboard
    ports:
      - "127.0.0.1:5173:5173"
    environment:
      VITE_OIDC_AUTHORITY: https://<your-domain>/realms/LanParty
    depends_on:
      - backend
```

### Build & Deploy

**Backend image:**

```bash
cd backend/cmd/server
docker build -t lanforge-backend .
```

**Dashboard image:**

```bash
cd dashboard
docker build -t lanforge-dashboard .
```

### Production Checklist

- [ ] Use a **strong admin password** for Keycloak and disable the `start-dev` flag
- [ ] Set **`KC_HOSTNAME`** to your Keycloak domain (e.g. `auth.yourlanparty.com`)
- [ ] Enable **TLS** on all public endpoints via a reverse proxy (Caddy, nginx, Traefik)
- [ ] Restrict Docker socket access — run the backend container **only** on the Docker host, not in a swarm
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
├── backend/
│   ├── cmd/server/               # Go entrypoint
│   ├── internal/
│   │   ├── api/                  # HTTP handlers & router
│   │   ├── auth/                 # OIDC middleware & RBAC
│   │   ├── docker/               # Docker client wrapper
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