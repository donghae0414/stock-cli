# Glossary

Use this glossary when translating command output or explaining CLI fields.

| CLI / English | Korean | Notes |
| --- | --- | --- |
| stock code | 종목 코드 | Six-digit Korean stock code. |
| stock name | 종목명 | Display name. |
| holdings | 보유종목 | Current account holdings. |
| current price | 현재가 | JSON field `current_price`. |
| purchase price | 매입가 | JSON field `purchase_price`. |
| profit rate | 수익률 | JSON field `profit_rate`. |
| purchase amount | 매입금액 | JSON field `purchase_amount`. |
| holding quantity | 보유수량 | JSON field `holding_quantity`. |
| orderable quantity | 주문가능수량 | JSON field `orderable_quantity`. |
| funding type | 자금 구분 | `cash` or `credit`. |
| cash | 현금 | Cash order or holding. |
| credit | 신용 | Credit order or holding. |
| loan date | 대출일 | Required for individual-loan credit sells. |
| open orders | 미체결 주문 | `stock orders list`. |
| order id | 주문번호 | Preserve leading zeros when provided as a string. |
| original order id | 원주문번호 | Used for cancel commands. |
| buy | 매수 | CLI side `buy`. |
| sell | 매도 | CLI side `sell`. |
| cancel | 취소 | Live cancel command. |
| limit order | 지정가 | `--order-type limit`; `--price` is per-share price. |
| market order | 시장가 | `--order-type market`; omit `--price`. |
| quantity | 수량 | Whole-share quantity. |
| trading venue | 거래소 경로 | `SOR`, `KRX`, or `NXT`. |
| tick size | 호가 단위 | `stock market tick`. |
| lower price | 아래 유효 호가 | JSON field `lower_price`. |
| upper price | 위 유효 호가 | JSON field `upper_price`. |
| valid tick | 유효 호가 여부 | JSON field `is_valid_tick`. |

