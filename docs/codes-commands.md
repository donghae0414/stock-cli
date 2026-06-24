# Stock Code Lookup Commands

`stock codes lookup` resolves Korean stock names to six-digit Kiwoom stock codes through the Kiwoom stock-info API.

```sh
./bin/stock codes lookup --name 삼성전자
./bin/stock codes lookup --name 삼성전자 --name 하이닉스 --limit 10
```

The command calls Kiwoom `/api/dostk/stkinfo` with api-id `ka10099`, fetches market types `0` and `10`, follows continuation headers, and returns an agent-readable JSON envelope.

Success output is always an object:

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

`status` is one of `exact`, `single_partial`, `ambiguous`, or `not_found`. If several candidates match, agents should stop before stock-code-only work and ask which candidate is intended.

The command filters out names containing `ETF`, `ETN`, `선물`, or `옵션`, and keeps `거래소`, `코스닥`, `리츠`, and `인프라투자금융` markets. Candidate ranking follows exact matches first, then partial matches ordered by prefix match, name length, code length, and code.

Validation, config, or Kiwoom upstream errors are also emitted as JSON:

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

`errors[].type` is one of:

- `ValidationError`: missing or blank `--name`, invalid `--limit`, or unexpected extra arguments.
- `ConfigError`: missing or unreadable Kiwoom credentials/config. Run `stock config set`.
- `KiwoomClientError`: Kiwoom token, network, HTTP, response decoding, or API return-code failure.
