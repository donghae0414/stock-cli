# Stock CLI Market Commands

`stock market` contains local market-rule helpers for agent workflows. These
commands do not call Kiwoom, do not load credentials, and do not touch the token
cache.

## Command

```sh
./bin/stock market tick --price 353333
```

Options:

| Option | Values | Purpose |
| --- | --- | --- |
| `--price` | positive whole-won integer | Price to check against the general Korean stock tick-size table. |

## Output Contract

The command returns one normalized JSON object:

```json
{
  "price": 353333,
  "tick_size": 500,
  "lower_price": 353000,
  "upper_price": 353500,
  "is_valid_tick": false
}
```

Fields:

| Field | Notes |
| --- | --- |
| `price` | Input price as a JSON number. |
| `tick_size` | Tick size selected for the input price band. |
| `lower_price` | Greatest valid tick price less than or equal to `price`. |
| `upper_price` | Smallest valid tick price greater than or equal to `price`. |
| `is_valid_tick` | `true` when `price` already matches the tick grid. |

## General Stock Tick Table

| Price range | Tick size |
| --- | ---: |
| Less than 2,000 | 1 |
| 2,000 through 4,999 | 5 |
| 5,000 through 19,999 | 10 |
| 20,000 through 49,999 | 50 |
| 50,000 through 199,999 | 100 |
| 200,000 through 499,999 | 500 |
| 500,000 or more | 1,000 |

## First-Pass Boundaries

This helper is intentionally offline. It does not call Kiwoom `ka10004`, does
not fetch the current orderbook, and does not modify `stock orders create`
validation. ETF, ETN, ELW, preferred-stock, product-type, and venue-specific
exceptions are also out of scope for this first pass.

## Verification Checklist

Before reporting completion:

```sh
gofmt -w pkg/cmd/market.go pkg/cmd/market_test.go pkg/cmd/cmd.go
go test ./...
go build -o bin/stock ./cmd/stock
./bin/stock market --help
./bin/stock market tick --price 353333
./bin/stock market tick --price 2001
```
