# Stock CLI Chart Commands

`stock chart day`, `stock chart week`, and `stock chart minute` return Kiwoom
chart candles for agent consumption. They call Kiwoom chart APIs and normalize
raw Kiwoom fields into English `snake_case` JSON.

The commands are read-only. They do not calculate indicators, export CSV,
lookup stock names, query multiple stocks, expose raw Kiwoom fields, or fetch
continuation pages beyond the first Kiwoom response.

## Commands

```sh
./bin/stock chart day --stock-code 005930
./bin/stock chart day --stock-code 005930 --count 10 --to 20260618
./bin/stock chart week --stock-code 005930 --count 10 --to 20260618
./bin/stock chart minute --stock-code 005930 --interval 1 --count 10
./bin/stock chart minute --stock-code 005930 --interval 1 --count 10 --to 20260618132000
```

## Options

Common options:

| Option | Values | Purpose |
| --- | --- | --- |
| `--stock-code` | six digits | Required stock code. |
| `--count` | `1..600` | Number of candles to return. Defaults to `120`. |
| `--to` | see below | End date or time. Defaults to current local date. |

Day/week `--to` accepts only `YYYYMMDD`.

Minute `--interval` is required and accepts `1`, `3`, `5`, `10`, `15`, `30`,
`45`, or `60`.

Minute `--to` accepts `YYYYMMDD` or `YYYYMMDDHHmmss`. Kiwoom receives only the
date part as `base_dt`; the full user-provided value is preserved in output
metadata.

## API Mapping

All chart requests use:

- Endpoint: `/api/dostk/chart`
- `upd_stkpc_tp: "1"`
- `stk_cd`: the user stock code with `_AL` appended for SOR routing
- `cont-yn: N`
- `next-key: ""`

| Command | API ID | Response list |
| --- | --- | --- |
| `stock chart day` | `ka10081` | `stk_dt_pole_chart_qry` |
| `stock chart week` | `ka10082` | `stk_stk_pole_chart_qry` |
| `stock chart minute` | `ka10080` | `stk_min_pole_chart_qry` |

## Output

The public output is one JSON object. `stock_code` appears once at the top
level and each row under `candles` is a normalized candle.

Day/week output:

```json
{
  "stock_code": "005930",
  "chart": "day",
  "to": "20260618",
  "count": 2,
  "candles": [
    {
      "date": "20260618",
      "close_price": 70100,
      "open_price": 69800,
      "high_price": 70500,
      "low_price": 69600,
      "trade_amount": 648525
    }
  ]
}
```

Minute output:

```json
{
  "stock_code": "005930",
  "chart": "minute",
  "to": "20260618132000",
  "count": 2,
  "candles": [
    {
      "timestamp": "20260618132000",
      "close_price": 78800,
      "open_price": 78850,
      "high_price": 78900,
      "low_price": 78800
    }
  ],
  "interval": 1
}
```

Kiwoom may include leading `+` or `-` signs in numeric strings. Public price and
trade amount fields are absolute JSON numbers. Dates and timestamps remain
strings.

## Safe Smoke Verification

After building, run small-count commands after configuring credentials with
`stock config set`:

```sh
./bin/stock chart day --stock-code 005930 --count 2
./bin/stock chart week --stock-code 005930 --count 2
./bin/stock chart minute --stock-code 005930 --interval 1 --count 2
```

Do not print tokens or credentials. For automation, record only sanitized
evidence such as exit status, top-level keys, candle count, and first candle
keys.
