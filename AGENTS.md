# AGENTS.md

## Project Contract

- Work autonomously on clear, reversible local changes and verify before reporting completion.
- Keep diffs small, reviewable, and aligned with the existing `stock` CLI structure.
- Prefer existing Go CLI patterns in this repo: `cmd/stock/main.go`, `pkg/cmd`, `pkg/config`, `pkg/kiwoom`, `pkg/stocklookup`, and focused tests.
- Do not print or commit real Kiwoom credentials or issued tokens.
- Keep `.env` local-only and permissioned as `0600`.
- Store long-lived Kiwoom config in `~/.stock/config`; do not store issued access tokens there.
- `stock accounts list` uses an issued-token cache at `~/.stock/token.json`; keep it separate from `~/.stock/config`, permissioned as `0600`, and never print or commit token values.
- Never run live order create or cancel commands against a real account during verification unless the user explicitly requests that exact action.
- Keep command output stable, machine-readable, and documented; update the relevant `docs/*-commands.md` file when changing CLI behavior or JSON schemas.
- Keep tests isolated from the user's real `~/.stock` data, credentials, token cache, and Kiwoom network state; use temp homes, fixtures, and `httptest` where possible.
- Do not commit `.env`, `.omx/`, `.gjc/`, `.agents/`, local `.stock/` data, generated binaries, package tarballs, coverage files, or `npm/bin/` build outputs.

## Verification

- Run `gofmt` on edited Go files when available.
- Run `go test ./...` when the Go toolchain is available.
- For release or npm packaging changes, also run the relevant npm checks from `docs/maintenance.md`, especially `npm run verify:package`, `npm run smoke`, and `npm run pack:dry-run`.
- If the Go toolchain is unavailable, report that gap explicitly and use static inspection as the fallback.
