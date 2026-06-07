# Entain Test — Wallet Service

An HTTP service that processes balance-changing transactions from 3rd-party
providers and exposes the current user balance. Go + Postgres, runnable with a
single `docker compose up -d`.

## Endpoints

### `POST /user/{userId}/transaction`

Updates a user's balance.

Headers:

- `Source-Type: game` — one of `game`, `server`, `payment` (required)
- `Content-Type: application/json`

Body:

```json
{ "state": "win", "amount": "10.15", "transactionId": "some-generated-id" }
```

- `state` — `win` (increases balance) or `lose` (decreases balance)
- `amount` — string, up to 2 decimal places
- `transactionId` — unique id; the same id is processed **only once**

Responses:

| Situation                                   | Status |
| ------------------------------------------- | ------ |
| success                                     | `200`  |
| bad JSON / state / amount / userId          | `400`  |
| missing or invalid `Source-Type`            | `400`  |
| insufficient funds (`lose` exceeds balance) | `400`  |
| user not found                              | `404`  |
| duplicate `transactionId`                   | `422`  |
| unexpected / database error                 | `500`  |

Non-2xx responses carry a JSON body: `{"error": "..."}`.

### `GET /user/{userId}/balance`

Returns the current balance:

```json
{ "userId": 1, "balance": "9.25" }
```

### `GET /health`

Liveness probe: `200 {"status":"ok"}`.

## Running

Requires Docker with the Compose plugin. From the project root:

```bash
docker compose up -d
```

Compose runs three services in order: Postgres, a one-shot `migrate` job that
applies the schema and seeds predefined users **1, 2, 3** (each balance
`0.00`), and the `app`. The app only starts once migration completes
successfully. The API listens on **http://localhost:8080**.

Stop and remove everything (including the database volume):

```bash
docker compose down -v
```

## Migrations

Migrations live in `cmd/migrate` and are embedded into a standalone binary.
Under Compose the `migrate` service applies them automatically, but they can
also be run manually against the running database:

```bash
# apply (up) — the default direction
docker compose run --rm --entrypoint /migrate migrate -direction up

# roll back (down)
docker compose run --rm --entrypoint /migrate migrate -direction down
```

## Try it

```bash
# Win 10.15 for user 1
curl -i -X POST http://localhost:8080/user/1/transaction \
  -H "Source-Type: game" \
  -H "Content-Type: application/json" \
  -d '{"state":"win","amount":"10.15","transactionId":"tx-1"}'

# Lose 1.15 for user 1
curl -i -X POST http://localhost:8080/user/1/transaction \
  -H "Source-Type: server" \
  -H "Content-Type: application/json" \
  -d '{"state":"lose","amount":"1.15","transactionId":"tx-2"}'

# Replaying tx-1 is rejected as already processed -> 422
curl -i -X POST http://localhost:8080/user/1/transaction \
  -H "Source-Type: game" \
  -H "Content-Type: application/json" \
  -d '{"state":"win","amount":"10.15","transactionId":"tx-1"}'

# Current balance -> {"userId":1,"balance":"9.00"}
curl -i http://localhost:8080/user/1/balance
```

## Tests

### Unit tests (no database)

```bash
go test ./...
```

Covers money parsing/formatting and the service's validation and error
handling. The service tests use a [mockery](https://github.com/vektra/mockery)
mock of the repository (`internal/service/mocks`, regenerate with `mockery`).
Integration tests skip automatically when `TEST_DATABASE_URL` is unset.

### Integration & concurrency tests (real Postgres)

These exercise the `SELECT ... FOR UPDATE` write path against a live database,
including a test that fires hundreds of concurrent win/lose operations at one
user and asserts the final balance is exact and never negative.

With the stack running (`docker compose up -d`), point the tests at the
database (published on host port **5433**):

```bash
TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5433/entaintest?sslmode=disable" \
  go test ./internal/repository/ -run Test_WalletRepositorySuite -v
```

If host port 5433 is taken by another local Postgres, run the tests inside the
compose network instead (unambiguously reaches the container's `db`):

```bash
docker run --rm --network entain-test_default \
  -v "$(pwd):/src" -w /src \
  -e TEST_DATABASE_URL="postgres://postgres:postgres@db:5432/entaintest?sslmode=disable" \
  golang:1.25-alpine sh -c "go test ./internal/repository/ -run Test_WalletRepositorySuite -v"
```

## Design notes

- **Layered architecture.** `controller` (router → controllers → middleware) →
  `service` (business rules) → `repository` (persistence), over shared `model`
  types and `common` (enums, business errors). Each layer depends only on the
  one below; the service depends on a repository interface, which keeps it
  unit-testable with a generated mock.
- **Money as integer cents.** Amounts are parsed from strings into `int64`
  minor units (`"10.15"` → `1015`) and stored as `BIGINT`. This avoids
  floating-point rounding errors entirely; output is formatted back to a
  2-decimal string.
- **Idempotency.** `transactions.transaction_id` has a `UNIQUE` constraint.
  The insert happens inside the DB transaction before any balance check, so a
  repeat fails the constraint and is reported as a duplicate (`422`) — there is
  no check-then-act race.
- **Concurrency & no negative balance.** Each transaction runs in a single DB
  transaction that locks the user row with `SELECT ... FOR UPDATE` *first*, then
  inserts and updates. Locking first gives a consistent lock order (no
  lock-upgrade deadlock) and serializes concurrent requests for the same user,
  while different users proceed in parallel. A `CHECK (balance >= 0)` constraint
  is a final backstop against ever persisting a negative balance.
- **Migrations** are idempotent (`CREATE TABLE IF NOT EXISTS`,
  `INSERT ... ON CONFLICT`) with matching up/down scripts, embedded into a
  dedicated `migrate` binary and applied by a one-shot Compose service.
- **Startup ordering.** Compose gates `migrate` on a Postgres healthcheck and
  gates `app` on `migrate` completing, and the app retries the initial
  connection — so `up -d` works without manual sequencing.

## Configuration

Defaults work out of the box under Compose. Overridable via environment
variables: `HTTP_PORT`, `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASS`, `DB_NAME`,
`DB_SSLMODE`.

## Layout

```
cmd/
  server/              entrypoint: config, DB pool, HTTP server
  migrate/             migration tool + embedded up/down SQL
internal/
  controller/          router, controllers, response helpers
    middleware/        Source-Type validation, recover, logging
  service/             business rules (validation, win/lose, error mapping)
    mocks/             generated mock of the repository interface (mockery)
  repository/          Postgres access (FOR UPDATE tx, named args)
  model/               domain types; request/ and response/ subpackages
  common/
    enum/              State, SourceType
    businesserror/     domain error constants
  money/               string <-> integer-cents conversion
  config/              environment configuration
```
