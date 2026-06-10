# stock-cli

Kiwoom REST API를 사용하기 위한 Go 기반 주식 CLI입니다. 현재 구현된 범위는
초기 프로젝트 구조와 Kiwoom 기본 설정 명령입니다.

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

성공 응답에서 `next-key`, `cont-yn`, `api-id` 헤더는 관측되지 않았습니다. 토큰
캐시는 아직 구현하지 않았고, `~/.stock/config`에는 발급 토큰을 저장하지
않습니다.
