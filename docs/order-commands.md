# Stock CLI Order 명령

`stock orders list`는 agent가 소비할 수 있도록 정규화된 Kiwoom 미체결 주문 row를
반환합니다. Kiwoom `ka10075`를 호출하고, 출력 전에 continuation page를 따라가며,
100 page safety limit까지 처리합니다.

`stock orders create cash`, `stock orders create credit`, `stock orders cancel
cash`, `stock orders cancel credit`은 live Kiwoom 주문 요청을 즉시 제출합니다.
확인, dry-run 정책, quote check, 현금 check, 담보 check, 대출 check, 보유수량
check, 가격제한 check는 이 CLI primitive가 아니라 이 CLI를 호출하는 Skill 또는
workflow의 책임입니다.

## 명령

```sh
./bin/stock orders list
./bin/stock orders list --side buy
./bin/stock orders list --side sell --stock-code 005930
./bin/stock orders create cash --side buy --stock-code 005930 --order-type limit --quantity 1 --price 74100
./bin/stock orders create cash --side sell --stock-code 005930 --order-type market --quantity 1
./bin/stock orders cancel cash --stock-code 005930 --original-order-id 0000140
./bin/stock orders create credit --side buy --stock-code 005930 --order-type limit --quantity 1 --price 74100
./bin/stock orders create credit --side sell --stock-code 005930 --order-type limit --quantity 3 --price 6450 --loan-selection aggregate
./bin/stock orders create credit --side sell --stock-code 005930 --order-type limit --quantity 3 --price 6450 --loan-selection individual --loan-date 20260601
./bin/stock orders cancel credit --stock-code 005930 --original-order-id 0001615
```

옵션:

| 옵션 | 값 | 목적 |
| --- | --- | --- |
| `--side` | `all`, `buy`, `sell` | Kiwoom `trde_tp` 값 `0`, `2`, `1`로 mapping합니다. 기본값은 `all`입니다. |
| `--stock-code` | 6자리 숫자 | 한 종목으로 filter합니다. 없으면 모든 종목을 조회합니다. |

거래소 구분은 옵션으로 노출하지 않습니다. 요청은 항상 통합 거래소 값
`stex_tp: "0"`을 사용합니다.

## Cash Create 명령

`stock orders create cash`는 Kiwoom 현금 매수 또는 현금 매도 주문 API를
호출합니다.

옵션:

| 옵션 | 값 | 목적 |
| --- | --- | --- |
| `--side` | `buy`, `sell` | `buy`는 `kt10000`, `sell`은 `kt10001`을 호출합니다. |
| `--stock-code` | 6자리 숫자 | 필수 종목 코드입니다. |
| `--order-type` | `limit`, `market` | `limit`은 Kiwoom `trde_tp: "0"`, `market`은 `trde_tp: "3"`으로 mapping합니다. |
| `--quantity` | 양의 정수 | 정수 주식 주문수량입니다. |
| `--price` | 양의 정수 | 1주당 지정가 가격입니다. `limit`에서는 필수이고, `market`에서는 거부됩니다. 총 주문금액이 아닙니다. |
| `--trading-venue` | `SOR`, `KRX`, `NXT` | 기본값은 `SOR`이며 Kiwoom `dmst_stex_tp`로 mapping합니다. |

지정가 매수 예시:

```sh
./bin/stock orders create cash --side buy --stock-code 005930 --order-type limit --quantity 1 --price 74100
```

시장가 매도 예시:

```sh
./bin/stock orders create cash --side sell --stock-code 005930 --order-type market --quantity 1
```

`--order-type market`과 함께 `--price`를 전달하면 credential 로드나 Kiwoom 요청
전에 CLI validation error가 됩니다.

지정가 매수 요청 body:

```json
{
  "dmst_stex_tp": "SOR",
  "stk_cd": "005930",
  "ord_qty": "1",
  "ord_uv": "74100",
  "trde_tp": "0",
  "cond_uv": ""
}
```

시장가 매도 요청 body:

```json
{
  "dmst_stex_tp": "SOR",
  "stk_cd": "005930",
  "ord_qty": "1",
  "ord_uv": "",
  "trde_tp": "3",
  "cond_uv": ""
}
```

성공 출력은 정규화됩니다.

```json
{
  "order_id": "0000024",
  "trading_venue": "SOR"
}
```

## Cash Cancel 명령

`stock orders cancel cash`는 Kiwoom 현금 취소 API `kt10003`을 호출합니다.

옵션:

| 옵션 | 값 | 목적 |
| --- | --- | --- |
| `--stock-code` | 6자리 숫자 | 필수 종목 코드입니다. |
| `--original-order-id` | 숫자 | 필수 원주문 번호입니다. 선행 zero를 포함해 문자열로 보존합니다. |
| `--quantity` | 양의 정수 | 선택 취소 수량입니다. 생략하면 `cncl_qty: "0"`으로 잔량 전체를 취소합니다. |
| `--trading-venue` | `SOR`, `KRX`, `NXT` | 기본값은 `SOR`이며 Kiwoom `dmst_stex_tp`로 mapping합니다. |

잔량 전체 취소:

```sh
./bin/stock orders cancel cash --stock-code 005930 --original-order-id 0000140
```

요청 body:

```json
{
  "dmst_stex_tp": "SOR",
  "orig_ord_no": "0000140",
  "stk_cd": "005930",
  "cncl_qty": "0"
}
```

성공 출력은 정규화됩니다.

```json
{
  "order_id": "0000141",
  "base_original_order_id": "0000140",
  "cancelled_quantity": 1
}
```

## Credit Create 명령

`stock orders create credit`은 Kiwoom 신용 매수 또는 신용 매도 주문 API를
호출합니다.

옵션:

| 옵션 | 값 | 목적 |
| --- | --- | --- |
| `--side` | `buy`, `sell` | `buy`는 `kt10006`, `sell`은 `kt10007`을 호출합니다. |
| `--stock-code` | 6자리 숫자 | 필수 종목 코드입니다. |
| `--order-type` | `limit`, `market` | `limit`은 Kiwoom `trde_tp: "0"`, `market`은 `trde_tp: "3"`으로 mapping합니다. |
| `--quantity` | 양의 정수 | 정수 주식 주문수량입니다. |
| `--price` | 양의 정수 | 1주당 지정가 가격입니다. `limit`에서는 필수이고, `market`에서는 거부됩니다. 총 주문금액이 아닙니다. |
| `--trading-venue` | `SOR`, `KRX`, `NXT` | 기본값은 `SOR`이며 Kiwoom `dmst_stex_tp`로 mapping합니다. |
| `--loan-selection` | `individual`, `aggregate` | 신용 매도에서만 필수입니다. `individual`은 `crd_deal_tp: "33"`, `aggregate`는 `crd_deal_tp: "99"`로 mapping합니다. |
| `--loan-date` | `YYYYMMDD` | `--loan-selection individual`과 함께 필수입니다. `aggregate`에서는 거부되고, 신용 매수에서도 거부됩니다. |

신용 매수 지정가 예시:

```sh
./bin/stock orders create credit --side buy --stock-code 005930 --order-type limit --quantity 1 --price 74100
```

신용 매도 aggregate 예시:

```sh
./bin/stock orders create credit --side sell --stock-code 005930 --order-type limit --quantity 3 --price 6450 --loan-selection aggregate
```

신용 매도 individual loan-date 예시:

```sh
./bin/stock orders create credit --side sell --stock-code 005930 --order-type limit --quantity 3 --price 6450 --loan-selection individual --loan-date 20260601
```

`--order-type market`과 함께 `--price`를 전달하면 credential 로드나 Kiwoom 요청
전에 CLI validation error가 됩니다. `--loan-selection aggregate`와 함께
`--loan-date`를 전달하는 것도 credential 또는 network 전에 validation error가
됩니다.

신용 매수 지정가 주문 요청 body:

```json
{
  "dmst_stex_tp": "SOR",
  "stk_cd": "005930",
  "ord_qty": "1",
  "ord_uv": "74100",
  "trde_tp": "0",
  "cond_uv": ""
}
```

신용 매도 aggregate 지정가 주문 요청 body:

```json
{
  "dmst_stex_tp": "SOR",
  "stk_cd": "005930",
  "ord_qty": "3",
  "ord_uv": "6450",
  "trde_tp": "0",
  "crd_deal_tp": "99",
  "crd_loan_dt": "",
  "cond_uv": ""
}
```

신용 매도 individual 시장가 주문 요청 body:

```json
{
  "dmst_stex_tp": "SOR",
  "stk_cd": "005930",
  "ord_qty": "3",
  "ord_uv": "",
  "trde_tp": "3",
  "crd_deal_tp": "33",
  "crd_loan_dt": "20260601",
  "cond_uv": ""
}
```

성공 출력은 정규화됩니다.

```json
{
  "order_id": "0001615",
  "trading_venue": "SOR"
}
```

## Credit Cancel 명령

`stock orders cancel credit`은 Kiwoom 신용 취소 API `kt10009`를 호출합니다.

옵션:

| 옵션 | 값 | 목적 |
| --- | --- | --- |
| `--stock-code` | 6자리 숫자 | 필수 종목 코드입니다. |
| `--original-order-id` | 숫자 | 필수 원주문 번호입니다. 선행 zero를 포함해 문자열로 보존합니다. |
| `--quantity` | 양의 정수 | 선택 취소 수량입니다. 생략하면 `cncl_qty: "0"`으로 잔량 전체를 취소합니다. |
| `--trading-venue` | `SOR`, `KRX`, `NXT` | 기본값은 `SOR`이며 Kiwoom `dmst_stex_tp`로 mapping합니다. |

잔량 전체 취소:

```sh
./bin/stock orders cancel credit --stock-code 005930 --original-order-id 0001615
```

요청 body:

```json
{
  "dmst_stex_tp": "SOR",
  "orig_ord_no": "0001615",
  "stk_cd": "005930",
  "cncl_qty": "0"
}
```

성공 출력은 정규화됩니다.

```json
{
  "order_id": "0001695",
  "base_original_order_id": "0001615",
  "cancelled_quantity": 1
}
```

## 오류 출력

성공 JSON은 최소 필드만 유지하고 Kiwoom `return_code` 또는 `return_msg`를 포함하지
않습니다. 오류 출력은 다릅니다. Kiwoom이 nonzero business `return_code`를
반환하면 CLI는 parse된 Kiwoom `return_msg`를 error에 포함하여 사람과 CLI 사용
agent가 upstream API 실패를 진단할 수 있게 합니다.

오류 shape 예시:

```text
Kiwoom cash buy order failed return_code=9 return_msg="cash order rejected"
```

```text
Kiwoom credit sell order failed return_code=9 return_msg="credit order rejected"
```

HTTP error summary도 Kiwoom이 제공하는 경우 parse된 `return_msg`를 포함합니다.
전체 raw response body, 발급 token 필드, `authorization` 필드, app key, secret
key는 dump하지 않습니다. `return_msg`는 Kiwoom이 제어하는 diagnostic output으로
취급하세요. 복구 loop에는 유용하지만 public log에 그대로 복사할 대상은 아닙니다.

## Kiwoom API Mapping

`stock orders list`는 Kiwoom 주문 조회 요청 `ka10075`를 호출합니다.

```http
POST https://api.kiwoom.com/api/dostk/acnt
Content-Type: application/json;charset=UTF-8
authorization: Bearer <token>
cont-yn: N
next-key:
api-id: ka10075
```

`stock orders create cash`와 `stock orders cancel cash`는 현금 주문 endpoint를
호출합니다.

```http
POST https://api.kiwoom.com/api/dostk/ordr
Content-Type: application/json;charset=UTF-8
authorization: Bearer <token>
cont-yn: N
next-key:
api-id: kt10000 | kt10001 | kt10003
```

`stock orders create credit`과 `stock orders cancel credit`은 신용 주문
endpoint를 호출합니다.

```http
POST https://api.kiwoom.com/api/dostk/crdordr
Content-Type: application/json;charset=UTF-8
authorization: Bearer <token>
cont-yn: N
next-key:
api-id: kt10006 | kt10007 | kt10009
```

모든 종목과 모든 side 요청 body:

```json
{
  "all_stk_tp": "0",
  "trde_tp": "0",
  "stk_cd": "",
  "stex_tp": "0"
}
```

삼성전자 매수 주문 조회 body:

```json
{
  "all_stk_tp": "1",
  "trde_tp": "2",
  "stk_cd": "005930",
  "stex_tp": "0"
}
```

응답 header가 `cont-yn: Y`를 반환하면 명령은 반환된 `next-key`로 다음 page를
요청하고, 응답이 끝나거나 100 page safety limit에 도달할 때까지 모든 page의
row를 merge합니다. Public output에는 pagination metadata가 포함되지 않습니다.
반복되는 `next-key`는 같은 page를 무한 재생하지 않도록 upstream continuation
error로 처리합니다.

## 출력 계약

명령은 bare JSON array를 반환합니다. Item은 정규화된 `snake_case` 필드만
포함합니다.

| 필드 | Source | Notes |
| --- | --- | --- |
| `order_id` | `ord_no` | 선행 zero를 포함해 문자열로 보존합니다. |
| `original_order_id` | `orig_ord_no` | `0000000`을 포함해 문자열로 보존합니다. |
| `stock_code` | `stk_cd` | 종목 코드입니다. |
| `stock_name` | `stk_nm` | 종목명입니다. |
| `side` | `io_tp_nm` | `io_tp_nm`이 `매수`만 포함하면 `buy`, `매도`만 포함하면 `sell`, 그 외는 `unknown`입니다. 앞의 `+`/`-` marker는 side 분류에 사용하지 않습니다. |
| `trading_venue` | `stex_tp` | `stex_tp == "0"`이면 `SOR`, `stex_tp == "1"`이면 `KRX`, `stex_tp == "2"`이면 `NXT`입니다. 알 수 없는 non-empty 값은 `UNKNOWN_<raw>`로 보존하고 blank 값은 `UNKNOWN`이 됩니다. |
| `ordered_quantity` | `ord_qty` | 절댓값 숫자 수량으로 parse합니다. |
| `ordered_price` | `ord_pric` | 절댓값 숫자 가격으로 parse합니다. |
| `unfilled_quantity` | `oso_qty` | 절댓값 숫자 수량으로 parse합니다. |
| `funding_type` | `io_tp_nm` | `io_tp_nm`에 `신용`이 포함되면 `credit`, 그 외는 `cash`입니다. |
| `filled_quantity` | `cntr_qty` | 절댓값 숫자 수량으로 parse합니다. |
| `current_price` | `cur_prc` | 절댓값 숫자 가격으로 parse합니다. Kiwoom `+`/`-` prefix는 marker입니다. |

Raw `io_tp_nm`, raw `stex_tp`, 그 외 raw Kiwoom 응답 필드는 의도적으로 출력하지
않습니다. `trading_venue`는 향후 cancel 또는 amend 명령에 필요한 주문 venue/routing
값을 전달하는 안정적인 agent-facing 필드입니다. 알 수 없는 Kiwoom `stex_tp` 값은
전체 list를 실패시키는 대신 정규화된 필드에 보존하므로, agent가 나중 작업에
필요한 주문 번호를 유지하면서 unmapped venue를 분명히 볼 수 있습니다.

`side`와 `funding_type`은 안정적인 정규화 필드를 소비할 수 있도록 공개 schema의
일부입니다. 둘 다 Kiwoom `io_tp_nm` wording에서 파생됩니다. 예상하지 못했거나
모호한 side wording은 전체 list를 실패시키거나 buy/sell 방향을 추측하지 않고
`unknown`으로 출력합니다.

예시 shape:

```json
[
  {
    "order_id": "0000069",
    "original_order_id": "0000000",
    "stock_code": "005930",
    "stock_name": "삼성전자",
    "side": "buy",
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

## 검증 체크리스트

완료를 보고하기 전에:

```sh
gofmt -w <edited-go-files>
go test ./...
go vet ./...
go build -o bin/stock ./cmd/stock
```

선택적 live smoke 검증은 raw order row를 출력하지 않아야 합니다.

```sh
set -euo pipefail
umask 077
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
    "side",
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
assert all(row["side"] in {"buy", "sell", "unknown"} for row in data)
assert all(row["trading_venue"] in {"SOR", "KRX", "NXT", "UNKNOWN"} or row["trading_venue"].startswith("UNKNOWN_") for row in data)
print({"count": len(data), "schema": sorted(expected)})
PY
```

설정된 계좌에 미체결 주문이 없으면 `count`가 `0`일 수 있습니다. 이는 정상적인
live state이며 실패가 아닙니다.

기본적으로 live create/cancel smoke 검증은 실행하지 마세요. 해당 명령은 실제
주문을 넣거나 취소할 수 있습니다. Live-order smoke를 부작용까지 명확히 인지하고
명시적으로 요청받은 경우가 아니라면 unit test와 local validation-only CLI check를
선호하세요.
