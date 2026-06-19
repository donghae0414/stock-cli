# Market Helpers

Use local market helpers for Korean stock market rules that do not require Kiwoom API calls.

## Tick Size

```bash
stock market tick --price 353333
```

Output:

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

| Field | Meaning |
| --- | --- |
| `price` | Input price. |
| `tick_size` | General stock tick size for the input price band. |
| `lower_price` | Greatest valid tick price less than or equal to input. |
| `upper_price` | Smallest valid tick price greater than or equal to input. |
| `is_valid_tick` | Whether the input already matches the tick grid. |

General stock tick table:

| Price range | Tick size |
| --- | ---: |
| Less than 2,000 | 1 |
| 2,000 through 4,999 | 5 |
| 5,000 through 19,999 | 10 |
| 20,000 through 49,999 | 50 |
| 50,000 through 199,999 | 100 |
| 200,000 through 499,999 | 500 |
| 500,000 or more | 1,000 |

## Boundaries

`stock market tick` is offline. It does not:

- Call Kiwoom.
- Load credentials.
- Touch `~/.stock/token.json`.
- Fetch orderbooks.
- Apply ETF, ETN, ELW, preferred-stock, or venue-specific exceptions.
- Change `orders create` validation.

