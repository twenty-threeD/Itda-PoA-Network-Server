# RPC 프록시 배포

프론트엔드 대시보드(네트워크 신뢰 지표)가 필요로 하는 읽기 전용 엔드포인트만
노출하기 위한 nginx 설정입니다.

## 왜 프록시인가

CometBFT RPC(26657)를 그대로 열면 `broadcast_tx_*`, `dump_consensus_state`,
`unsafe_flush_mempool` 이 함께 열립니다. `authority` 게이트 덕분에 상태 변경은
막히지만 mempool 오염과 DoS 는 가능합니다. 화이트리스트 프록시를 앞에 두고
26657 은 방화벽에서 닫습니다.

## 트래픽 경로

```
브라우저 ──(같은 오리진)──> Next.js /api/blockchain/*
                                 │  서버 → 서버
                                 ├─> blockchain.idta.store   (SDK REST 1317)
                                 │     검증자 수, 블록 헤더 시각
                                 └─> rpc.idta.store/net_info (이 프록시)
                                       피어 수

브라우저 ──(직접)──> wss://rpc.idta.store/websocket   (이 프록시)
                       NewBlock 구독 → 실시간 블록 높이
```

브라우저가 이 프록시에 직접 붙는 경로는 **WebSocket 하나뿐**입니다. 나머지는
Next 서버가 호출하므로 CORS 헤더가 필요 없습니다.

## 치환 대상

| 자리표시자 | 넣을 값 | 위치 |
|---|---|---|
| `rpc.idta.store` | RPC 공개 도메인 | `server_name` 2곳, 인증서 경로 2곳 |
| `https://www.idta.store` | 프론트 오리진 | `map $http_origin` 블록 |

## 배포 절차

```bash
# 1) 노드가 RPC 를 로컬에만 열고 CORS 는 비워 두는지 확인
grep -A2 '^\[rpc\]' /var/lib/itdad/config/config.toml
#   laddr = "tcp://127.0.0.1:26657"
#   cors_allowed_origins = []
#   ↑ 다르면 수정 후 sudo systemctl restart itdad

# 2) DNS A 레코드: rpc.idta.store -> 노드 공인 IP

# 3) 설정 배치
sudo cp deploy/nginx/itda-rpc.conf /etc/nginx/conf.d/   # Rocky Linux
# Debian/Ubuntu 라면 sites-available + sites-enabled 심볼릭 링크

# SELinux 가 Enforcing 이면 업스트림 연결이 막혀 502 가 난다
getenforce && sudo setsebool -P httpd_can_network_connect 1

# 4) Cloudflare Origin Certificate 배치 (아래 "인증서" 절 참고)
sudo mkdir -p /etc/nginx/ssl
sudo install -m 600 /dev/stdin /etc/nginx/ssl/idta-origin.pem   # 발급받은 인증서 붙여넣기
sudo install -m 600 /dev/stdin /etc/nginx/ssl/idta-origin.key   # 발급받은 개인키 붙여넣기

# 5) 문법 검사 후 반영
sudo nginx -t && sudo systemctl reload nginx

# 6) 26657 직접 접근 차단
sudo firewall-cmd --permanent --add-service=http --add-service=https
sudo firewall-cmd --permanent --remove-port=26657/tcp 2>/dev/null
sudo firewall-cmd --reload
```

## DNS 레코드

`idta.store` 의 네임서버는 Cloudflare 다. `dash.cloudflare.com` > `idta.store` >
**DNS** > **Add record** 에서 추가한다.

| 항목 | 값 |
|---|---|
| Type | `A` |
| Name | `rpc` |
| IPv4 address | 노드 공인 IP (`curl -4 ifconfig.me`) |
| Proxy status | **Proxied** (주황 구름) |
| TTL | Auto |

기존 `blockchain`, `api` 레코드와 같은 서버라면 그 레코드의 Content 값을 그대로
쓰면 된다. Proxied 로 두면 오리진 IP 가 가려지고 Cloudflare 의 DDoS 방어를
받는다. WebSocket 은 Cloudflare 프록시에서 기본 지원되므로 실시간 블록 높이도
그대로 동작한다.

## 인증서 — Cloudflare Origin Certificate

Proxied 상태에서는 `certbot --nginx` 의 HTTP-01 검증을 Cloudflare 가 가로채
실패할 수 있다. Origin Certificate 를 쓰면 이 문제가 없고 유효기간이 15년이라
갱신 작업도 사라진다.

1. `dash.cloudflare.com` > `idta.store` > **SSL/TLS** > **Origin Server**
2. **Create Certificate** — 기본값(RSA 2048, 15년), 호스트네임에
   `rpc.idta.store` 추가
3. 화면에 한 번만 표시되는 **Origin Certificate** 와 **Private Key** 를 각각
   `/etc/nginx/ssl/idta-origin.pem`, `/etc/nginx/ssl/idta-origin.key` 로 저장
   (권한 600, 개인키는 재발급 외에 다시 볼 수 없으니 즉시 보관)
4. **SSL/TLS** > **Overview** 에서 모드를 **Full (strict)** 로 설정

> Origin Certificate 는 Cloudflare 만 신뢰하는 인증서다. 브라우저가 오리진에
> 직접 붙으면 경고가 나지만, 정상 트래픽은 전부 Cloudflare 를 거치므로 문제되지
> 않는다. 오히려 우회 접근을 막고 싶다면 conf 의 Authenticated Origin Pulls
> 주석을 풀면 된다.

certbot 을 쓰고 싶다면 레코드를 잠시 **DNS only**(회색 구름)로 바꾸고
`sudo certbot --nginx -d rpc.idta.store` 를 실행한 뒤 다시 Proxied 로 되돌린다.
이 경우 conf 의 `ssl_certificate` 경로를 `/etc/letsencrypt/live/...` 로 바꾼다.

## 동작 확인

```bash
# 허용 경로 — 200
curl -s https://rpc.idta.store/net_info | jq '.result.n_peers'

# 차단 경로 — 403
curl -s -o /dev/null -w '%{http_code}\n' https://rpc.idta.store/dump_consensus_state
curl -s -o /dev/null -w '%{http_code}\n' -X POST https://rpc.idta.store/

# 오리진 없는 WebSocket — 403
curl -s -o /dev/null -w '%{http_code}\n' \
  -H 'Connection: Upgrade' -H 'Upgrade: websocket' \
  https://rpc.idta.store/websocket
```

## 프론트엔드 연결 (23D-Web)

프록시가 뜬 뒤 환경변수 두 개를 채웁니다.

```bash
CHAIN_RPC_URL=https://rpc.idta.store              # 서버 전용. 피어 수 조회
NEXT_PUBLIC_CHAIN_RPC_WS=wss://rpc.idta.store/websocket   # 브라우저. 실시간 높이
```

- `CHAIN_RPC_URL` 을 채우면 `/api/blockchain/network-info` 의 `nodes` 가
  `null` 에서 실제 값으로 바뀝니다.
- `NEXT_PUBLIC_CHAIN_RPC_WS` 를 채우면 블록 높이가 5초 폴링에서 NewBlock
  구독으로 전환됩니다. 비워 두면 폴링으로 계속 동작합니다.

## Cloudflare 를 앞에 두는 경우

- WebSocket 은 Cloudflare 프록시(주황 구름)에서도 동작합니다.
- `$remote_addr` 이 전부 Cloudflare IP 가 되므로 레이트 리밋 키를
  `CF-Connecting-IP` 로 잡아 두었습니다.
- 기존 `blockchain.idta.store`(SDK REST 1317)도 `POST /cosmos/tx/v1beta1/txs`
  가 열려 있습니다. Cloudflare 쪽에도 경로 화이트리스트를 거는 것을 권합니다.
