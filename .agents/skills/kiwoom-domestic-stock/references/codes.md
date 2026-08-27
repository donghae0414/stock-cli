# Stock Code Lookup

Use `codes lookup` when a user provides a Korean stock name but a downstream command requires a
six-digit stock code. This is a private Kiwoom command: it requires configured credentials and may
issue or refresh an access token.

```bash
stock codes lookup --name 삼성전자
stock codes lookup --name 삼성전자 --name 하이닉스 --limit 10
```

Each query result has one of four statuses:

| Status | Meaning | Agent action |
| --- | --- | --- |
| `exact` | One exact name match | Continue only when there is exactly one candidate. |
| `single_partial` | One partial-name candidate | Show it and ask the user to confirm the intended stock. |
| `ambiguous` | Multiple candidates | Show the candidates and ask the user to choose. |
| `not_found` | No candidate | Stop and ask for another name or a six-digit code. |

Never infer a live-order stock code from `single_partial` or `ambiguous`. Stop on `ok: false` or a
non-empty `errors` array and summarize the typed error without exposing credentials or tokens.

Successful envelope:

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

Preserve leading zeros in `candidates[].code`. `--limit` limits candidates per query; it does not
change the exact-match safety rule. See `docs/codes-commands.md` for the authoritative schema and
error types when working inside the stock-cli repository.
