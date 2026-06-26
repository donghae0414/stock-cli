# Stock CLI Market 명령

`stock market`은 agent workflow에서 쓰는 로컬 시장 규칙 helper를 담습니다.
이 명령들은 Kiwoom을 호출하지 않고, credential을 로드하지 않으며, token 캐시를
건드리지 않습니다.

## 명령

```sh
./bin/stock market tick --price 353333
```

옵션:

| 옵션 | 값 | 목적 |
| --- | --- | --- |
| `--price` | 양의 정수 원화 가격 | 일반 국내주식 호가 단위 표에 맞춰 확인할 가격입니다. |

## 출력 계약

명령은 정규화된 JSON object 하나를 반환합니다.

```json
{
  "price": 353333,
  "tick_size": 500,
  "lower_price": 353000,
  "upper_price": 353500,
  "is_valid_tick": false
}
```

필드:

| 필드 | 설명 |
| --- | --- |
| `price` | 입력 가격을 JSON number로 출력합니다. |
| `tick_size` | 입력 가격 구간에 선택된 호가 단위입니다. |
| `lower_price` | `price` 이하에서 가장 큰 유효 호가입니다. |
| `upper_price` | `price` 이상에서 가장 작은 유효 호가입니다. |
| `is_valid_tick` | `price`가 이미 호가 grid에 맞으면 `true`입니다. |

## 일반 주식 호가 단위 표

| 가격 구간 | 호가 단위 |
| --- | ---: |
| 2,000 미만 | 1 |
| 2,000 이상 4,999 이하 | 5 |
| 5,000 이상 19,999 이하 | 10 |
| 20,000 이상 49,999 이하 | 50 |
| 50,000 이상 199,999 이하 | 100 |
| 200,000 이상 499,999 이하 | 500 |
| 500,000 이상 | 1,000 |

## 1차 범위

이 helper는 의도적으로 오프라인입니다. Kiwoom `ka10004`를 호출하지 않고, 현재
호가창을 가져오지 않으며, `stock orders create` validation을 변경하지 않습니다.
ETF, ETN, ELW, 우선주, 상품 유형, 거래소별 예외도 이 1차 범위에서는 제외합니다.

## 검증 체크리스트

완료를 보고하기 전에:

```sh
gofmt -w pkg/cmd/market.go pkg/cmd/market_test.go pkg/cmd/cmd.go
go test ./...
go build -o bin/stock ./cmd/stock
./bin/stock market --help
./bin/stock market tick --price 353333
./bin/stock market tick --price 2001
```
