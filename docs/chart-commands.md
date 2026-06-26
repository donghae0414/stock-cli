# Stock CLI Chart 명령

`stock chart day`, `stock chart week`, `stock chart minute`은 agent가 소비할 수
있는 Kiwoom 차트 candle을 반환합니다. Kiwoom 차트 API를 호출하고, 원본 Kiwoom
필드를 영어 `snake_case` JSON으로 정규화합니다.

이 명령들은 읽기 전용입니다. 지표를 계산하지 않고, CSV를 export하지 않으며,
종목명 조회, 여러 종목 조회, 원본 Kiwoom 필드 노출, 첫 Kiwoom 응답 이후의
continuation page 조회를 하지 않습니다.

## 명령

```sh
./bin/stock chart day --stock-code 005930
./bin/stock chart day --stock-code 005930 --count 10 --to 20260618
./bin/stock chart week --stock-code 005930 --count 10 --to 20260618
./bin/stock chart minute --stock-code 005930 --interval 1 --count 10
./bin/stock chart minute --stock-code 005930 --interval 1 --count 10 --to 20260618132000
```

## 옵션

공통 옵션:

| 옵션 | 값 | 목적 |
| --- | --- | --- |
| `--stock-code` | 6자리 숫자 | 필수 종목 코드입니다. |
| `--count` | `1..600` | 반환할 candle 수입니다. 기본값은 `120`입니다. |
| `--to` | 아래 설명 참고 | 종료 날짜 또는 시각입니다. 기본값은 현재 local date입니다. |

Day/week `--to`는 `YYYYMMDD`만 허용합니다.

Minute `--interval`은 필수이며 `1`, `3`, `5`, `10`, `15`, `30`, `45`, `60`을
허용합니다.

Minute `--to`는 `YYYYMMDD` 또는 `YYYYMMDDHHmmss`를 허용합니다. Kiwoom에는
date 부분만 `base_dt`로 전달하고, 사용자가 제공한 전체 값은 출력 metadata에
보존합니다.

## API Mapping

모든 chart 요청은 다음 값을 사용합니다.

- Endpoint: `/api/dostk/chart`
- `upd_stkpc_tp: "1"`
- `stk_cd`: 사용자 종목 코드 뒤에 SOR routing을 위한 `_AL`을 붙인 값
- `cont-yn: N`
- `next-key: ""`

| 명령 | API ID | 응답 list |
| --- | --- | --- |
| `stock chart day` | `ka10081` | `stk_dt_pole_chart_qry` |
| `stock chart week` | `ka10082` | `stk_stk_pole_chart_qry` |
| `stock chart minute` | `ka10080` | `stk_min_pole_chart_qry` |

## 출력

공개 출력은 JSON object 하나입니다. `stock_code`는 top level에 한 번만 나타나고,
`candles` 아래 각 row는 정규화된 candle입니다.

Day/week 출력:

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

Minute 출력:

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

Kiwoom은 숫자 문자열 앞에 `+` 또는 `-` sign을 포함할 수 있습니다. 공개 가격과
거래대금 필드는 절댓값 JSON number입니다. 날짜와 timestamp는 문자열로 유지합니다.

## 안전한 Smoke 검증

빌드 후 `stock config set`으로 credential을 설정한 뒤 작은 count로 실행합니다.

```sh
./bin/stock chart day --stock-code 005930 --count 2
./bin/stock chart week --stock-code 005930 --count 2
./bin/stock chart minute --stock-code 005930 --interval 1 --count 2
```

Token이나 credential을 출력하지 마세요. 자동화에서는 exit status, top-level key,
candle count, 첫 candle key처럼 sanitization된 증거만 기록하세요.
