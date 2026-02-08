# Metabase Flexibee Adapter

ETL daemon that syncs data from [ABRA Flexibee](https://www.flexibee.eu/) (Czech ERP/accounting system) into PostgreSQL, enabling analytics via [Metabase](https://www.metabase.com/).

```
Flexibee REST API  →  Go Sync Daemon  →  PostgreSQL  →  Metabase
```

## Quick Start

```bash
docker compose up
```

This starts PostgreSQL, the adapter (syncing from Flexibee demo), and Metabase at [http://localhost:3000](http://localhost:3000).

## Docker Compose

The included `docker-compose.yml` runs the full stack:

| Service | Description | Port |
|---|---|---|
| `postgres` | PostgreSQL 17 database | `5432` |
| `adapter` | Flexibee sync daemon | - |
| `metabase` | Metabase analytics UI | `3000` |

### Production usage

Copy `docker-compose.yml` and override the environment variables for your Flexibee instance:

```bash
FLEXIBEE_URL=https://your-flexibee.example.com \
FLEXIBEE_COMPANY=your_company \
FLEXIBEE_USERNAME=your_user \
FLEXIBEE_PASSWORD=your_password \
docker compose up -d
```

### Useful commands

```bash
# Start all services in background
docker compose up -d

# View adapter logs
docker compose logs -f adapter

# Restart only the adapter (e.g. after config change)
docker compose restart adapter

# Stop everything
docker compose down

# Stop and remove data (fresh start)
docker compose down -v
```

## Configuration

All settings can be configured via environment variables or CLI flags. Flags take priority over env vars.

| Env Variable | Flag | Default | Description |
|---|---|---|---|
| `FLEXIBEE_URL` | `--flexibee-url` | *required* | Flexibee base URL |
| `FLEXIBEE_COMPANY` | `--flexibee-company` | *required* | Flexibee company code |
| `FLEXIBEE_USERNAME` | `--flexibee-username` | *required* | Flexibee username |
| `FLEXIBEE_PASSWORD` | `--flexibee-password` | *required* | Flexibee password |
| `DATABASE_URL` | `--database-url` | *required* | PostgreSQL connection URL |
| `SYNC_INTERVAL` | `--sync-interval` | `5m` | How often to sync |
| `SYNC_BATCH_SIZE` | `--sync-batch-size` | `100` | Records per API page |
| `SYNC_CONCURRENCY` | `--sync-concurrency` | `4` | Max parallel evidence syncs |
| `RETENTION_DAYS` | `--retention-days` | `365` | Data retention (0 = keep forever) |
| `CLEANUP_INTERVAL` | `--cleanup-interval` | `24h` | How often to run cleanup |
| `CLEANUP_BATCH_SIZE` | `--cleanup-batch-size` | `1000` | Delete batch size |
| `LOG_LEVEL` | `--log-level` | `info` | debug, info, warn, error |
| `LOG_FORMAT` | `--log-format` | `json` | json or text |

## Synced Evidence Types

The adapter syncs ~30 Flexibee evidence types into `flexibee_*` PostgreSQL tables:

**Sales & Invoicing:** prodejka, faktura-vydana, faktura-prijata, pohledavka, zavazek

**Orders:** objednavka-prijata, objednavka-vydana, nabidka-vydana, nabidka-prijata, poptavka-vydana, poptavka-prijata

**Inventory:** sklad, skladovy-pohyb, skladova-karta

**Contacts:** adresar, kontakt

**Cash & Banking:** banka, pokladni-pohyb, bankovni-ucet, pokladna

**Products:** cenik, skupina-zbozi, merna-jednotka

**Accounting:** stredisko, zakazka, cinnost, ucet, sazba-dph, kurz

**Contracts:** smlouva, dodavatelska-smlouva

**Assets:** majetek

## How It Works

1. On startup, the adapter fetches property definitions from Flexibee and creates/updates PostgreSQL tables dynamically.
2. Every sync interval, it fetches records modified since the last sync (incremental, using `lastUpdate`).
3. Records are upserted into PostgreSQL with full JSON stored in `raw_data` and typed columns for each property.
4. A cleanup job runs daily to remove transactional records older than the retention period. Master data (reference tables) is never cleaned up.

## Development

```bash
# Run tests
go test -v -race -count=1 ./...

# Build
go build -o adapter ./cmd/adapter

# Lint
golangci-lint run
```

## License

MIT
