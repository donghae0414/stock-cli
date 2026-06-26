# Stock CLI Initial Configuration Commands

`stock config` is the initial configuration surface for Kiwoom-specific
credentials. Kiwoom REST API host is fixed and is not user configuration.

## Stock CLI Commands

`stock-cli` provides the same initial configuration command set:

| Stock command | Purpose |
| --- | --- |
| `stock config set` | Set Kiwoom REST API credentials interactively. |
| `stock config show` | Show current Kiwoom credentials in masked form and show their source. |
| `stock config path` | Print the Stock CLI config file path. |

For Kiwoom, stored long-lived credentials are only `appkey` and `secretkey`.
Runtime token issuance uses the fixed Kiwoom host `https://api.kiwoom.com`.
Issued access tokens are not part of these initial config commands.

## Stock CLI Credential Storage

Store Kiwoom credentials in:

```text
~/.stock/config
```

Use TOML with one Kiwoom credential section:

```toml
[kiwoom]
appkey = "..."
secretkey = "..."
```

Credential resolution uses only the config file:

1. Config file: `~/.stock/config`.
2. Missing credential message when the config file does not provide credentials.

Create `~/.stock/` with `0700` permissions and write `~/.stock/config` with
`0600` permissions. `stock config set` is intentionally TTY-only. It prompts
for App Key and Secret Key only.

After `stock config set` successfully saves credentials, it removes the default
issued-token cache at `~/.stock/token.json`. A missing token cache is a harmless
no-op. The next token-using command, such as `stock accounts list`, will issue a
new token using the newly saved config credentials. This invalidation is scoped
to successful `stock config set`.

## Confirmed Kiwoom Token Issue Spec

Confirmed by direct call on 2026-06-10 KST:

```http
POST https://api.kiwoom.com/oauth2/token
Content-Type: application/json;charset=UTF-8
```

Request body:

```json
{
  "grant_type": "client_credentials",
  "appkey": "...",
  "secretkey": "..."
}
```

Observed successful response:

- HTTP status: `200`
- `expires_dt`: compact `YYYYMMDDHHMMSS` string
- `token_type`: observed as `Bearer`
- `token`: access token string
- `return_code`: observed as `0`
- `return_msg`: observed as `정상적으로 처리되었습니다`

The successful token issue response did not include `next-key`, `cont-yn`, or
`api-id` headers in the observed response.

## Token Storage Policy

Do not store issued API tokens in `~/.stock/config`. The config file stores only
long-lived user credentials.

Token caching is implemented by API commands that need Kiwoom access tokens,
starting with `stock accounts list`. The cache uses `~/.stock/token.json`,
stays separate from `~/.stock/config`, and must not print or commit issued token
values. Successful `stock config set` removes this cache so the next token-using
command issues a fresh token using the active credential source.
