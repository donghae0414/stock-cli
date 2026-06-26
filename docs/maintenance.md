# Stock CLI Maintenance

이 문서는 maintainer용 점검과 배포 메모를 첫 실행용 README와 분리해 둡니다.

## 안전한 로컬 점검

배포, PR 생성, release 준비 전에 다음을 실행하세요.

```sh
go test ./...
go vet ./...
go build -o bin/stock ./cmd/stock
npm run build:local
npm run build:targets
npm run verify:package
npm run smoke
npm run smoke:install
npm run pack:dry-run
```

`npm run smoke:install`은 임시 directory에 package tarball을 설치하고,
repo-local fallback에 의존하지 않은 상태로 package binary layout을 확인합니다.

## Secret 및 Generated File Audit

배포 전에 repository 상태와 ignore된 local path를 확인하세요.

```sh
git status --short --branch
git remote -v
git check-ignore -v .env .omx/ .gjc/ .stock/ node_modules/ bin/ npm/bin/ dist/ build/ coverage.out
```

`.env`, `.omx/`, `.gjc/`, local `.stock/` data, generated binaries, build
outputs, coverage files, Kiwoom credential, 발급된 token을 commit하지 마세요.
