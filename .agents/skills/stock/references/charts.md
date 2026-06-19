# Charts

Use chart commands for read-only Kiwoom candle queries.

## Commands

```bash
stock chart day --stock-code 005930
stock chart day --stock-code 005930 --count 10 --to 20260618
stock chart week --stock-code 005930 --count 10 --to 20260618
stock chart minute --stock-code 005930 --interval 1 --count 10
stock chart minute --stock-code 005930 --interval 1 --count 10 --to 20260618132000
```

## Options

| Option | Values | Purpose |
| --- | --- | --- |
| `--stock-code` | six digits | Required stock code. |
| `--count` | `1..600` | Number of candles. Defaults to `120`. |
| `--to` | `YYYYMMDD` or minute timestamp | End date/time. |
| `--interval` | `1`, `3`, `5`, `10`, `15`, `30`, `45`, `60` | Required for minute charts. |

Day/week `--to` accepts `YYYYMMDD`. Minute `--to` accepts `YYYYMMDD` or
`YYYYMMDDHHmmss`.

## Output

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

## Safety

Chart commands are read-only. They do not need `CONFIRM` or `확인`.

Do not expose credentials or token values when reporting command evidence.
