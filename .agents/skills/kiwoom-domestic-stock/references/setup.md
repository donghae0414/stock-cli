# Stock CLI Setup

Guide the user through installing the Stock CLI from npm and configuring Kiwoom credentials.

## Step 1 - Check Prerequisites

Check Node.js availability:

```bash
node --version
```

Requires Node.js 18 or later. If not installed, direct the user to https://nodejs.org.

## Step 2 - Install stock CLI

Install the CLI with npm:

```bash
npm install -g @donghae0414/stock-cli
```

After installation, verify:

```bash
stock --version
```

If `stock` is not found after installation, the user may need to update their PATH. Common fixes:

- macOS/Linux:
  ```bash
  export PATH="$(npm prefix -g)/bin:$PATH"
  ```
  Add it to `~/.zshrc` or `~/.bashrc` to persist.
- Windows: Restart the terminal or add the directory from `npm prefix -g` to the system PATH.

## Step 3 - Get Kiwoom API Credentials

The CLI needs a Kiwoom REST API App Key and Secret Key for private Kiwoom endpoints. Public or
local-only helper commands do not need credentials.

Required Kiwoom permissions depend on the command:

| Permission area | Required for |
| --- | --- |
| 계좌/잔고 조회 | `accounts list` |
| 종목정보 조회 | `codes lookup` |
| 차트 조회 | `chart day`, `chart week`, `chart minute` |
| 주문 조회 | `orders list` |
| 주문 실행 | `orders create`, `orders cancel` |

## Step 4 - Configure Credentials

`stock config set` requires an interactive terminal (TTY). The user must run it directly in a
terminal that can accept hidden credential input.

```bash
stock config set
```

`config set` prompts for:

1. App Key
2. Secret Key

Credentials are saved to:

```text
~/.stock/config
```

The config directory should be `0700`; the config file should be `0600`.

## Step 5 - Verify Setup

Check the active credential source without exposing raw values:

```bash
stock config path
stock config show
```

`config show` masks credential values. Expected output includes masked `appkey`, masked
`secretkey`, and `config_file: ~/.stock/config`.

Then test with a local read-only command that never loads credentials:

```bash
stock market tick --price 353333
```

After credentials are configured, test with a private read-only command:

```bash
stock accounts list
```

An authentication or missing-credential error means the keys were not saved correctly or are not
valid for the Kiwoom API. Run `stock config set` again with the correct values.

## Credential Source

The CLI resolves credentials only from:

```text
~/.stock/config
```

There is no environment-variable or inline-flag credential priority for the current Stock CLI.

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
- `stock codes lookup`
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

## Troubleshooting

| Symptom | Fix |
| --- | --- |
| `stock: command not found` | Install with npm or update `PATH`. |
| `config set requires an interactive terminal` | Run `stock config set` directly in a real terminal. |
| `No credentials configured` | Run `stock config set` to save Kiwoom App Key and Secret Key. |
| Authentication failure on private commands | Re-run `stock config set` with valid Kiwoom credentials and required permissions. |
