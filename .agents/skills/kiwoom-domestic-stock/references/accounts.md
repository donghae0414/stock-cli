# Accounts

Use account commands to inspect current holdings before analysis or order preparation.

## Commands

```bash
stock accounts list
stock accounts list --credit-detail
```

Use `./bin/stock` instead of `stock` when working from this repository and the repo-local binary is
available.

## Output Shape

`stock accounts list` returns a JSON array of normalized holdings:

```json
[
  {
    "stock_code": "005930",
    "stock_name": "삼성전자",
    "current_price": 74100,
    "purchase_price": 70000,
    "profit_rate": 5.86,
    "purchase_amount": 70000,
    "holding_quantity": 1,
    "orderable_quantity": 1,
    "funding_type": "cash"
  }
]
```

Field meanings:

| Field | Meaning |
| --- | --- |
| `stock_code` | Six-digit Korean stock code. |
| `stock_name` | Stock name. |
| `current_price` | Current price as a JSON number. |
| `purchase_price` | Purchase price as a JSON number. |
| `profit_rate` | Percent profit/loss rounded to two decimals; `null` when purchase price is zero. |
| `purchase_amount` | Purchase amount. |
| `holding_quantity` | Current holding quantity. |
| `orderable_quantity` | Quantity currently available for order workflows. |
| `funding_type` | `cash` or `credit`. |

## Credit Detail

Use credit detail before preparing credit sell orders:

```bash
stock accounts list --credit-detail
```

Credit detail adds:

| Field | Meaning |
| --- | --- |
| `loan_date` | Kiwoom loan date for credit detail rows; empty string for cash rows. |

For credit sell with `--loan-selection individual`, use a valid `loan_date` from this output.

## Safety

Account commands are read-only. They do not need `CONFIRM` or `확인`.

Do not store full account output in persistent files unless the user asks. When reporting evidence,
prefer sanitized summaries such as row count, field names, and selected user-requested values.
