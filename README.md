# Entain Test — Wallet Service

An HTTP service that processes balance-changing transactions from 3rd-party
providers and exposes the current user balance. Go + Postgres.

## Running

Requires Docker with the Compose plugin. From the project root:

```bash
docker compose up -d
```

Compose runs three services in order: Postgres, a one-shot `migrate` job that
applies the schema and seeds predefined users **1, 2, 3** (each balance
`0.00`), and the `app` (which starts only once migration succeeds). The API
listens on **http://localhost:8080**.

Stop and remove everything (including the database volume):

```bash
docker compose down -v
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

# Current balance
curl -i http://localhost:8080/user/1/balance
```

## Endpoints

### `POST /user/{userId}/transaction`

Updates a user's balance.

Headers:

- `Source-Type: game` — one of `game`, `server`, `payment` (required)
- `Content-Type: application/json`

Body:

```json
{
  "state": "win",
  "amount": "10.15",
  "transactionId": "some-generated-id"
}
```

- `state` — `win` (increases balance) or `lose` (decreases balance)
- `amount` — string, up to 2 decimal places
- `transactionId` — unique id; the same id is processed **only once**

On success returns `200` with the resulting state:

```json
{
  "transactionId": "tx-1",
  "userId": 1,
  "state": "win",
  "sourceType": "game",
  "amount": "10.15",
  "balance": "10.15"
}
```

### `GET /user/{userId}/balance`

Returns the current balance:

```json
{
  "userId": 1,
  "balance": "9.25"
}
```

### `GET /health`

Liveness probe: `200`.

```json
{
  "status": "ok"
}
```

### Status codes

| Situation                                         | Status |
| ------------------------------------------------- | ------ |
| success                                           | `200`  |
| malformed body or non-positive `userId`           | `400`  |
| business errors (invalid field, unknown user,     | `422`  |
| insufficient funds, duplicate `transactionId`)    |        |
| unexpected / database error                       | `500`  |

Non-2xx responses carry a JSON body:

```json
{
  "error": "..."
}
```

## Migrations

Migrations live in `cmd/migrations/sql` and are embedded into the binary. Under
Compose the `migrate` service applies them automatically; they can also be run
manually against the running database:

```bash
# apply (up) — the default direction
docker compose run --rm migrate migrate -direction up

# roll back (down)
docker compose run --rm migrate migrate -direction down
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

With the stack running (`docker compose up -d`):

```bash
TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5432/entaintest?sslmode=disable" \
  go test ./internal/repository/ -run Test_WalletRepositorySuite -v
```

## Design notes

- **Layered architecture.** `infrastructure` (router, middleware, response
  adapter) → `controller` (HTTP mapping) → `service` (business rules) →
  `repository` (persistence). Each layer has its own request/response models and
  mappers, and depends on the layer below through an interface, which keeps the
  service and controller unit-testable with generated mocks.
- **Money as integer cents.** Amounts are parsed from strings into `int64`
  minor units (`"10.15"` → `1015`) and stored as `BIGINT`, avoiding
  floating-point rounding errors; output is formatted back to a 2-decimal
  string.
- **Idempotency.** `transactions.transaction_id` is `UNIQUE`. The insert runs
  inside the DB transaction, so a repeat fails the constraint and is reported as
  a duplicate — there is no check-then-act race.
- **Concurrency & no negative balance.** Each transaction locks the user row
  with `SELECT ... FOR UPDATE` first, then inserts and updates. Locking first
  gives a consistent lock order (no lock-upgrade deadlock) and serializes
  concurrent requests for the same user, while different users proceed in
  parallel. A `lose` larger than the balance is rejected, and a
  `CHECK (balance >= 0)` constraint is a final backstop.
- **Migrations** are idempotent with matching up/down scripts, embedded into the
  binary and applied by a one-shot Compose service.

## Configuration

Defaults work out of the box under Compose. Overridable via environment
variables: `HTTP_PORT`, `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASS`, `DB_NAME`,
`DB_SSLMODE`.

## Layout

```
cmd/
  main.go                dispatches the server / migrate subcommands
  server/                run.go: config, DB pool, HTTP server
  migrations/            run.go + embedded up/down SQL
internal/
  infrastructure/        router, response adapter, error -> status mapping
    middleware/          Source-Type validation, recover, logging
  controller/            HTTP controllers, service interface, model/, mapper
  service/               business rules, repository interface, model/, mapper
    mocks/               generated mock of the repository interface (mockery)
  repository/            Postgres access (FOR UPDATE tx, named args), model/
  common/
    enum/                TransactionState, SourceType
    businesserror/       domain error constants
  money/                 string <-> integer-cents conversion
  config/                environment configuration
```
