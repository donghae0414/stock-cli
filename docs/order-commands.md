# Stock CLI Order Commands

`stock orders list` returns normalized Kiwoom open/unfilled order rows for agent
consumption. It calls Kiwoom `ka10075` and follows continuation pages before
printing output, up to the 100-page safety limit.

## Command

```sh
./bin/stock orders list
./bin/stock orders list --side buy
./bin/stock orders list --side sell --stock-code 005930
```

Options:

| Option | Values | Purpose |
| --- | --- | --- |
| `--side` | `all`, `buy`, `sell` | Maps to Kiwoom `trde_tp` values `0`, `2`, `1`. Defaults to `all`. |
| `--stock-code` | six digits | Filters to one stock. When absent, all stocks are queried. |

The exchange type is not exposed as an option. Requests always use integrated
exchange value `stex_tp: "0"`.

## Kiwoom API Mapping

`stock orders list` calls Kiwoom order list request `ka10075`.

```http
POST https://api.kiwoom.com/api/dostk/acnt
Content-Type: application/json;charset=UTF-8
authorization: Bearer <token>
cont-yn: N
next-key:
api-id: ka10075
```

Request body for all stocks and all sides:

```json
{
  "all_stk_tp": "0",
  "trde_tp": "0",
  "stk_cd": "",
  "stex_tp": "0"
}
```

Request body for Samsung buy orders:

```json
{
  "all_stk_tp": "1",
  "trde_tp": "2",
  "stk_cd": "005930",
  "stex_tp": "0"
}
```

If a response header returns `cont-yn: Y`, the command requests the next page
with the returned `next-key` and merges rows from every page until the response
ends or the 100-page safety limit is reached. Public output does not include
pagination metadata. A repeated `next-key` is treated as an upstream
continuation error to avoid replaying the same page indefinitely.

## Output Contract

The command returns a bare JSON array. Items contain only normalized snake_case
fields:

| Field | Source | Notes |
| --- | --- | --- |
| `order_id` | `ord_no` | Preserved as a string, including leading zeros. |
| `original_order_id` | `orig_ord_no` | Preserved as a string, including `0000000`. |
| `stock_code` | `stk_cd` | Stock code. |
| `stock_name` | `stk_nm` | Stock name. |
| `trading_venue` | `stex_tp` | `SOR` when `stex_tp == "0"`, `KRX` when `stex_tp == "1"`, `NXT` when `stex_tp == "2"`. Unknown non-empty values are preserved as `UNKNOWN_<raw>`; blank values become `UNKNOWN`. |
| `ordered_quantity` | `ord_qty` | Parsed as an absolute numeric quantity. |
| `ordered_price` | `ord_pric` | Parsed as an absolute numeric price. |
| `unfilled_quantity` | `oso_qty` | Parsed as an absolute numeric quantity. |
| `funding_type` | `io_tp_nm` | `credit` when `io_tp_nm` contains `신용`, otherwise `cash`. This heuristic is provisional until live non-empty order rows are confirmed. |
| `filled_quantity` | `cntr_qty` | Parsed as an absolute numeric quantity. |
| `current_price` | `cur_prc` | Parsed as an absolute numeric price; Kiwoom `+`/`-` prefixes are markers. |

Raw `io_tp_nm`, raw `stex_tp`, and other raw Kiwoom response fields are
intentionally not emitted. `trading_venue` is the stable agent-facing field to
carry the ordered venue/routing value needed by future cancel or amend commands.
Unknown Kiwoom `stex_tp` values are preserved in the normalized field instead of
failing the whole list, so agents can still retain order identifiers for later
operations while making the unmapped venue obvious.

`funding_type` is part of the public schema so agents can consume one stable
field, but its current cash/credit classification is provisional. Treat the
field as a normalized best-effort value until a live non-empty order row confirms
the Kiwoom `io_tp_nm` wording. When live rows become available, add a regression
fixture for the observed `io_tp_nm` values and keep the JSON field name stable.

Example shape:

```json
[
  {
    "order_id": "0000069",
    "original_order_id": "0000000",
    "stock_code": "005930",
    "stock_name": "삼성전자",
    "trading_venue": "SOR",
    "ordered_quantity": 1,
    "ordered_price": 0,
    "unfilled_quantity": 1,
    "funding_type": "cash",
    "filled_quantity": 0,
    "current_price": 74100
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

Optional live smoke verification should avoid printing raw order rows:

```sh
set -euo pipefail
umask 077
set -a
. ./.env
set +a
out="$(mktemp -t stock-orders-list.XXXXXX.json)"
export out
trap 'rm -f "$out"' EXIT
./bin/stock orders list > "$out"
python3 - <<'PY'
import json
import os

expected = {
    "order_id",
    "original_order_id",
    "stock_code",
    "stock_name",
    "trading_venue",
    "ordered_quantity",
    "ordered_price",
    "unfilled_quantity",
    "funding_type",
    "filled_quantity",
    "current_price",
}

data = json.load(open(os.environ["out"]))
assert isinstance(data, list)
assert all(set(row) == expected for row in data)
assert all(isinstance(row["order_id"], str) for row in data)
assert all(isinstance(row["original_order_id"], str) for row in data)
assert all(row["trading_venue"] in {"SOR", "KRX", "NXT", "UNKNOWN"} or row["trading_venue"].startswith("UNKNOWN_") for row in data)
print({"count": len(data), "schema": sorted(expected)})
PY
```

When the configured account has no open/unfilled orders, `count` can be `0`.
Treat that as normal live state, not a failure.
