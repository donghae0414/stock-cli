# Stock CLI 초기 설정 명령

`stock config`는 Kiwoom 전용 credential을 처음 설정하는 명령 표면입니다.
Kiwoom REST API host는 고정값이며 사용자가 설정하지 않습니다.

## Stock CLI 명령

`stock-cli`는 다음 초기 설정 명령을 제공합니다.

| Stock 명령 | 목적 |
| --- | --- |
| `stock config set` | Kiwoom REST API credential을 대화형으로 설정합니다. |
| `stock config show` | 현재 Kiwoom credential을 마스킹해서 보여 주고, 각 값의 출처를 표시합니다. |
| `stock config path` | Stock CLI 설정 파일 경로를 출력합니다. |

Kiwoom에서 장기 저장되는 credential은 `appkey`와 `secretkey`뿐입니다.
런타임 token 발급은 고정 Kiwoom host `https://api.kiwoom.com`을 사용합니다.
발급된 access token은 이 초기 설정 명령의 대상이 아닙니다.

## Stock CLI Credential 저장소

Kiwoom credential은 다음 위치에 저장합니다.

```text
~/.stock/config
```

하나의 Kiwoom credential 섹션을 가진 TOML 형식을 사용합니다.

```toml
[kiwoom]
appkey = "..."
secretkey = "..."
```

Credential 해석은 설정 파일만 사용합니다.

1. 설정 파일: `~/.stock/config`.
2. 설정 파일에 credential이 없으면 누락 메시지를 반환합니다.

`~/.stock/`는 `0700` 권한으로 만들고, `~/.stock/config`는 `0600` 권한으로
씁니다. `stock config set`은 의도적으로 TTY 전용입니다. App Key와 Secret
Key만 입력받습니다.

`stock config set`이 credential 저장에 성공하면 기본 발급-token 캐시인
`~/.stock/token.json`을 제거합니다. token 캐시가 없어도 오류가 아닙니다.
다음 token 사용 명령, 예를 들어 `stock accounts list`, 은 새로 저장된
credential로 token을 다시 발급합니다. 이 무효화는 성공한 `stock config set`에만
적용됩니다.

## 확인된 Kiwoom Token 발급 명세

2026-06-10 KST의 직접 호출로 확인했습니다.

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

관측된 성공 응답:

- HTTP status: `200`
- `expires_dt`: compact `YYYYMMDDHHMMSS` 문자열
- `token_type`: `Bearer`로 관측됨
- `token`: access token 문자열
- `return_code`: `0`으로 관측됨
- `return_msg`: `정상적으로 처리되었습니다`로 관측됨

관측된 성공 token 발급 응답에는 `next-key`, `cont-yn`, `api-id` header가
포함되지 않았습니다.

## Token 저장 정책

발급된 API token은 `~/.stock/config`에 저장하지 않습니다. 설정 파일에는
장기 사용자 credential만 저장합니다.

Token 캐싱은 Kiwoom access token이 필요한 API 명령에서 구현하며,
`stock accounts list`부터 사용합니다. 캐시는 `~/.stock/token.json`을 사용하고,
`~/.stock/config`와 분리되어 있으며, 발급된 token 값을 출력하거나 commit하면
안 됩니다. 성공한 `stock config set`은 이 캐시를 제거하므로 다음 token 사용
명령은 활성 credential source로 새 token을 발급합니다.
