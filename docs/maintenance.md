# Stock CLI Maintenance

This document keeps maintainer-facing checks and publication notes separate
from the first-run README.

## Safe Local Checks

Before publishing, opening a PR, or preparing a release, run:

```sh
go test ./...
go vet ./...
go build -o bin/stock ./cmd/stock
npm run build:local
npm run build:targets
npm run verify:package
npm run smoke
npm run smoke:install
npm run pack:dry-run
```

`npm run smoke:install` installs the package tarball in a temporary directory
and checks the package binary layout without relying on a repo-local fallback.

## Secret and Generated-File Audit

Before publishing, check repository state and ignored local paths:

```sh
git status --short --branch
git remote -v
git check-ignore -v .env .omx/ .gjc/ .stock/ node_modules/ bin/ npm/bin/ dist/ build/ coverage.out
```

Do not commit `.env`, `.omx/`, `.gjc/`, local `.stock/` data, generated
binaries, build outputs, coverage files, Kiwoom credentials, or issued tokens.

## Command Documentation Map

| Document | Contents |
| --- | --- |
| [config-commands.md](config-commands.md) | Credential storage, token invalidation, token issue spec. |
| [account-commands.md](account-commands.md) | Account API mapping, holdings filtering, token cache policy. |
| [codes-commands.md](codes-commands.md) | Stock-code lookup ranking, envelope shape, error categories. |
| [market-commands.md](market-commands.md) | Offline tick-size table and helper boundaries. |
| [chart-commands.md](chart-commands.md) | Chart API mapping, candle output contracts, smoke guidance. |
| [order-commands.md](order-commands.md) | Order APIs, validation, output contracts, live-order cautions. |
