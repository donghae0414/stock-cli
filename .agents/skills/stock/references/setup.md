# Stock CLI Setup

Guide the user through building the Stock CLI and configuring Kiwoom credentials.

## Build and Verify

From the project directory:

```bash
cd /Users/dongwuk/apps/stock-cli
go test ./...
go build -o bin/stock ./cmd/stock
./bin/stock --help
```

If `go` is installed but not on `PATH`, this project commonly uses:

```bash
export PATH="/usr/local/go/bin:$PATH"
go version
```

## Credential Configuration

Configure long-lived Kiwoom credentials interactively:

```bash
stock config set
```

When using the repo-local binary:

```bash
./bin/stock config set
```

`config set` is TTY-only and prompts for:

1. App Key
2. Secret Key

Credentials are stored in:

```text
~/.stock/config
```

The config directory should be `0700`; the config file should be `0600`.

## Credential Source

The CLI resolves credentials only from:

```text
~/.stock/config
```

Check the active credential source without exposing raw values:

```bash
stock config show
stock config path
```

`config show` masks credential values.

## Token Cache Policy

Issued Kiwoom access tokens are stored separately from long-lived credentials:

```text
~/.stock/token.json
```

Rules:

- Never store issued tokens in `~/.stock/config`.
- Never print, summarize, save, or commit token values.
- Keep `~/.stock/token.json` permissioned as `0600`.
- Treat missing, malformed, expired, or near-expiry token cache as a cache miss.
- A successful `stock config set` removes the default token cache so the next private command issues
  a fresh token.

## Private vs Local Commands

Private Kiwoom API commands require credentials and may issue or refresh a token:

- `stock accounts list`
- `stock orders list`
- `stock orders create cash`
- `stock orders create credit`
- `stock orders cancel cash`
- `stock orders cancel credit`
- `stock chart day`
- `stock chart week`
- `stock chart minute`

Local-only helper:

- `stock market tick` does not load credentials, call Kiwoom, or touch the token cache.
