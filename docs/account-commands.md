# Stock CLI Account 명령

`stock accounts list`는 Kiwoom REST API를 사용하고, agent가 소비할 수 있도록
정규화된 계좌 보유 종목을 반환합니다. 이 명령은 resource 기반 경로를 유지하고
예상하지 않은 추가 인자를 거부하므로 호출자는 결정적인 동작을 얻습니다.

## Stock CLI 명령

`stock-cli`는 다음을 제공합니다.

| Stock 명령 | 목적 |
| --- | --- |
| `stock accounts list` | 현재 Kiwoom 계좌 보유 종목을 정규화해서 반환합니다. |
| `stock accounts list --credit-detail` | 현금 보유 row와 대출일자별 신용 detail row를 함께 반환합니다. |

명령별 옵션:

| 옵션 | 목적 |
| --- | --- |
| `--credit-detail` | 현금 row를 유지하면서 Kiwoom 신용 detail row를 대출일자별로 표시합니다. |

## Kiwoom Token 정책

장기 Kiwoom credential은 다음 위치에 둡니다.

```text
~/.stock/config
```

발급된 access token은 별도로 캐시합니다.

```text
~/.stock/token.json
```

Token 캐시 규칙:

- 발급된 access token을 `~/.stock/config`에 저장하지 않습니다.
- 지원되는 환경에서는 `~/.stock/`를 `0700` 권한으로 만듭니다.
- 지원되는 환경에서는 `~/.stock/token.json`을 `0600` 권한으로 씁니다.
- token 캐시가 없거나, malformed 상태이거나, 읽을 수 없거나, 만료되었거나, 곧 만료되면 cache miss로 취급합니다.
- 캐시된 token이 1분 안에 만료되면 갱신합니다.
- token 발급 응답이 성공한 뒤에만 캐시를 교체합니다.
- 성공한 `stock config set` 이후 `~/.stock/token.json`을 제거하여 다음 account 명령이 새로 저장된 config credential로 fresh token을 발급하게 합니다.
- 해당 제거는 성공한 `stock config set`에만 한정합니다.
- 발급된 token 값을 출력하거나 commit하지 않습니다.

## Kiwoom API Mapping

### Token 발급

```http
POST https://api.kiwoom.com/oauth2/token
Content-Type: application/json;charset=UTF-8
```

요청 body:

```json
{
  "grant_type": "client_credentials",
  "appkey": "...",
  "secretkey": "..."
}
```

사용하는 응답 필드:

| 필드 | 의미 |
| --- | --- |
| `token` | 발급된 access token입니다. |
| `token_type` | 기대하는 bearer token type입니다. |
| `expires_dt` | `YYYYMMDDHHMMSS` 형식의 token 만료 시각입니다. |
| `return_code` | Kiwoom business result code입니다. |
| `return_msg` | Kiwoom business result message입니다. |

### 계좌 보유 종목

`stock accounts list`는 Kiwoom 계좌 수익률 요청 `ka10085`를 호출합니다.

```http
POST https://api.kiwoom.com/api/dostk/acnt
Content-Type: application/json;charset=UTF-8
authorization: Bearer <token>
cont-yn: N
next-key:
api-id: ka10085
```

요청 body:

```json
{
  "stex_tp": "0"
}
```

## 출력 계약

명령은 bare JSON array를 반환합니다.

기본 `stock accounts list` item은 정규화된 `snake_case` 필드만 포함합니다.

| 필드 | Source | Notes |
| --- | --- | --- |
| `stock_code` | `stk_cd` | 종목 코드입니다. |
| `stock_name` | `stk_nm` | 종목명입니다. |
| `current_price` | `cur_prc` | 절댓값 숫자 가격으로 parse합니다. Kiwoom `+`/`-` prefix는 marker로 취급합니다. |
| `purchase_price` | `pur_pric` | 숫자 가격으로 parse합니다. |
| `profit_rate` | computed | `(current_price - purchase_price) / purchase_price * 100`을 소수점 둘째 자리로 반올림합니다. 매입가가 `0`이면 `null`입니다. |
| `purchase_amount` | `pur_amt` | 숫자 매입금액으로 parse합니다. |
| `holding_quantity` | `rmnd_qty` | 숫자 보유수량으로 parse합니다. |
| `orderable_quantity` | `clrn_alow_qty` | 숫자 주문가능수량으로 parse합니다. |
| `funding_type` | `crd_tp` | `crd_tp == "00"`이면 `cash`, 그 외는 `credit`입니다. |

Filtering 규칙:

- `rmnd_qty`가 `0`인 row는 제외합니다.
- `stk_nm`이 `*`로 시작하는 row는 제외합니다.
- 별표가 없는 aggregate row만 반환합니다.

예시 shape:

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
    "funding_type": "cash"
  }
]
```

Breaking change: account JSON은 더 이상 `is_credit`을 내보내지 않습니다. 대신
`funding_type == "credit"`을 사용하세요. 유효한 값은 정확히 `cash`와
`credit`입니다. 주문 명령은 `stock orders create cash`와
`stock orders create credit` subcommand로 이 cash/credit 축을 구분합니다.
`order_type`은 MARKET/LIMIT 주문 방식에 예약되어 있습니다. Live order 명령
계약은 order command 문서를 확인하세요.

### Credit detail 출력

`stock accounts list --credit-detail`은 기본 필드에 필드 하나를 추가해 반환합니다.

| 필드 | Source | Notes |
| --- | --- | --- |
| `loan_date` | `loan_dt` | 신용 detail row의 Kiwoom 대출일자입니다. 현금 row에서는 빈 문자열입니다. |

Filtering 규칙:

- `rmnd_qty`가 `0`인 row는 제외합니다.
- `crd_tp == "00"`인 모든 현금 row를 포함하고 `loan_date`를 빈 문자열로 설정합니다.
- `stk_nm`이 `*`로 시작하고 `crd_tp != "00"`인 신용 detail row를 포함합니다.
- 표시되는 신용 detail `stock_name`에서는 앞의 `*`를 제거합니다.
- 별표가 없고 `crd_tp != "00"`인 신용 aggregate row는 제외합니다.

예시 shape:

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
    "funding_type": "credit",
    "loan_date": "20260601"
  }
]
```

## 검증 체크리스트

완료를 보고하기 전에:

```sh
gofmt -w <edited-go-files>
go test ./...
go build -o bin/stock ./cmd/stock
```

최종 직접 CLI smoke 검증은 raw holdings를 disk에 남기지 않아야 합니다.

```sh
set -euo pipefail
umask 077
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
    "funding_type",
}

data = json.load(open(os.environ["out"]))
assert isinstance(data, list)
assert all(set(row) == expected for row in data)
assert all(not row["stock_name"].startswith("*") for row in data)
assert all(row["holding_quantity"] != 0 for row in data)
print({"count": len(data), "schema": sorted(expected)})
PY
```

Raw portfolio output을 명시적으로 요청받지 않았다면 요약된 `count`와 `schema`만
보고하세요.

Credit-detail smoke 검증도 raw holdings를 출력하지 않아야 합니다.

```sh
set -euo pipefail
umask 077
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
    "funding_type",
}
detail_expected = default_expected | {"loan_date"}

default_data = json.load(open(os.environ["default_out"]))
detail_data = json.load(open(os.environ["detail_out"]))
credit_detail_rows = [
    row for row in detail_data
    if row["funding_type"] == "credit" and row["loan_date"] != ""
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

설정된 계좌에 현재 신용 row가 없으면 최종 `credit_detail_rows_with_loan_date`
count가 `0`일 수 있습니다. Raw `ka10085` 응답에 `crd_tp != "00"` row가 없음을
확인했다면 이는 live-state inconclusive로 취급하세요.
