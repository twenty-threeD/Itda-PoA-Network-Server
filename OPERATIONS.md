# 운영 가이드 — 배포된 체인에서 결제 기록(거래) 제출하기

블록은 올라가지만 아직 "거래 제출"이 안 되는 상태를 정리한 문서입니다.
현재 코드(`x/payment`, `cmd/itdad/cmd/genesis.go`, `app/app.go`) 기준으로 작성했습니다.

---

## 1. 현재 구조 요약

| 항목 | 값 |
|---|---|
| 거래 = | `MsgRecordPayment` → `PaymentRecord` (order_id 기본키) |
| 제출자(서명자) | **`authority` 단 하나** (`option (cosmos.msg.v1.signer) = "authority"`) |
| 권한 검사 | `msg_server.go` 에서 `k.GetAuthority(ctx) != msg.Authority` 이면 거부 |
| authority 출처 | **제네시스의 `app_state.payment.authority`** 로만 설정됨 |
| 수수료 | ante 에 `DeductFeeDecorator` 포함 (auth 기본 핸들러) |
| CLI | `itdad tx payment record-payment ...` (autocli 로 자동 생성) |

핵심: **`buyer_address` 는 서명자가 아니라 데이터 필드**입니다. 구매자가 직접 트랜잭션을
보내는 구조가 아니라, 백엔드 서버가 authority 키로 대신 기록(대행 제출)하는 모델입니다.
`buyer_signature` 도 체인에서 검증하지 않고 길이만 확인 후 저장합니다(≤256자).

---

## 2. 지금 빠져 있는 작업 (원인 진단)

블록이 올라간다 = `InitGenesis` 는 통과했다는 뜻입니다. `payment` 의 `GenesisState.Validate()`
는 authority 가 비면 실패하므로, **authority 주소 자체는 제네시스에 들어가 있습니다.**
그럼에도 거래가 안 들어가는 전형적인 원인은 아래 셋입니다.

1. **authority 주소의 키를 아무도 갖고 있지 않음.**
   `genesis set-authority` 에 임의 주소를 넣었다면 그 키가 없으면 영원히 제출 불가입니다.
   authority 는 온체인 상태(`AuthorityKey`)로 저장되고, 이를 바꾸는 Msg 가 없으므로
   **잘못 넣었으면 체인 재초기화 외에는 방법이 없습니다.**
2. **authority 계정이 auth 모듈에 존재하지 않음.**
   `genesis add-genesis-account` 를 authority 주소에 대해 실행하지 않았다면 서명 시
   `account ... not found` 로 트랜잭션이 거부됩니다.
3. **수수료 잔액 / min-gas-prices 불일치.**
   `app.toml` 의 `minimum-gas-prices` 가 설정돼 있는데 authority 계정에 해당 denom 잔액이
   없으면 `insufficient funds` 로 mempool 에서 탈락합니다.

참고로 `payment` 쿼리에는 authority 조회 RPC 가 없습니다. 현재 authority 확인은
`itdad export --home /var/lib/itdad | jq '.app_state.payment.authority'` 로 합니다.

---

## 3. 결정 트리 — 재초기화가 필요한가?

```
authority 주소의 개인키를 갖고 있는가?
├─ 아니오 → 체인 재초기화 필요 (4장)
└─ 예
   └─ 그 계정이 genesis account 로 등록돼 있는가? (itdad q auth account <addr>)
      ├─ 아니오 → 재초기화 필요 (auth 계정은 제네시스나 코인 수령으로만 생성됨)
      └─ 예 → 재초기화 불필요. 5장(제출 절차)로 진행
```

> 계정 존재 여부: `itdad query auth account <authority-addr> --node http://<host>:26657`
> `key not found` 가 나오면 등록 안 된 것입니다.

---

## 4. 체인 재초기화 절차 (제네시스를 고쳐야 하는 경우)

기존 상태는 전부 버려집니다. 운영 중이라면 유지보수 공지 후 진행하세요.

```bash
# 0) 서버에서 중지 + 백업
sudo systemctl stop itdad
sudo cp -a /var/lib/itdad /var/lib/itdad.bak.$(date +%F)

# 1) 제출자(authority) 키 생성 — 백엔드 서버가 보관할 키
itdad keys add itda-authority --keyring-backend file --home /var/lib/itdad
AUTH=$(itdad keys show itda-authority -a --keyring-backend file --home /var/lib/itdad)
# 니모닉은 반드시 별도 보관. 이 키를 잃으면 어떤 결제도 기록할 수 없습니다.

# 2) 상태 초기화 + 제네시스 재생성 (priv_validator_key.json 은 유지하고 싶으면 unsafe-reset-all 사용)
itdad tendermint unsafe-reset-all --home /var/lib/itdad --keep-addr-book
itdad init <moniker> --chain-id itda-1 --home /var/lib/itdad --overwrite

# 3) authority 계정을 제네시스 계정으로 등록 (2번 원인 해소)
itdad genesis add-genesis-account $AUTH 1000000000stake --home /var/lib/itdad

# 4) poa / payment authority 지정
itdad genesis set-authority $AUTH --home /var/lib/itdad

# 5) 제네시스 검증자 등록 (poa 는 validators 가 비면 InitGenesis 실패)
itdad genesis add-genesis-validator $AUTH --moniker <moniker> --power 10 --home /var/lib/itdad

# 6) 검증 후 기동
itdad genesis validate-genesis --home /var/lib/itdad
sudo systemctl start itdad
```

`app.toml` 권장 설정 (퍼미션드 체인이라 수수료가 의미 없음):

```toml
minimum-gas-prices = "0stake"

[api]
enable = true
address = "tcp://0.0.0.0:1317"

[grpc]
enable = true
address = "0.0.0.0:9090"
```

> 검증자를 2대 이상 두려면 각 노드에서 `itdad init` 후 동일한 genesis.json 을 복사하고,
> 각 노드의 `priv_validator_key.json` 공개키로 `add-genesis-validator` 를 각각 실행한
> 제네시스를 만들어 배포해야 합니다. 이미 기동한 체인에 노드를 추가할 때는
> `itdad tx poa add-validator` 를 authority 로 실행하면 됩니다.

---

## 5. 거래(결제) 제출 절차

### CLI 로 1건 기록

```bash
itdad tx payment record-payment \
  ORD-20260822-0001 \
  cosmos1buyer... \
  50000 \
  2026-08-22T10:30:00Z \
  https://itda.example/contracts/ORD-20260822-0001.pdf \
  sha256:abcdef... \
  "" \
  --from itda-authority \
  --keyring-backend file \
  --chain-id itda-1 \
  --gas auto --gas-adjustment 1.3 --fees 0stake \
  --node http://<host>:26657 \
  --home /var/lib/itdad -y
```

인자 제약 (`x/payment/types/record.go`):

- `order_id`: 비어 있으면 안 됨, ≤128자, **중복 불가** (재제출 시 `ErrDuplicateOrderID`)
- `buyer_address`: 유효한 bech32 (`cosmos1...`)
- `amount`: **양수만** (0 이하 거부)
- `paid_at`: **RFC 3339** (`2026-08-22T10:30:00Z`)
- `contract_url`: 비어 있으면 안 됨, ≤2048자
- `payment_hash` ≤128자, `buyer_signature` ≤256자 (둘 다 빈 값 허용)

### 조회

```bash
itdad query payment payment ORD-20260822-0001 --node http://<host>:26657
itdad query payment payments --node http://<host>:26657
# REST (api.enable = true 인 경우)
curl http://<host>:1317/itda/payment/v1/payments/ORD-20260822-0001
```

### 백엔드 서버 연동 시 유의점

- authority 키는 **백엔드 1곳에서만** 보관·사용합니다. 계정 시퀀스가 하나이므로
  **동시 제출은 시퀀스 충돌**을 일으킵니다. 결제 기록은 큐로 직렬화해서 보내세요.
- 제출 실패 시 재시도는 `order_id` 기준 멱등 처리 — 이미 기록됐으면 `Payment` 쿼리로
  확인 후 성공 처리(중복 제출은 어차피 거부됩니다).
- 성공 여부는 tx 해시가 아니라 **블록 포함 후 `code == 0`** 으로 판정합니다.
  이벤트 `record_payment` 의 `order_id` / `recorded_height` 속성을 인덱싱하면 됩니다.

---

## 6. 이후 개선 후보 (지금은 안 되어 있는 것)

- **authority 교체 수단 부재** — `MsgUpdateAuthority` 가 없어 키 유출 시 재초기화뿐입니다.
  운영 전에 추가를 권장합니다.
- **`buyer_signature` 미검증** — 체인은 길이만 봅니다. 구매자 서명의 진위는 백엔드가
  검증하거나, `MsgRecordPayment` 에 secp256k1 서명 검증을 넣어야 실효가 생깁니다.
- **authority 조회 쿼리 부재** — `Query.Authority` RPC 추가 시 운영 확인이 쉬워집니다.
- **`Params` 가 빈 메시지** — 수수료/보존기간 등 정책을 넣을 자리로 남아 있습니다.
