# PocketID → Listmonk Sync Sidecar

![GitHub License](https://img.shields.io/github/license/lbenicio/pocketid-sidecar-listmonk?style=flat&color=blue)
![GitHub Release](https://img.shields.io/github/v/release/lbenicio/pocketid-sidecar-listmonk?style=flat&color=blue)
[![Dependabot Updates](https://github.com/lbenicio/pocketid-sidecar-listmonk/actions/workflows/dependabot/dependabot-updates/badge.svg)](https://github.com/lbenicio/pocketid-sidecar-listmonk/actions/workflows/dependabot/dependabot-updates)
[![CodeQL](https://github.com/lbenicio/pocketid-sidecar-listmonk/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/lbenicio/pocketid-sidecar-listmonk/actions/workflows/github-code-scanning/codeql)
[![Dependency Graph](https://github.com/lbenicio/pocketid-sidecar-listmonk/actions/workflows/dependabot/update-graph/badge.svg)](https://github.com/lbenicio/pocketid-sidecar-listmonk/actions/workflows/dependabot/update-graph)

A tiny Go service that reconciles **[PocketID](https://github.com/pocket-id/pocket-id)** users with a specific mailing list in **[Listmonk](https://listmonk.app/)**. Run it as a cron job to keep your newsletter list automatically in sync with your identity provider.

## How it works

```text
┌──────────────┐     ┌──────────────────┐     ┌───────────┐
│   PocketID   │────▶│  sync sidecar    │────▶│  Listmonk │
│   (users)    │     │  (reconciliation)│     │  (list)   │
└──────────────┘     └──────────────────┘     └───────────┘
```

Each run performs a full three-way reconciliation:

| Event                        | Action                                                |
| ---------------------------- | ----------------------------------------------------- |
| New PocketID user            | Subscriber **created** in the target list             |
| User **deleted** in PocketID | Subscriber **deleted** from the list                  |
| User name/email changed      | Subscriber **updated** in the list                    |

Matching between PocketID and Listmonk is done via a `pocketid_id` attribute stored on each subscriber record.

## Configuration

All configuration lives in environment variables. Copy `.env.example` to `.env` and fill in your values.

```env
# PocketID
POCKETID_BASE_URL=http://localhost:8090
POCKETID_API_KEY=your-pocketid-admin-api-key

# Listmonk
LISTMONK_BASE_URL=http://localhost:9000
LISTMONK_USERNAME=admin
LISTMONK_PASSWORD=your-listmonk-password
LISTMONK_LIST_ID=1

# Optional
SYNC_DRY_RUN=false    # Set to true to preview changes without mutating
```

## Usage

### Local

```bash
# Build
make build
```

```bash
# Run (reads env vars from your shell)
source .env && ./bin/pocketid-sidecar-listmonk
```

```bash
# Dry run to preview changes
SYNC_DRY_RUN=true go run ./cmd/sync
```

### Docker

```bash
docker build -t pocketid-sidecar-listmonk .
```

```bash
docker run --rm --env-file .env pocketid-sidecar-listmonk
```

### Kubernetes CronJob

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: pocketid-listmonk-sync
spec:
  schedule: "*/5 * * * *"        # every 5 minutes
  jobTemplate:
    spec:
      template:
        spec:
          restartPolicy: Never
          containers:
            - name: sync
              image: pocketid-sidecar-listmonk:latest
              env:
                - name: POCKETID_BASE_URL
                  value: "http://pocketid:8090"
                - name: POCKETID_API_KEY
                  valueFrom:
                    secretKeyRef:
                      name: pocketid-credentials
                      key: api-key
                - name: LISTMONK_BASE_URL
                  value: "http://listmonk:9000"
                - name: LISTMONK_USERNAME
                  value: "admin"
                - name: LISTMONK_PASSWORD
                  valueFrom:
                    secretKeyRef:
                      name: listmonk-credentials
                      key: password
                - name: LISTMONK_LIST_ID
                  value: "1"
```

### Docker Compose (with ofelia cron)

```yaml
services:
  sync:
    build: .
    restart: "no"
    env_file: .env
    labels:
      ofelia.enabled: "true"
      ofelia.job-run.sync.schedule: "@every 5m"

  ofelia:
    image: mcuadros/ofelia:latest
    restart: unless-stopped
    depends_on:
      - sync
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
```

### Plain cron (Linux)

```text
# /etc/cron.d/pocketid-sync — every 5 minutes
*/5 * * * * root . /etc/pocketid-sync.env && /usr/local/bin/pocketid-sidecar-listmonk >> /var/log/pocketid-sync.log 2>&1
```

## Dry run

Set `SYNC_DRY_RUN=true` to see what the syncer **would** do without actually mutating any data in Listmonk. Useful for testing and debugging.

```log
[sync] fetching PocketID users...
[sync] found 3 users in PocketID
[sync] fetching Listmonk subscribers for list...
[sync] found 2 subscribers in Listmonk list 1
[dry-run] would CREATE  subscriber for pocketid=abc123 email=jane@example.com name=Jane Doe
[dry-run] would DELETE  subscriber id=7 (pocketid=olduser) email=old@example.com name=Old User
--- sync complete ---
created: 0
updated: 0
deleted: 0
errors:  0
```

## Building

```bash
make build    # builds to bin/pocketid-sidecar-listmonk
```

```bash
make lint     # runs go vet
```

```bash
make clean    # removes bin/
```
