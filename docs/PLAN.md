# cldpi — 실행 계획 (Plan / Roadmap)

> 목표 **12주 × 주 14시간 ≈ 168시간** (현실적으로 14~15주). 날짜가 아니라 마일스톤이 단위다.
---

## 운영 원칙

1. **날짜가 아니라 마일스톤이 단위다.** 각 M에는 완료 조건(DoD)이 있다. DoD를 못 채우면 **범위를 깎지 말고 날짜를 미룬다.** 반쯤 만든 기능 5개보다 완성된 기능 3개가 낫다.
2. **M3(약 6주차)부터 매일 실사용한다.** 완성을 기다리지 않는다. 5주차부터 실제 문서를 파이에 넣고 진짜로 쓴다. 매일 쓰는 도구는 고칠 이유가 계속 생겨 죽지 않는다. **이게 12주짜리 계획을 살리는 유일한 장치다.**
3. **매주 커밋 + 30분 회고.** 뭘 배웠는지 한 문단씩 적는다. 나중에 이 프로젝트를 설명할 때 자산이 된다.
4. **한 문제에 4시간 이상 막히면 우회한다.** 특히 cgo 빌드 에러와 FUSE 마운트 행(hang)은 혼자 이틀을 태우기 쉽다. 타임박스를 지킨다.

---

## 마일스톤 로드맵

### M0 — 환경 구축 (Week 1, 14h)
> **DoD:** 노트북에서 라즈베리파이의 Go 프로그램에 TCP로 접속해 문자열을 주고받는다. macFUSE 예제가 맥북에 마운트된다.

- [ ] Go Tour + Effective Go — 구조체·인터페이스·goroutine·channel·`defer`·에러 처리 (6h)
- [ ] Pi OS Lite + SSD 부팅, SSH 키 전용, `ufw`, `unattended-upgrades` (3h)
- [ ] Tailscale 설치 → 노트북에서 파이 접속 확인 (1h)
- [ ] TCP echo 서버/클라이언트(`net.Listen`, goroutine per conn) (3h)
- [ ] macFUSE 설치 + cgofuse memfs 마운트 검증 (Go→cgo→cgofuse→macFUSE→파인더 전체 스택)

> 🔴 **macFUSE 마운트가 진짜 게이트다.** M5(9주차)에 발견하면 계획이 무너진다. **1주차에 반드시 확인한다.**
> **핵심 개념:** TCP는 바이트 스트림이다 — 메시지 경계가 없다.

### M1 — 프로토콜 코덱 (Week 2, 14h)
> **DoD:** `PING` 100개를 동시에 던져도 응답이 정확히 짝지어진다. 랜덤 바이트를 디코더에 넣어도 panic이 안 난다.

- [ ] `internal/proto`: `Frame` 구조체, `Encode`/`Decode` (5h)
- [ ] `io.ReadFull` 기반 프레이밍 + `PAYLOAD_LEN` 상한 검증 (2h)
- [ ] 요청 다중화: `request_id` → `chan Response` 맵 (4h)
- [ ] `PING`/`PONG` + read/write deadline (2h)
- [ ] 퍼징 테스트(`testing.F`) (1h)

> **핵심 개념:** 길이 접두사 프레이밍, 부분 read, 요청 다중화, 신뢰할 수 없는 입력 방어.

### M2 — CAS 저장소 (Week 3-4, 28h)
> **DoD:** 10MB 파일을 올리고 받으면 SHA-256이 일치한다. 1바이트만 고쳐서 다시 올리면 청크 1개만 전송된다. 중간에 끊고 재시도하면 이어서 올라간다.

- [ ] `internal/blob`: `Put`/`Get`/`Has` — tmp→fsync→rename, 2단계 서브디렉터리 (5h)
- [ ] `internal/meta`: SQLite 스키마, 트리 조회/생성 (6h)
- [ ] `internal/chunk`: `Chunker` 인터페이스 + 고정 4MiB (2h)
- [ ] Opcode: `HAS_CHUNKS`·`PUT_CHUNK`·`PUT_COMMIT`·`GET_MANIFEST`·`GET_CHUNK` (6h)
- [ ] ★ 서버 측 해시 재검증 (클라이언트를 믿지 말 것) (2h)
- [ ] `STAT`·`LIST`·`MKDIR`·`RENAME`·`DELETE` (4h)
- [ ] 통합 테스트 + 재개 시나리오 (3h)

> **핵심 개념:** 원자적 쓰기(rename+fsync), 내용 주소 지정, 트랜잭션.

### M3 — 보안 + CLI 🎉 실사용 시작 (Week 5-6, 28h)
> **DoD:** 인증 없는 연결은 전부 거절된다. 경로 탈출 테스트가 모두 통과한다. **그리고 오늘부터 실제 문서를 파이에 넣기 시작한다.**

- [ ] 자체 CA + 서버/클라이언트 인증서(`crypto/x509`) (3h)
- [ ] `tls.Listen` + mTLS (2h)
- [ ] Ed25519 기기 키, 페어링 플로우(6자리 코드) (4h)
- [ ] `AUTH_CHALLENGE`/`AUTH_RESPONSE` 챌린지-리스폰스 (2h)
- [ ] ★ 경로 격리 + 유니코드 정규화(NFC/NFD) + 공격 유닛테스트 (3h)
- [ ] `cldpi pair`/`push`/`pull`/`ls`/`mkdir`/`mv`/`rm`/`stat` (8h)
- [ ] 청크 병렬 전송(`errgroup`), 진행률 바 (4h)
- [ ] 연결 재시도 + 지수 백오프 (2h)

> 🟢 **여기가 분기점이다.** M3가 끝나면 "장난감"에서 "도구"가 된다. 이때부터 매일 쓴다.
> **핵심 개념:** TLS 핸드셰이크, 공개키 인증, 리플레이 방어, 병렬 I/O, 컨텍스트 취소.

### M4 — 버전 관리 + GC (Week 7-8, 28h)
> **DoD:** 파일을 3번 덮어쓴 뒤 1번째 버전으로 복원된다. GC가 고아 청크를 회수하고, **GC 도중 업로드가 들어와도 데이터가 깨지지 않는다.**

- [ ] `LIST_VERSIONS`/`RESTORE` + `cldpi versions`/`restore` (5h)
- [ ] 휴지통(soft delete) + `trash`/`undelete` (3h)
- [ ] `chunk_refs` refcount 관리 (4h)
- [ ] `cldpi-server gc` — ★ 유예 기간(grace period) 필수 (6h)
- [ ] 대용량 파일(1GB+) 테스트, 파이 메모리/CPU 프로파일링 (4h)
- [ ] `meta.db` 자동 백업 스크립트 (2h)
- [ ] systemd 유닛(자동 재시작, 권한 최소화) (4h)

> **핵심 개념:** 참조 카운팅, GC 안전성, append-only 설계, 운영/배포.

### M5 — FUSE 마운트: 읽기 전용 (Week 9-10, 28h)
> **DoD:** 파인더/탐색기에서 `~/cldpi`를 열면 전체 트리가 보이고, PDF를 더블클릭하면 미리보기가 뜬다. `ls`가 답답하지 않다. 윈도우에서도 `Z:`로 마운트된다.

- [ ] `cgofuse` 별도 모듈 세팅, macOS 빌드 (3h)
- [ ] `Getattr`/`Readdir`/`Statfs` → `ls` 동작 (5h)
- [ ] 메타데이터 캐시(TTL + negative 캐시) (3h)
- [ ] `Open`/`Read`/`Release` — manifest offset→chunk 매핑 (6h)
- [ ] 로컬 청크 캐시(CAS 레이아웃, LRU) (4h)
- [ ] 선반입(readahead) — 순차 읽기 감지 (3h)
- [ ] Windows + WinFsp 검증 (4h)

> **핵심 개념:** POSIX 파일시스템 의미론, 커널/유저스페이스 경계, 지연 숨기기(캐시).

### M6 — FUSE 마운트: 쓰기 가능 🏁 완성 (Week 11-12, 28h)
> **DoD:** 파인더에서 파일을 드래그해 넣으면 업로드된다. 워드로 문서를 열어 수정·저장하면 새 버전이 생긴다. 이름 변경/삭제가 된다. **오프라인에서 저장한 것도 재연결 시 올라간다.**

- [ ] `Create`/`Write`/`Truncate` — 로컬 캐시 파일에 pwrite (5h)
- [ ] dirty 추적 + `Flush`/`Release`에서 업로드 (5h)
- [ ] ★ rename-후-저장 패턴 대응 + 통합 테스트(vim, 워드, TextEdit) (5h)
- [ ] `Unlink`/`Rmdir`/`Rename`/`Mkdir` (3h)
- [ ] 업로드 큐 + 재시도(디스크에 상태 기록, 재시작 복구) (6h)
- [ ] dirty 파일 pin(eviction 금지), `-ENOSPC` 처리 (2h)
- [ ] 오래 열린 파일 자동 flush (2h)

> **핵심 개념:** 캐시 일관성, write-back, 실패 복구, 큐 설계.

---

## 시간 총계

| 마일스톤 | 주 | 시간 |
|---|---|---|
| M0 환경 | 1 | 14h |
| M1 프로토콜 | 2 | 14h |
| M2 CAS 저장소 | 3-4 | 28h |
| M3 보안 + CLI 🎉 | 5-6 | 28h |
| M4 버전 관리 + GC | 7-8 | 28h |
| M5 FUSE 읽기 | 9-10 | 28h |
| M6 FUSE 쓰기 🏁 | 11-12 | 28h |
| **합계** | **12주** | **168h** |

> 여유가 전혀 없는 계획이다. 처음 쓰는 언어이고 cgo 빌드·FUSE 디버깅에서 반드시 시간이 새므로 **14~15주**로 보는 게 현실적이다. 그게 실패가 아니라 정상이다.

---

## 실행 리스크 & 대응

| 리스크 | 대응 |
|---|---|
| macFUSE 보안 정책(Apple Silicon 시스템 확장 승인·재부팅) | **1주차(M0)에 미리 검증.** 후반(M5)에 발견하면 계획이 무너진다. |
| cgo 빌드/FUSE 디버깅 시간 초과 | 한 문제 4시간 타임박스 후 우회·질문 |
| `cgofuse`는 cgo라 크로스 컴파일 불가 | CLI는 순수 Go로 유지, FUSE(`cldpifs`)만 별도 모듈로 분리해 OS별 빌드 |
| `meta.db` 유실 → 256GB의 무의미한 해시 덩어리만 남음 | 자동 백업(M4, 선택 아님) |
| 순진한 GC가 커밋 전 청크를 지워 데이터 파손 | 생성 후 N시간 이내 청크는 GC 제외(grace period) |
| 좀비 마운트(데몬이 Unmount 없이 죽음) | 시그널 핸들러에서 `host.Unmount()` 필수 |
| 동기 소진(사이드 프로젝트의 실제 사망 원인) | M3부터 매일 사용, 마감은 밀되 범위는 유지 |

---

## 착수 전 즉시 확인할 것

1. **CGNAT 여부** — 공유기 WAN IP와 `curl ifconfig.me` 결과 비교. 다르면 포트포워딩이 애초에 불가능하다(Tailscale 계획엔 무관하지만 미리 알아둔다).
2. **macFUSE가 내 맥북에서 뜨는지** — Apple Silicon + 최신 macOS는 보안 정책 때문에 번거로울 수 있다. 후반에 발견하면 늦다.
3. **SSD가 USB 어댑터인지 NVMe HAT인지** — USB3면 ~350MB/s, NVMe HAT면 ~800MB/s로 처리량 기대치가 다르다.
