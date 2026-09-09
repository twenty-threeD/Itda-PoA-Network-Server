# Itda PoA Network Server

Cosmos SDK 기반 Proof-of-Authority 체인. 검증자 집합이 온체인 권한(authority)에
의해서만 변경되는 퍼미션드 네트워크입니다.

## Stack

| 항목 | 값 |
|---|---|
| Language | Go 1.26.5 |
| Framework | Cosmos SDK v0.53.8 |
| Proto | gogoproto v1.7.2 (`protoc-gen-gocosmos`), buf v1.50.0 |
| Module path | `github.com/twenty-threeD/Itda-PoA-Network-Server` |
| Binary | `itdad` |

## Commands

```bash
make tools       # buf + protoc-gen-gocosmos 를 GOBIN 에 설치 (최초 1회)
make proto-gen   # proto/ -> x/<module>/types/*.pb.go 재생성
make proto-lint  # proto 스타일 검사
make build       # build/itdad 생성
make test        # go test ./...
```

`*.pb.go` 는 **생성물이므로 직접 수정하지 않습니다.** 스키마 변경은 `proto/` 를
고치고 `make proto-gen` 을 돌리는 방식으로만 합니다.

## Layout

```
proto/itda/poa/v1/    poa|tx|query|genesis.proto   — 유일한 스키마 소스
x/poa/types/          생성된 pb.go + 수기 작성 타입 코드(codec, errors, keys, validation)
x/poa/keeper/         상태 저장/조회 및 Msg·Query 서버 구현
x/poa/                module.go — AppModule / AppModuleBasic 배선
x/poa/simulation/     (선택) 시뮬레이션
app/                  앱 배선 (app.go, app_config, ante)
cmd/itdad/            체인 바이너리 엔트리포인트
```

## Domain model

- `Params.max_validators` — 활성 검증자 수 상한. 초과 추가는 거부.
- `Params.default_power` — `MsgAddValidator.power == 0` 일 때 적용할 기본 투표권.
- `Validator.operator_address` — bech32 계정 주소이자 **검증자 엔트리의 기본 키**.
- `Validator.consensus_pubkey` — `cosmos.crypto.PubKey` 를 구현하는 `Any`.
- 모든 `Msg` 는 `authority` 서명자로 게이트됨. 이 게이트가 체인을 퍼미션드로 만드는
  핵심이므로, 어떤 상태 변경 경로도 authority 검사를 우회해서는 안 됩니다.
- `GenesisState.validators` 는 비어 있으면 안 됨 (블록 생산 주체 부재).

## Working agreements

- 상세 코딩/테스트 규약은 `.Codex/convention.md` 를 따릅니다.
- 코드 작성·테스트·보안 검토 요청 시 `harness` 스킬을 사용합니다.
- 사용자 응답은 한국어로 합니다.
