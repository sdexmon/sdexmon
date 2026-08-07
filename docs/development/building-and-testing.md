# Building and testing

## Requirements

The Go toolchain version declared in `go.mod`. No other tooling is required to
build or test. Regenerating `README.ansi` needs `python3`.

## Running during development

```bash
./run
```

The `run` script loads a local `.env` if present, sets `HORIZON_URL` and
`DEBUG=true`, requests a 140x60 terminal, injects the version and commit from
`git describe`, and executes `go run ./cmd/sdexmon`.

Without the helper:

```bash
go run ./cmd/sdexmon
```

## Makefile targets

```
make build         go build -o sdexmon ./cmd/sdexmon
make fmt           go fmt ./...
make vet           go vet ./...
make test          go test ./...
make readme        cat README.ansi
make readme-gen    regenerate README.ansi from README.md
make readme-check  fail if README.ansi has drifted from README.md
```

## Building

```bash
go build -o sdexmon ./cmd/sdexmon
./sdexmon --version
```

A plain build reports `dev (build unknown)`, because `appVersion` and
`gitCommit` are only set through linker flags.

With version information:

```bash
go build -ldflags="\
  -X main.appVersion=$(git describe --tags --always) \
  -X main.gitCommit=$(git rev-parse --short HEAD)" \
  -o sdexmon ./cmd/sdexmon
```

Cross-compiling for a Raspberry Pi, matching what CI checks:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build ./cmd/sdexmon
```

## Testing

```bash
go test ./...
go test -race ./...
go test -run '^TestName$' ./...
```

CI runs `go test -race ./...` and `go vet ./...` on every push to `main` and on
every pull request, plus a `linux/arm64` cross-build.

### What is covered

- `cmd/sdexmon/upgrade_test.go` -- upgrade notice behaviour.
- `cmd/sdexmon/maintenance_test.go` -- maintenance routing, search source
  selection, and pair removal.
- `cmd/sdexmon/layout_test.go` -- panel fitting across terminal sizes.
- `cmd/sdexmon/main_test.go` -- order book request shape.
- `internal/config/user_config_test.go` -- pair add, remove and list.
- `internal/stellar/toml_test.go` -- stellar.toml parsing and issuer validation.
- `internal/version/checker_test.go` -- semantic version comparison.

### What is not covered

View snapshots and mocked stellar.expert asset searches. External dependencies
(Horizon and stellar.expert) are not mocked anywhere, so nothing in the test
suite touches the network.

## Before committing

```bash
make fmt
make vet
make test
```

If you edited `README.md`, also run `make readme-gen` and commit both files.
`make readme-check` fails the build when they drift.

## Code standards

- Packages: lowercase, single word. Files: snake_case.
- Exported identifiers: PascalCase. Unexported: camelCase.
- Interfaces: PascalCase with an "er" suffix where it reads naturally.
- Handle every error explicitly.
- Use `context.Context` for cancellation and timeouts on outbound calls.
- Prefer composition over inheritance.
- Keep comments about intent and constraints, not restatements of the code. The
  existing comments in `main.go` about terminal auto-wrap, alternate-screen
  corruption, and order book direction are the model to follow.

## Stellar-specific rules

- Amounts never exceed 7 decimal places. Displays show at least 2.
- Thousand separators are spaces, not commas.
- Do not assume a currency symbol.
- The issuer is part of the asset identity. Never compare or deduplicate assets
  on code alone.
- Development runs against mainnet with the public Horizon endpoint. SDEXMON is
  read-only and submits no transactions.

## Adding a screen

See [Screens and routing](../architecture/screens-and-routing.md#adding-a-screen).

## Adding a trading pair

Pairs live in `~/.config/sdexmon/config.yaml`, not in code. See
[Pair management](../guide/pair-management.md). The curated tables in
`cmd/sdexmon/main.go` are fallbacks only, and the similar tables in
`internal/models/constants.go` are dead code.
