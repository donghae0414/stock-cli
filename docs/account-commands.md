# Stock CLI Account Commands

`stock-cli` mirrors the `upbit accounts list` command shape while using Kiwoom
REST APIs and returning normalized account holdings for agent consumption.

## Upbit CLI Reference

Confirmed Upbit account command:

| Upbit command | Purpose |
| --- | --- |
| `upbit accounts list` | Get account balances. |

Reference source:

- `/Users/dongwuk/apps/upbit-cli/pkg/cmd/account.go`
- `/Users/dongwuk/apps/upbit-cli/docs/commands.md`

Important Upbit command-shape facts:

- Command path is `accounts list`.
- The command has no command-specific options.
- Unexpected extra arguments are rejected.

## Stock CLI Command

`stock-cli` provides:

| Stock command | Purpose |
| --- | --- |
| `stock accounts list` | Return normalized current Kiwoom account holdings. |
| `stock accounts list --credit-detail` | Return cash holdings plus loan-date credit detail rows. |

Command-specific options:

| Option | Purpose |
| --- | --- |
| `--credit-detail` | Show Kiwoom credit detail rows by loan date while preserving cash rows. |

## Kiwoom Token Policy

Long-lived Kiwoom credentials remain in:

```text
~/.stock/config
```

Issued access tokens are cached separately in:

```text
~/.stock/token.json
```

Token cache rules:

- Never store issued access tokens in `~/.stock/config`.
- Create `~/.stock/` with `0700` permissions when supported.
- Write `~/.stock/token.json` with `0600` permissions when supported.
- Treat missing, malformed, unreadable, expired, or near-expiry token cache as a cache miss.
- Refresh when the cached token expires within 1 minute.
- Replace the cache only after a successful token issue response.
- Never print or commit issued token values.

## Kiwoom API Mapping

### Token issue

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

Response fields used:

| Field | Meaning |
| --- | --- |
| `token` | Issued access token. |
| `token_type` | Expected bearer token type. |
| `expires_dt` | Token expiry in `YYYYMMDDHHMMSS` format. |
| `return_code` | Kiwoom business result code. |
| `return_msg` | Kiwoom business result message. |

### Account holdings

`stock accounts list` calls Kiwoom account profit-rate request `ka10085`.

```http
POST https://api.kiwoom.com/api/dostk/acnt
Content-Type: application/json;charset=UTF-8
authorization: Bearer <token>
cont-yn: N
next-key:
api-id: ka10085
```

Request body:

```json
{
  "stex_tp": "0"
}
```

## Output Contract

The command returns a bare JSON array.

Default `stock accounts list` items contain only normalized snake_case fields:

| Field | Source | Notes |
| --- | --- | --- |
| `stock_code` | `stk_cd` | Stock code. |
| `stock_name` | `stk_nm` | Stock name. |
| `current_price` | `cur_prc` | Parsed as absolute numeric price; Kiwoom `+`/`-` prefix is treated as a marker. |
| `purchase_price` | `pur_pric` | Parsed as numeric price. |
| `profit_rate` | computed | `(current_price - purchase_price) / purchase_price * 100`, rounded to 2 decimal places; `null` when purchase price is `0`. |
| `purchase_amount` | `pur_amt` | Parsed numeric purchase amount. |
| `holding_quantity` | `rmnd_qty` | Parsed numeric holding quantity. |
| `orderable_quantity` | `clrn_alow_qty` | Parsed numeric orderable quantity. |
| `is_credit` | `crd_tp` | `false` when `crd_tp == "00"`, otherwise `true`. |

Filtering rules:

- Exclude rows where `rmnd_qty` is `0`.
- Exclude rows where `stk_nm` starts with `*`.
- Return only unstarred aggregate rows.

Example shape:

```json
[
  {
    "stock_code": "000001",
    "stock_name": "Synthetic Alpha",
    "current_price": 1200,
    "purchase_price": 1000,
    "profit_rate": 20.00,
    "purchase_amount": 3000,
    "holding_quantity": 3,
    "orderable_quantity": 2,
    "is_credit": false
  }
]
```

### Credit detail output

`stock accounts list --credit-detail` returns the default fields plus one
additional field:

| Field | Source | Notes |
| --- | --- | --- |
| `loan_date` | `loan_dt` | Kiwoom loan date for credit detail rows; empty string for cash rows. |

Filtering rules:

- Exclude rows where `rmnd_qty` is `0`.
- Include every cash row where `crd_tp == "00"` and set `loan_date` to an empty string.
- Include credit detail rows where `stk_nm` starts with `*` and `crd_tp != "00"`.
- Remove the leading `*` from displayed credit detail `stock_name`.
- Exclude unstarred credit aggregate rows where `crd_tp != "00"`.

Example shape:

```json
[
  {
    "stock_code": "000003",
    "stock_name": "Synthetic Credit Detail",
    "current_price": 2100,
    "purchase_price": 2000,
    "profit_rate": 5.00,
    "purchase_amount": 10000,
    "holding_quantity": 5,
    "orderable_quantity": 5,
    "is_credit": true,
    "loan_date": "20260601"
  }
]
```

## Verification Checklist

Before reporting completion:

```sh
gofmt -w <edited-go-files>
go test ./...
go build -o bin/stock ./cmd/stock
```

Final direct CLI smoke verification should avoid leaving raw holdings on disk:

```sh
set -euo pipefail
umask 077
set -a
. ./.env
set +a
out="$(mktemp -t stock-accounts-list.XXXXXX.json)"
export out
trap 'rm -f "$out"' EXIT
./bin/stock accounts list > "$out"
python3 - <<'PY'
import json
import os

expected = {
    "stock_code",
    "stock_name",
    "current_price",
    "purchase_price",
    "profit_rate",
    "purchase_amount",
    "holding_quantity",
    "orderable_quantity",
    "is_credit",
}

data = json.load(open(os.environ["out"]))
assert isinstance(data, list)
assert all(set(row) == expected for row in data)
assert all(not row["stock_name"].startswith("*") for row in data)
assert all(row["holding_quantity"] != 0 for row in data)
print({"count": len(data), "schema": sorted(expected)})
PY
```

Report only the summarized `count` and `schema` unless raw portfolio output is
explicitly requested.

Credit-detail smoke verification should also avoid printing raw holdings:

```sh
set -euo pipefail
umask 077
set -a
. ./.env
set +a
default_out="$(mktemp -t stock-accounts-list.XXXXXX.json)"
detail_out="$(mktemp -t stock-accounts-credit-detail.XXXXXX.json)"
export default_out detail_out
trap 'rm -f "$default_out" "$detail_out"' EXIT
./bin/stock accounts list > "$default_out"
./bin/stock accounts list --credit-detail > "$detail_out"
python3 - <<'PY'
import json
import os

default_expected = {
    "stock_code",
    "stock_name",
    "current_price",
    "purchase_price",
    "profit_rate",
    "purchase_amount",
    "holding_quantity",
    "orderable_quantity",
    "is_credit",
}
detail_expected = default_expected | {"loan_date"}

default_data = json.load(open(os.environ["default_out"]))
detail_data = json.load(open(os.environ["detail_out"]))
credit_detail_rows = [
    row for row in detail_data
    if row["is_credit"] is True and row["loan_date"] != ""
]

assert isinstance(default_data, list)
assert isinstance(detail_data, list)
assert all(set(row) == default_expected for row in default_data)
assert all(set(row) == detail_expected for row in detail_data)
assert all(not row["stock_name"].startswith("*") for row in detail_data)
assert all(row["holding_quantity"] != 0 for row in default_data + detail_data)
print({
    "default_count": len(default_data),
    "detail_count": len(detail_data),
    "detail_schema": sorted(detail_expected),
    "credit_detail_rows_with_loan_date": len(credit_detail_rows),
})
PY
```

When the configured account currently has no credit rows, the final
`credit_detail_rows_with_loan_date` count can be `0`; treat that as live-state
inconclusive after confirming the raw `ka10085` response has no
`crd_tp != "00"` rows.
