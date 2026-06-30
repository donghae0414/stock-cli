# stock-cli

Kiwoom REST API를 사용하기 위한 Go 기반 국내 주식 CLI입니다.  
이 프로젝트는 [`upbit-official/upbit-cli`](https://github.com/upbit-official/upbit-cli)의  
resource-based CLI 흐름을 참고해 Kiwoom 증권 API용으로 만든 도구입니다.

현재 지원 범위는 설정, 계좌 보유 종목 조회, 종목코드 조회, 시장 호가 단위 계산, 차트 조회, 주문 조회, 현금/신용 주문 생성 및 취소입니다.

## Installation

### npm package

npm 배포 후에는 package manager로 설치할 수 있습니다.

```sh
npm install -g @donghae0414/stock-cli
stock --help
```

Node.js 18 이상이 필요합니다. npm package는 플랫폼별 `stock` binary를 포함하고,
설치 후 `stock` 명령을 제공합니다.

### Go source build

소스에서 바로 빌드하려면 Go toolchain이 필요합니다.

```sh
git clone https://github.com/donghae0414/stock-cli.git
cd stock-cli
go test ./...
go build -o bin/stock ./cmd/stock
./bin/stock --help
```

아래 예시는 설치된 binary를 `stock`으로 표기합니다. 소스 빌드만 사용한다면
`stock` 대신 `./bin/stock`을 사용하면 됩니다.

## Quick Start

1. Kiwoom REST API credential을 저장합니다.

```sh
stock config set
```

2. 설정 저장 위치와 마스킹된 설정을 확인합니다.

```sh
stock config path
stock config show
```

3. read-only 명령으로 동작을 확인합니다.

```sh
stock market tick --price 353333
stock codes lookup --name 삼성전자
stock accounts list
```

`stock accounts list`, `stock codes lookup`, `stock chart ...`, `stock orders ...`
명령은 Kiwoom credential이 필요합니다. `stock market tick`은 로컬 계산만 수행하며
credential이나 token cache를 읽지 않습니다.

## Credential and Token Safety

`stock config set`은 Kiwoom App Key와 Secret Key를 `~/.stock/config`에 저장합니다.
설정 디렉토리는 `0700`, 설정 파일은 `0600` 권한으로 보정됩니다.

발급된 Kiwoom access token은 `~/.stock/token.json`에 별도로 cache됩니다.
`~/.stock/config`에는 장기 credential만 저장하며 token 값은 저장하지 않습니다.
문서, 로그, commit, issue에 App Key, Secret Key, access token을 넣지 마세요.

자세한 token 발급/캐시 정책은 [docs/config-commands.md](docs/config-commands.md)와
[docs/account-commands.md](docs/account-commands.md)를 참고하세요.

## Command Shape

```sh
stock [resource] <command> [options]
```

Global options:

| Option | Description |
| --- | --- |
| `--help`, `-h` | Show help. |
| `--version`, `-v` | Print the CLI version. |

Top-level command groups:

| Group | Purpose |
| --- | --- |
| `config` | Manage Kiwoom credential configuration. |
| `accounts` | Read normalized Kiwoom account holdings. |
| `codes` | Resolve Korean stock names to six-digit stock codes. |
| `market` | Run local market-rule helpers. |
| `chart` | Read normalized Kiwoom chart candles. |
| `orders` | Read, create, and cancel Kiwoom orders. |

## Config Commands

`stock config` stores and displays Kiwoom credential configuration.

### Commands

| Command | Description |
| --- | --- |
| `stock config set` | Interactively save Kiwoom App Key and Secret Key. |
| `stock config show` | Show masked credential values and their source. |
| `stock config path` | Print the config file path. |

### Options

| Command | Options |
| --- | --- |
| `stock config set` | none |
| `stock config show` | none |
| `stock config path` | none |

### Example input

```sh
stock config path
stock config show
```

### Example output

```text
/Users/alice/.stock/config
```

```text
appkey:      abcd****wxyz  (source: /Users/alice/.stock/config)
secretkey:   ****          (source: /Users/alice/.stock/config)
config_file: ~/.stock/config
```

Detailed config behavior is documented in [docs/config-commands.md](docs/config-commands.md).

## Account Commands

`stock accounts list` calls Kiwoom account APIs and prints normalized holdings as
a JSON array.

### Commands

| Command | Description |
| --- | --- |
| `stock accounts list` | Return current holdings, excluding zero-quantity and Kiwoom detail rows. |
| `stock accounts list --credit-detail` | Return cash rows plus credit loan-date detail rows. |

### Options

| Option | Description |
| --- | --- |
| `--credit-detail` | Show credit holdings by loan date. Defaults to `false`. |

### Example input

```sh
stock accounts list
stock accounts list --credit-detail
```

### Example output

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

Credit detail output adds `loan_date`:

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
    "funding_type": "credit",
    "loan_date": "20260601"
  }
]
```

Detailed account API mapping and filtering rules are documented in
[docs/account-commands.md](docs/account-commands.md).

## Code Lookup Commands

`stock codes lookup` resolves Korean stock names to Kiwoom six-digit stock
codes.

### Commands

| Command | Description |
| --- | --- |
| `stock codes lookup` | Return exact, partial, ambiguous, or not-found candidates for one or more names. |

### Options

| Option | Description |
| --- | --- |
| `--name string` | Stock name query. Repeat for multiple names. Required at least once. |
| `--limit string` | Maximum candidates per query. Defaults to `10`. |

### Example input

```sh
stock codes lookup --name 삼성전자
stock codes lookup --name 삼성전자 --name 하이닉스 --limit 10
```

### Example output

```json
{
  "ok": true,
  "queries": [
    {
      "query": "삼성전자",
      "status": "exact",
      "match_type": "exact",
      "candidates": [
        {
          "code": "005930",
          "name": "삼성전자",
          "market_name": "거래소",
          "match_type": "exact",
          "up_name": "전기전자"
        }
      ],
      "total_candidates": 1,
      "truncated": false
    }
  ],
  "errors": []
}
```

If `status` is `ambiguous`, choose a candidate before running stock-code-only
commands such as `chart` or `orders`.

Detailed lookup ranking and error output are documented in
[docs/codes-commands.md](docs/codes-commands.md).

## Market Helper Commands

`stock market` contains local helper commands. These commands do not call
Kiwoom, do not read credentials, and do not touch token cache.

### Commands

| Command | Description |
| --- | --- |
| `stock market tick` | Calculate the Korean stock tick size and nearest valid tick prices. |

### Options

| Option | Description |
| --- | --- |
| `--price string` | Positive whole-won stock price. Required. |

### Example input

```sh
stock market tick --price 353333
```

### Example output

```json
{
  "price": 353333,
  "tick_size": 500,
  "lower_price": 353000,
  "upper_price": 353500,
  "is_valid_tick": false
}
```

Detailed tick-size boundaries are documented in
[docs/market-commands.md](docs/market-commands.md).

## Chart Commands

`stock chart` returns normalized Kiwoom candle data.

### Commands

| Command | Description |
| --- | --- |
| `stock chart day` | Return daily candles. |
| `stock chart week` | Return weekly candles. |
| `stock chart minute` | Return minute candles. |

### Options

| Option | Commands | Description |
| --- | --- | --- |
| `--stock-code string` | day, week, minute | Six-digit stock code. Required. |
| `--count int` | day, week, minute | Number of candles, `1..600`. Defaults to `120`. |
| `--to string` | day, week | End date in `YYYYMMDD`. Defaults to current local date. |
| `--interval int` | minute | Minute interval: `1`, `3`, `5`, `10`, `15`, `30`, `45`, or `60`. Required. |
| `--to string` | minute | End time in `YYYYMMDD` or `YYYYMMDDHHmmss`. Defaults to current local date. |

### Example input

```sh
stock chart day --stock-code 005930 --count 2
stock chart week --stock-code 005930 --count 2 --to 20260618
stock chart minute --stock-code 005930 --interval 1 --count 2 --to 20260618132000
```

### Example output

Day/week:

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

Minute:

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

Detailed chart API mapping and normalization rules are documented in
[docs/chart-commands.md](docs/chart-commands.md).

## Order Commands

`stock orders` reads open orders and can submit live Kiwoom order requests.

Important: `stock orders create ...` places real Kiwoom orders immediately, and
`stock orders cancel ...` sends real cancellation requests immediately. The CLI
does not perform confirmation, dry-run, quote checks, cash checks, collateral
checks, loan checks, holdings checks, or price-limit checks. Those safety checks
belong in the workflow or human process that calls this CLI.

### Commands

| Command | Description |
| --- | --- |
| `stock orders list` | Return current open/unfilled order rows. |
| `stock orders create cash` | Submit a cash buy or sell order. |
| `stock orders create credit` | Submit a credit buy or sell order. |
| `stock orders cancel cash` | Cancel a cash order. |
| `stock orders cancel credit` | Cancel a credit order. |

### Options

`stock orders list`:

| Option | Description |
| --- | --- |
| `--side string` | `all`, `buy`, or `sell`. Defaults to `all`. |
| `--stock-code string` | Optional six-digit stock code filter. |

`stock orders create cash`:

| Option | Description |
| --- | --- |
| `--side string` | `buy` or `sell`. Required. |
| `--stock-code string` | Six-digit stock code. Required. |
| `--order-type string` | `limit` or `market`. Required. |
| `--quantity string` | Positive whole-share quantity. Required. |
| `--price string` | Per-share limit price. Required for `limit`; invalid for `market`. |
| `--trading-venue string` | `SOR`, `KRX`, or `NXT`. Defaults to `SOR`. |

`stock orders create credit` adds:

| Option | Description |
| --- | --- |
| `--loan-selection string` | Credit sell only: `individual` or `aggregate`. |
| `--loan-date string` | Required for `--loan-selection individual`; invalid with `aggregate` and credit buy. |

`stock orders cancel cash` and `stock orders cancel credit`:

| Option | Description |
| --- | --- |
| `--stock-code string` | Six-digit stock code. Required. |
| `--original-order-id string` | Original order identifier. Required and preserved as a string. |
| `--quantity string` | Positive cancel quantity. Omit to cancel all remaining quantity. |
| `--trading-venue string` | `SOR`, `KRX`, or `NXT`. Defaults to `SOR`. |

### Example input

Read-only order list:

```sh
stock orders list
stock orders list --side buy --stock-code 005930
```

Live order examples:

```sh
stock orders create cash --side buy --stock-code 005930 --order-type limit --quantity 1 --price 74100
stock orders create credit --side sell --stock-code 005930 --order-type limit --quantity 3 --price 6450 --loan-selection aggregate
stock orders cancel cash --stock-code 005930 --original-order-id 0000140
stock orders cancel credit --stock-code 005930 --original-order-id 0001615
```

### Example output

Order list:

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

Create order:

```json
{
  "order_id": "0000024",
  "trading_venue": "SOR"
}
```

Cancel order:

```json
{
  "order_id": "0000141",
  "base_original_order_id": "0000140",
  "cancelled_quantity": 1
}
```

Detailed order API mapping, validation rules, and live smoke cautions are
documented in [docs/order-commands.md](docs/order-commands.md).
