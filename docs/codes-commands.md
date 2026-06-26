# Stock Code 조회 명령

`stock codes lookup`은 한국 종목명을 Kiwoom 6자리 종목 코드로 해석합니다.
Kiwoom 종목 정보 API를 사용합니다.

```sh
./bin/stock codes lookup --name 삼성전자
./bin/stock codes lookup --name 삼성전자 --name 하이닉스 --limit 10
```

이 명령은 Kiwoom `/api/dostk/stkinfo`를 api-id `ka10099`로 호출하고,
시장 구분 `0`과 `10`을 조회하며, continuation header를 따라간 뒤 agent가 읽기
좋은 JSON envelope를 반환합니다.

성공 출력은 항상 object입니다.

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

`status`는 `exact`, `single_partial`, `ambiguous`, `not_found` 중 하나입니다.
여러 후보가 매칭되면 agent는 종목 코드만으로 진행하는 작업을 멈추고, 어떤
후보가 의도한 종목인지 확인해야 합니다.

명령은 이름에 `ETF`, `ETN`, `선물`, `옵션`이 들어간 항목을 제외하고, `거래소`,
`코스닥`, `리츠`, `인프라투자금융` 시장을 유지합니다. 후보 ranking은 정확 일치,
부분 일치 순으로 적용하며, 부분 일치 안에서는 prefix match, 이름 길이, 코드
길이, 코드 순서로 정렬합니다.

Validation, config, Kiwoom upstream 오류도 JSON으로 출력합니다.

```json
{
  "ok": false,
  "queries": [],
  "errors": [
    {
      "type": "ValidationError",
      "message": "at least one --name is required"
    }
  ]
}
```

`errors[].type`은 다음 중 하나입니다.

- `ValidationError`: `--name`이 없거나 blank인 경우, `--limit`이 유효하지 않은 경우, 예상하지 않은 추가 인자가 있는 경우.
- `ConfigError`: Kiwoom credential/config가 없거나 읽을 수 없는 경우. `stock config set`을 실행하세요.
- `KiwoomClientError`: Kiwoom token, network, HTTP, response decoding, API return-code 실패.
