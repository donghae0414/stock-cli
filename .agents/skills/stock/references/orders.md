# Orders

Use order commands to list, create, or cancel Kiwoom stock orders.

`orders create` and `orders cancel` are live trading commands. They submit Kiwoom requests
immediately, so the confirmation protocol in `SKILL.md` is mandatory.

The confirmation gate is enforced by the agent workflow, not by the `stock` binary. Do not tell the
user that direct CLI invocation is protected unless CLI-side confirmation has been implemented.

## List Open Orders

Read-only:

```bash
stock orders list
stock orders list --side buy
stock orders list --side sell --stock-code 005930
```

Options:

| Option | Values | Purpose |
| --- | --- | --- |
| `--side` | `all`, `buy`, `sell` | Defaults to `all`. |
| `--stock-code` | six digits | Optional stock filter. |

Output:

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

## Cash Buy/Sell

Live trading command:

```bash
stock orders create cash \
  --side buy \
  --stock-code 005930 \
  --order-type limit \
  --quantity 1 \
  --price 74100 \
  --trading-venue SOR
```

```bash
stock orders create cash \
  --side sell \
  --stock-code 005930 \
  --order-type market \
  --quantity 1 \
  --trading-venue SOR
```

Options:

| Option | Values | Purpose |
| --- | --- | --- |
| `--side` | `buy`, `sell` | Buy or sell. |
| `--stock-code` | six digits | Required stock code. |
| `--order-type` | `limit`, `market` | Limit or market order style. |
| `--quantity` | positive integer | Whole-share quantity. |
| `--price` | positive integer | Required for limit orders; rejected for market orders. It is a per-share price, not a total amount. |
| `--trading-venue` | `SOR`, `KRX`, `NXT` | Defaults to `SOR`. |

Limit orders require `--price`. Market orders must omit `--price`.

## Credit Buy/Sell

Live trading command:

```bash
stock orders create credit \
  --side buy \
  --stock-code 005930 \
  --order-type limit \
  --quantity 1 \
  --price 74100 \
  --trading-venue SOR
```

Credit sell requires loan selection:

```bash
stock orders create credit \
  --side sell \
  --stock-code 005930 \
  --order-type limit \
  --quantity 3 \
  --price 6450 \
  --trading-venue SOR \
  --loan-selection aggregate
```

```bash
stock orders create credit \
  --side sell \
  --stock-code 005930 \
  --order-type limit \
  --quantity 3 \
  --price 6450 \
  --trading-venue SOR \
  --loan-selection individual \
  --loan-date 20260601
```

Additional credit sell options:

| Option | Values | Purpose |
| --- | --- | --- |
| `--loan-selection` | `individual`, `aggregate` | Required for credit sell only. |
| `--loan-date` | `YYYYMMDD` | Required with `individual`; rejected with `aggregate`; invalid for credit buy. |

Use `stock accounts list --credit-detail` before individual-loan credit sells to find available
loan dates.

## Cancel Orders

Live trading command:

```bash
stock orders cancel cash \
  --stock-code 005930 \
  --original-order-id 0000140 \
  --trading-venue SOR
```

```bash
stock orders cancel credit \
  --stock-code 005930 \
  --original-order-id 0001615 \
  --trading-venue SOR
```

Options:

| Option | Values | Purpose |
| --- | --- | --- |
| `--stock-code` | six digits | Required stock code. |
| `--original-order-id` | digits | Required original order id; preserve leading zeros. |
| `--quantity` | positive integer | Optional cancel quantity. Omit to cancel all remaining quantity. |
| `--trading-venue` | `SOR`, `KRX`, `NXT` | Defaults to `SOR`. |

## Confirmation Template

Before live execution, send a message like:

```text
이 주문은 Kiwoom live 주문을 즉시 전송합니다. 아래 명령을 실행하려면 CONFIRM 또는 확인 중 하나를 한 단어로만 입력해 주세요.
```

Then show:

```bash
stock orders create cash --side buy --stock-code 005930 --order-type limit --quantity 1 --price 74100 --trading-venue SOR
```

Risk summary:

- Side: buy
- Funding type: cash
- Stock code: 005930
- Order type: limit
- Quantity: 1
- Price behavior: per-share limit price 74100
- Trading venue: SOR

Run the command only if the next user message is exactly one of these single-word confirmations:

```text
CONFIRM
확인
```

## Result Handling

Successful create output:

```json
{
  "order_id": "0000024",
  "trading_venue": "SOR"
}
```

Successful cancel output:

```json
{
  "order_id": "0000141",
  "base_original_order_id": "0000140",
  "cancelled_quantity": 1
}
```

Summarize returned order ids and trading venue. Do not claim fill status from create output alone;
use `orders list` or later account/order evidence when needed.
