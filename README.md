# stock-cli

Kiwoom REST API를 사용하기 위한 Go 기반 주식 CLI입니다. 현재 구현된 범위는
Kiwoom 기본 설정, 계좌 보유 종목 조회, 주문 리스트 조회, 현금/신용 주문
생성 및 취소 명령입니다.

## 준비

프로젝트 디렉토리에서 터미널을 열고 실행합니다.

```sh
cd /Users/dongwuk/apps/stock-cli
```

이 환경에서는 Go가 `/usr/local/go/bin/go`에 있습니다. `go` 명령이 바로 동작하지
않으면 아래처럼 `PATH`를 먼저 잡습니다.

```sh
export PATH="/usr/local/go/bin:$PATH"
go version
```

## 빌드와 테스트

```sh
go test ./...
go vet ./...
go build -o bin/stock ./cmd/stock
```

`bin/`은 `.gitignore`에 포함되어 있습니다.

## 설정 확인

현재 `.env`는 로컬 테스트용 Kiwoom 키를 담고 있으며 파일 권한은 `0600`이어야
합니다. 값 자체를 출력하지 말고 권한만 확인합니다.

```sh
stat -f "%Lp %N" .env
```

`.env`를 현재 터미널 세션에 로드합니다.

```sh
set -a
. ./.env
set +a
```

설정 명령을 호출합니다.

```sh
./bin/stock --help
./bin/stock config --help
./bin/stock config path
./bin/stock config show
./bin/stock accounts --help
./bin/stock accounts list
```

`config show`는 키를 마스킹해서 보여줍니다.

## 직접 설정 저장

실제 사용자 홈에 Kiwoom 설정을 저장하려면 아래 명령을 실행합니다.

```sh
./bin/stock config set
```

프롬프트 순서:

1. App Key
2. Secret Key

Kiwoom REST API host는 `https://api.kiwoom.com`으로 고정되어 있으며 설정
파일에 저장하지 않습니다. 설정 파일은 `~/.stock/config`에 저장되고,
디렉토리는 `0700`, 파일은 `0600` 권한으로 보정됩니다.

## 안전한 임시 HOME 테스트

실제 `~/.stock/config`를 건드리지 않고 `config set/show/path`를 테스트하려면
임시 HOME을 사용합니다.

```sh
QA_HOME="$(mktemp -d)"
HOME="$QA_HOME" ./bin/stock config path
HOME="$QA_HOME" ./bin/stock config show
HOME="$QA_HOME" ./bin/stock config set
HOME="$QA_HOME" ./bin/stock config show
rm -rf "$QA_HOME"
```

`config set`은 TTY 전용 명령이라 사람이 직접 입력해야 합니다. 자동화된 검증은
테스트 코드와 UltraQA harness에서 임시 HOME과 fake 키로 수행합니다.

## 계좌 보유 종목 조회

`stock accounts list`는 Kiwoom `ka10085` 계좌수익률요청 API를 호출해서 현재
보유 중인 종목을 JSON 배열로 출력합니다.

```sh
set -a
. ./.env
set +a
./bin/stock accounts list
```

출력은 Kiwoom 원본 필드명이 아니라 agent가 읽기 쉬운 snake_case 필드만
포함합니다.

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

보유수량이 `0`인 행과 종목명이 `*`로 시작하는 Kiwoom 상세 행은 제외하고,
별표 없는 종목별 합산 행만 반환합니다. `profit_rate`는 현재가와 매입가로
계산해서 소수점 둘째 자리까지 반올림해 출력하며, 매입가가 `0`이면 `null`입니다.

신용 매수 상세를 대출일별로 확인하려면 `--credit-detail`을 사용합니다.

```sh
./bin/stock accounts list --credit-detail
```

이 모드에서는 현금 종목은 계속 표시하고, 종목명이 `*`로 시작하면서
`crd_tp`가 `00`이 아닌 Kiwoom 신용 상세 행을 개별 표시합니다. 별표 없는
신용 합산 행은 중복 집계를 피하기 위해 제외합니다. 출력 이름에서는 leading
`*`를 제거하고, `loan_dt`는 `loan_date`로 출력합니다.

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

`is_credit` 필드는 제거되었고 `funding_type`을 사용합니다. `funding_type` 값은 `cash` 또는 `credit`이며, 기존 `is_credit == true` 소비자는 `funding_type == "credit"`로 마이그레이션해야 합니다. 주문 command의 현금/신용 입력은 `stock orders create cash` / `stock orders create credit` subcommand로 구분하고, `order_type`은 MARKET/LIMIT 주문 방식에만 사용합니다.

`--credit-detail` 출력에는 기존 `funding_type` 필드에 `loan_date`만 추가됩니다. 현금 행의
`loan_date`는 빈 문자열입니다.

## 주문 리스트 조회

`stock orders list`는 Kiwoom `ka10075` 주문 리스트 API를 호출해서 현재
미체결/주문 rows를 JSON 배열로 출력합니다. 연속조회 응답이 있으면 100 page
safety limit 안에서 모든 page를 자동으로 조회해서 합친 뒤 출력합니다.

```sh
./bin/stock orders list
./bin/stock orders list --side buy --stock-code 005930
```

출력은 agent가 읽기 쉬운 snake_case 필드만 포함하며, 주문 취소에 재사용될 수
있는 주문 식별자는 문자열 그대로 보존합니다.

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

`--side`는 `all`, `buy`, `sell`을 지원하고 기본값은 `all`입니다.
`--stock-code`는 6자리 종목 코드이며 없으면 전체 종목을 조회합니다.
`trading_venue`는 Kiwoom 응답 `stex_tp`를 정규화한 값이며, `0`은 `SOR`,
`1`은 `KRX`, `2`는 `NXT`로 출력합니다. 이 값은 향후 주문 취소/정정에서
주문된 거래 경로를 판단하는 데 사용될 수 있습니다. 아직 매핑하지 않은 값은
주문 식별자를 잃지 않도록 `UNKNOWN_<raw>` 형식으로 보존하고, 빈 값은
`UNKNOWN`으로 출력합니다.
`funding_type`은 `io_tp_nm`에 `신용`이 포함되면 `credit`, 아니면 `cash`로
정규화합니다. 현재 계정의 live API 응답에 주문 row가 없어 이 판정은 문서화된
임시 heuristic이며, 실제 주문 row가 확인되면 관측된 `io_tp_nm` 값으로 회귀
fixture를 추가해 재검증해야 합니다. JSON field 이름은 유지하되, 그 전까지
cash/credit 분류 의미는 best-effort 값으로 취급합니다.

자세한 API mapping과 안전한 smoke 검증은 `docs/order-commands.md`를 참고합니다.

## 시장 규칙 helper

`stock market tick`은 한국 주식 일반 호가 단위에 맞춰 입력 가격의 아래/위
유효 호가를 계산하는 로컬 helper입니다. Kiwoom API를 호출하지 않고 credential
또는 token cache도 읽지 않습니다.

```sh
./bin/stock market tick --price 353333
```

```json
{
  "price": 353333,
  "tick_size": 500,
  "lower_price": 353000,
  "upper_price": 353500,
  "is_valid_tick": false
}
```

첫 구현 범위는 일반 주식 호가 단위 계산뿐입니다. `ka10004` orderbook 조회,
`orders create` 가격 검증 변경, ETF/ETN/ELW 등 상품별 예외 처리는 포함하지
않습니다.

자세한 command contract는 `docs/market-commands.md`를 참고합니다.

## 주문 생성/취소

`stock orders create cash`, `stock orders create credit`,
`stock orders cancel cash`, `stock orders cancel credit`는 Kiwoom 주문 API를
즉시 호출합니다. CLI 자체 confirmation, dry-run, quote 확인, 잔고/담보/보유수량
확인은 하지 않고, 이런 안전 정책은 CLI를 호출하는 Agent workflow가 담당합니다.

신용 매도에서는 `--loan-selection individual|aggregate`를 사용합니다.
`individual`은 `--loan-date YYYYMMDD`가 필수이고, `aggregate`와 함께
`--loan-date`를 넘기면 validation error입니다.

자세한 command contract와 API mapping은 `docs/order-commands.md`를 참고합니다.

## Kiwoom 토큰 발급 스펙

직접 호출로 확인한 토큰 발급 스펙은 다음과 같습니다.

```http
POST https://api.kiwoom.com/oauth2/token
Content-Type: application/json;charset=UTF-8
```

요청 본문:

```json
{
  "grant_type": "client_credentials",
  "appkey": "...",
  "secretkey": "..."
}
```

응답 주요 필드:

```json
{
  "expires_dt": "YYYYMMDDHHMMSS",
  "token_type": "Bearer",
  "token": "...",
  "return_code": 0,
  "return_msg": "정상적으로 처리되었습니다"
}
```

성공 응답에서 `next-key`, `cont-yn`, `api-id` 헤더는 관측되지 않았습니다.

## Kiwoom 토큰 캐시

`stock accounts list`는 실행 시 토큰이 없거나 만료 1분 이내이면 새 토큰을
발급받고, 발급 토큰을 `~/.stock/token.json`에 저장합니다. `~/.stock/config`에는
장기 보관용 `appkey`와 `secretkey`만 저장하고 발급 토큰은 저장하지 않습니다.

`stock config set`이 새 credential 저장에 성공하면 기존 `~/.stock/token.json`을
삭제합니다. 따라서 다음 token-using command는 활성 credential source로 토큰을
다시 발급받습니다. 환경 변수가 없으면 방금 저장한 config credential을 사용하고,
`KIWOOM_APPKEY`/`KIWOOM_SECRETKEY`가 있으면 기존 resolution order대로 환경
변수가 우선합니다. 이미 token cache가 없으면 정상 no-op로 처리됩니다. 이 동작은
성공한 `stock config set`에만 적용되며, 환경 변수 변경은 token cache를 자동으로
삭제하지 않습니다.

토큰 캐시 파일은 `0600`, 상위 `~/.stock` 디렉토리는 `0700` 권한으로 보정합니다.
토큰 값은 CLI 출력, 로그, 문서, git diff에 노출하지 않아야 합니다.
