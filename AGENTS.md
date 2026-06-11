# AGENTS.md

## Project Contract

- Work autonomously on clear, reversible local changes and verify before reporting completion.
- Keep diffs small, reviewable, and aligned with `/Users/dongwuk/apps/upbit-cli` where this project intentionally mirrors it.
- Prefer existing Go CLI patterns from `upbit-cli`: `cmd/<binary>/main.go`, `pkg/cmd`, `pkg/config`, and focused tests.
- Do not print or commit real Kiwoom credentials or issued tokens.
- Keep `.env` local-only and permissioned as `0600`.
- Store long-lived Kiwoom config in `~/.stock/config`; do not store issued access tokens there.
- `stock accounts list` uses an issued-token cache at `~/.stock/token.json`; keep it separate from `~/.stock/config`, permissioned as `0600`, and never print or commit token values.
- Trading API commands are future work unless explicitly requested.

## Verification

- Run `gofmt` on edited Go files when available.
- Run `go test ./...` when the Go toolchain is available.
- If the Go toolchain is unavailable, report that gap explicitly and use static inspection as the fallback.
