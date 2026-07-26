# cldpi — 아키텍처

> Go / 자체 바이너리 프로토콜(TLS 1.3 + mTLS) / CAS 기반 저장·버전 관리 / FUSE 마운트(읽기 → 쓰기).
> 라즈베리파이5 + SSD 위에서 동작하는 단일 사용자 개인 클라우드 스토리지.

---

## 1. 설계의 심장 — CAS(Content-Addressed Storage)

이 프로젝트의 모든 설계는 **"파일은 청크의 목록이고, 청크는 자기 해시로 식별된다"** 는 한 문장에서 나온다(Git과 같은 원리).

```
파일 (report.pdf, 10MB)
  ├─ 4MiB 고정 청크로 분할 ──────► [chunk0][chunk1][chunk2]
  ├─ 각 청크를 SHA-256 해시 ──────► a3f9..., b71c..., 09de...
  ├─ 청크를 해시 이름으로 디스크에 저장
  │     /data/blobs/a3/f9/a3f9c8...
  └─ Manifest = 청크 해시의 순서 있는 목록
        {"chunks": ["a3f9...", "b71c...", "09de..."], "size": 10485760}
        └─ Manifest 직렬화 바이트의 SHA-256 = "파일 버전 ID"
```

이 구조가 공짜로 주는 것:

| 기능 | 원리 |
|---|---|
| 버전 관리 | 수정하면 새 manifest 생성, 옛 manifest 보존 → 복원은 옛 manifest를 가리키게 하기 |
| 증분 업로드 | 10MB에서 1바이트만 고쳐도 바뀐 청크 1개만 전송, 나머지는 서버에 이미 존재 |
| 중복 제거 | 같은 내용 = 같은 해시 = 디스크 1벌 |
| 업로드 재개 | 도착한 청크는 서버에 남음. `HAS_CHUNKS`로 없는 것만 재전송 |
| 무결성 검증 | 다운로드한 청크의 해시를 재계산·대조해 조용한 손상 탐지 |
| 원자적 커밋 | 청크를 다 올린 뒤 manifest만 등록. 중간에 죽으면 아무 일도 없음 |

**청킹 전략:** 고정 4MiB로 시작한다. 단점은 파일 앞부분에 1바이트를 *삽입*하면 이후 청크가 전부 밀려 새 해시가 되는 것(boundary shift). 진짜 삽입-내성 증분 동기화가 필요하면 FastCDC(Content-Defined Chunking)로 가지만, 문서/사진 위주 사용에서는 손해가 크지 않다. **`Chunker` 인터페이스로 분리**해 두면 나중에 FastCDC 교체가 하루 작업이다.

## 2. 시스템 구성

```
┌────────────────────── 노트북/데스크탑 ──────────────────────┐
│                                                             │
│   ┌───────────────┐        ┌───────────────────────────┐   │
│   │  cldpi (CLI)  │        │  cldpifs (FUSE 마운트)    │   │
│   │  push/pull    │        │  ~/cldpi (읽기→쓰기)      │   │
│   │  ls/mv/rm     │        │  + 로컬 청크 캐시(CAS)    │   │
│   │  versions     │        │  + 메타데이터 캐시(TTL)   │   │
│   └───────┬───────┘        └─────────────┬─────────────┘   │
│           └──────────┬───────────────────┘                 │
│              ┌────────▼─────────┐                           │
│              │  client/protocol │ ← 공유 라이브러리         │
│              │  (Go 패키지)     │   프레임 인·디코딩,       │
│              └────────┬─────────┘   요청 다중화, 재연결     │
└───────────────────────┼─────────────────────────────────────┘
                        │  TLS 1.3 + mTLS  (Tailscale 사설망)
┌───────────────────────▼──────────── Raspberry Pi 5 ─────────┐
│  Transport Layer                                            │
│    tls.Listener → conn당 goroutine → request_id로 다중화    │
│  ─────────────────────────────────────────────────────────  │
│  Session / Auth Layer                                       │
│    기기 인증(Ed25519), 세션 상태, rate limit                │
│  ─────────────────────────────────────────────────────────  │
│  Handler Layer                                              │
│    opcode → 핸들러 라우팅, ★ 경로 격리(path traversal 방어) │
│  ─────────────────────────────────────────────────────────  │
│  ┌ Metadata (SQLite) ┐        ┌ Blob Store (CAS on SSD) ┐  │
│  │ 트리·버전·manifest │◄──────►│ /data/blobs/aa/bb/aabb… │  │
│  │ ·기기              │        │                         │  │
│  └────────────────────┘        └─────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 계층 원칙 (이걸 지키면 저장소 교체·암호화 추가 시 protocol 코드 무수정)
1. **Blob Store는 파일명을 모른다** — 오직 해시 → 바이트만 안다.
2. **Metadata DB는 바이트를 모른다** — 오직 경로 → manifest 해시만 안다.
3. **Protocol Layer는 저장 방식을 모른다** — 인터페이스로만 접근한다.

## 3. 저장소 설계

### 3-1. Blob Store 디스크 레이아웃
```
/data/
  blobs/     a3/f9/a3f9c8...       ← 청크 파일. 이름 = SHA-256 hex
  manifests/ 4e/2a/4e2a7b...       ← manifest JSON. 이름 = manifest 해시
  tmp/       upload-<sess>-<n>     ← 업로드 중 임시 파일 (완료 시 rename)
  meta.db                          ← SQLite
```
- **2단계 서브디렉터리**(`a3/f9/`): 한 디렉터리에 파일 수십만 개가 쌓이면 ext4도 느려진다.
- **쓰기는 반드시 tmp → fsync → rename**: rename은 같은 파일시스템 내에서 원자적이라, 중간에 전원이 나가도 반쪽짜리 blob이 생기지 않는다.

### 3-2. Metadata (SQLite 스키마)
```sql
-- 등록된 기기
CREATE TABLE devices (
    id         INTEGER PRIMARY KEY,
    name       TEXT NOT NULL,           -- "macbook", "desktop"
    pubkey     BLOB NOT NULL UNIQUE,    -- Ed25519 공개키 (32바이트)
    created_at INTEGER NOT NULL,
    last_seen  INTEGER,
    revoked    INTEGER NOT NULL DEFAULT 0
);

-- 파일시스템 트리 (디렉터리 + 파일의 "현재 상태")
CREATE TABLE nodes (
    id          INTEGER PRIMARY KEY,
    parent_id   INTEGER REFERENCES nodes(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    is_dir      INTEGER NOT NULL,
    cur_version INTEGER REFERENCES versions(id),  -- 파일이면 현재 버전을 가리킴
    created_at  INTEGER NOT NULL,
    updated_at  INTEGER NOT NULL,
    UNIQUE(parent_id, name)                        -- 같은 폴더에 같은 이름 금지
);

-- 파일의 모든 버전 (append-only, 절대 UPDATE/DELETE 안 함)
CREATE TABLE versions (
    id            INTEGER PRIMARY KEY,
    node_id       INTEGER NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    manifest_hash TEXT NOT NULL,                   -- CAS의 manifest 해시
    size          INTEGER NOT NULL,
    created_at    INTEGER NOT NULL,
    device_id     INTEGER REFERENCES devices(id),
    comment       TEXT
);
CREATE INDEX idx_versions_node ON versions(node_id, created_at DESC);

-- 청크 참조 카운트 (GC용)
CREATE TABLE chunk_refs (
    chunk_hash TEXT PRIMARY KEY,
    refcount   INTEGER NOT NULL,
    size       INTEGER NOT NULL
);
```

**핵심 포인트**
- `versions`는 **절대 수정하지 않는다**(append-only). 덮어쓰기 = 새 row 추가 + `nodes.cur_version`만 이동. 이게 버전 관리의 전부다.
- 삭제는 `nodes`에서만 제거하고 `versions`·blob은 보존한다(휴지통). 실제 blob 삭제는 GC 명령 전용.
- `chunk_refs`가 없으면 어떤 청크를 지워도 되는지 알 수 없다. 매 커밋마다 refcount를 갱신한다.

### 3-3. Manifest 포맷
```json
{ "v": 1, "size": 10485760, "chunk_size": 4194304,
  "chunks": ["a3f9c8...", "b71c22...", "09de41..."] }
```
이 JSON을 직렬화한 바이트의 SHA-256이 곧 manifest 해시이자 파일 버전 ID다.

## 4. 프로토콜 (v1)

### 4-1. 설계 원칙
- **길이 접두사 프레이밍:** TCP는 스트림이라 메시지 경계가 없다. 직접 만들어야 한다.
- **요청 다중화:** 하나의 TLS 연결 위에서 여러 요청이 동시에 오간다. `request_id`로 응답을 짝짓는다(FUSE는 병렬 요청을 마구 던진다).
- **바이너리 헤더 + 유연한 페이로드:** 헤더는 고정 크기 바이너리, 페이로드는 JSON(제어) 또는 raw bytes(데이터).

### 4-2. 프레임 포맷 (헤더 16 bytes)
```
+-------+-------+-------+-------+-------+-------+-------+-------+
| MAGIC "CLDP" (4 bytes)        | VER   | OPCODE| FLAGS (2B)    |
+-------+-------+-------+-------+-------+-------+-------+-------+
| REQUEST_ID (uint32, BE)       | PAYLOAD_LEN (uint32, BE)      |
+-------+-------+-------+-------+-------+-------+-------+-------+
| PAYLOAD ... (PAYLOAD_LEN bytes)                               |
+---------------------------------------------------------------+
```
| 필드 | 의미 |
|---|---|
| MAGIC | `"CLDP"` — 잘못된 연결 조기 탐지 |
| VER | uint8, 현재 1 |
| OPCODE | uint8 |
| FLAGS | uint16 — bit0 IS_RESPONSE, bit1 IS_ERROR, bit2 MORE(스트리밍 계속) |
| REQUEST_ID | uint32 — 클라이언트가 증가시키고, 응답은 같은 id를 되돌려줌 |
| PAYLOAD_LEN | uint32 — ★ 반드시 상한(예: 8MiB). 없으면 대용량 할당 요청으로 서버가 죽는다(DoS) |

구현은 `io.ReadFull`로 16바이트를 읽고 `PAYLOAD_LEN`을 파싱한 뒤 그만큼 다시 `ReadFull`하는 코덱이 기본이다.

### 4-3. Opcode 목록
| Opcode | 이름 | 방향 | 페이로드 | 설명 |
|---|---|---|---|---|
| 0x01 | `HELLO` | C→S | JSON | 프로토콜 버전 협상, 기기 이름 |
| 0x02 | `AUTH_CHALLENGE` | S→C | 32B raw | 서버가 랜덤 nonce 전송 |
| 0x03 | `AUTH_RESPONSE` | C→S | 64B raw | 클라이언트가 nonce에 Ed25519 서명 |
| 0x04 | `PAIR_REQUEST` | C→S | JSON | 최초 기기 등록(페어링 코드 + 공개키) |
| 0x10 | `STAT` | C→S | JSON `{path}` | 파일/폴더 메타데이터 |
| 0x11 | `LIST` | C→S | JSON `{path}` | 디렉터리 목록 |
| 0x12 | `MKDIR` | C→S | JSON `{path}` | 폴더 생성 |
| 0x13 | `RENAME` | C→S | JSON `{from,to}` | 이름 변경/이동 |
| 0x14 | `DELETE` | C→S | JSON `{path}` | 삭제(휴지통으로) |
| 0x20 | `HAS_CHUNKS` | C→S | JSON `{hashes[]}` | ★ 어떤 청크가 이미 있는지 질의 |
| 0x21 | `PUT_CHUNK` | C→S | raw bytes | 청크 1개 업로드 |
| 0x22 | `PUT_COMMIT` | C→S | JSON `{path, manifest}` | manifest 등록 = 새 버전 확정 |
| 0x30 | `GET_MANIFEST` | C→S | JSON `{path, version?}` | manifest 조회 |
| 0x31 | `GET_CHUNK` | C→S | JSON `{hash}` | 청크 1개 다운로드 |
| 0x40 | `LIST_VERSIONS` | C→S | JSON `{path}` | 버전 이력 |
| 0x41 | `RESTORE` | C→S | JSON `{path, version}` | 특정 버전으로 되돌리기 |
| 0x50 | `PING` | C→S | — | keepalive |
| 0xFF | `ERROR` | S→C | JSON `{code, msg}` | 에러 응답 |

### 4-4. 업로드 흐름 (재개 + 중복 제거)
```
1. 로컬에서 파일을 4MiB로 청킹 + 각 청크 SHA-256 계산
2. HAS_CHUNKS {hashes:[h0,h1,h2,h3]}  →  서버가 {missing:[h1,h3]} 반환
3. 없는 청크만 PUT_CHUNK  →  서버가 tmp에 쓰고 ★ 해시 재계산·검증 후 fsync→rename
4. PUT_COMMIT {path, manifest}  →  트랜잭션:
     manifest 저장 · versions INSERT · nodes.cur_version 갱신 · chunk_refs 증가
   →  {version_id}
```
★ **서버는 클라이언트가 보낸 해시를 절대 믿지 않는다.** 받은 바이트를 직접 해싱해 대조한다. 안 그러면 악의적 클라이언트가 임의 해시로 임의 내용을 심어 CAS를 오염시킬 수 있다.
중간에 끊겨도 도착한 청크는 남으므로, 재시도 시 `HAS_CHUNKS`가 알아채고 나머지만 올린다 — 재개는 별도 로직이 필요 없다.

### 4-5. 에러 코드
```
1000 BAD_FRAME           프레임 파싱 실패
1001 PAYLOAD_TOO_LARGE   상한 초과
2000 UNAUTHENTICATED     인증 안 된 상태로 명령
2001 AUTH_FAILED         서명 검증 실패
2002 DEVICE_REVOKED      차단된 기기
3000 NOT_FOUND           경로 없음
3001 ALREADY_EXISTS      이름 충돌
3002 NOT_A_DIR / IS_A_DIR
3003 INVALID_PATH        ★ 경로 탈출 시도
4000 CHUNK_HASH_MISMATCH 해시 검증 실패
5000 INTERNAL            서버 내부 오류
```

## 5. 보안

### 5-1. 기기 페어링 (최초 1회)
```
1. 파이에서:   cldpi-server pair       → 콘솔에 6자리 코드 출력(5분 유효)
2. 노트북에서: cldpi pair --server pi.tailnet --code 384921
              → Ed25519 키쌍 생성, 개인키는 OS 키체인 또는 0600 파일로 저장
              → 공개키 + 페어링 코드를 PAIR_REQUEST로 전송
3. 서버:      코드 검증 → devices 테이블에 공개키 등록
```
이후 모든 연결은 비밀번호 없이 챌린지-리스폰스로 인증된다.

### 5-2. 경로 격리 — 가장 중요한 방어
클라이언트가 보내는 경로는 **전부 적대적 입력**으로 간주한다.
```go
// ❌ 절대 금지 — "../../etc/passwd"가 통과됨
full := filepath.Join(rootDir, userPath)

// ✅ 경로 문자열을 파일시스템 경로로 매핑하지 않고 DB 트리로만 해석
//    "/docs/a.pdf" → nodes 테이블을 부모부터 따라 내려가며 id로 해석
//    → ".."는 애초에 유효한 이름이 아니게 된다
```
**숨은 장점:** blob은 해시 이름으로만 저장되고 논리 경로는 SQLite에만 존재한다. 즉 사용자 경로가 파일시스템 경로로 변환되는 지점이 아예 없어 path traversal이 구조적으로 불가능하다.

그래도 반드시 검증한다:
- 경로 요소에 `..`, `.`, `/`, `\`, NUL 금지
- 이름 길이 제한(255바이트)
- **유니코드 정규화**(macOS는 NFD, 나머지는 NFC) — 안 하면 맥에서 만든 한글 파일명이 다른 기기에서 안 보인다
- `../`, `..\`, `%2e%2e`, 절대경로, NUL 바이트 등을 전부 거부하는지 **유닛테스트로 고정**한다.

### 5-3. TLS 구성
```go
tlsConfig := &tls.Config{
    MinVersion:   tls.VersionTLS13,
    Certificates: []tls.Certificate{serverCert},
    ClientAuth:   tls.RequireAndVerifyClientCert,  // ★ mTLS
    ClientCAs:    caPool,
}
// InsecureSkipVerify는 절대 true 금지. 클라이언트도 RootCAs를 자체 CA로만 지정.
```
자체 CA를 만들어 서버·클라이언트 인증서를 발급한다.

### 5-4. 그 외 필수 방어
- `PAYLOAD_LEN` 상한(8MiB) — 없으면 메모리 고갈 DoS
- 연결당 read/write 타임아웃(`SetReadDeadline`) — 없으면 slowloris
- 인증 전 상태에서 처리 가능한 opcode를 화이트리스트로 제한
- 페어링 시도 rate limit
- 서버는 전용 유저로 실행(root 금지), systemd + `ProtectSystem=strict`

## 6. FUSE 마운트 (읽기 → 쓰기)

### 6-1. 핵심 통찰 — 구글 드라이브도 부분 쓰기를 하지 않는다
"쓰기 가능 FUSE"의 난이도는 **진짜 POSIX 파일시스템**(mmap, DB 파일, 하드링크, offset 단위 원격 쓰기)을 만들 때의 얘기다. 구글 드라이브(및 rclone VFS)의 실제 동작은 다음과 같다:
```
open()  → 서버에서 파일을 로컬 캐시로 다운로드(또는 청크 온디맨드)
write() → 로컬 캐시 파일에 그냥 pwrite. 네트워크 0.
read()  → 로컬 캐시 파일에서 그냥 pread. 네트워크 0.
close() → dirty면: 청킹 → HAS_CHUNKS → 바뀐 청크만 업로드 → PUT_COMMIT
```
**부분 쓰기의 복잡성을 로컬 캐시 파일이 전부 흡수한다.** 서버는 늘 "통째로 새 버전"만 받되, CAS 덕분에 실제로 전송되는 건 바뀐 청크뿐이다. 그래서 2단계로 간다:
- **읽기 전용(M5):** 파인더에 뜨고 열리고 읽힌다. 이미 실용적이다.
- **쓰기 가능(M6):** 드래그 앤 드롭, 저장, 삭제, 이름변경. 읽기 경로(manifest→offset→chunk 매핑, 캐시)를 먼저 안정화한 뒤 쓰기를 얹는다.

### 6-2. 구현할 콜백
**읽기 전용 — 6개**

| 콜백 | 하는 일 |
|---|---|
| `Getattr` | 크기·mtime·모드 → **캐시 필수**(FUSE가 미친 듯이 호출) |
| `Readdir` | `LIST` → 자식 목록 |
| `Open` | 파일 핸들 발급, `GET_MANIFEST` 받아둠 |
| `Read` | ★ 핵심 — offset→chunk 매핑(아래) |
| `Release` | 핸들 정리 |
| `Statfs` | 용량 정보(파인더가 요구) |

이 단계에서 `Write`/`Create`/`Unlink` 등은 전부 `-EROFS` 반환.

**쓰기 가능 — 7개 추가**

| 콜백 | 하는 일 |
|---|---|
| `Create` | 로컬 캐시에 빈 파일 생성 + dirty 표시 |
| `Write` | 로컬 캐시 파일에 pwrite. 네트워크 없음 |
| `Truncate` | 로컬 캐시 파일 자르기 |
| `Flush`/`Fsync` | dirty면 서버로 업로드(★ 오래 열린 파일 대응) |
| `Release` | 남은 dirty 업로드 + 핸들 정리 |
| `Unlink`/`Rmdir` | `DELETE` opcode |
| `Rename` | `RENAME` opcode(★ 덮어쓰기·원자성 주의) |
| `Mkdir` | `MKDIR` opcode |

### 6-3. Read의 핵심 — 청크 캐시
```
Read(offset=5_000_000, size=4096):
  idx = offset / CHUNK_SIZE          = 1
  off_in_chunk = offset % CHUNK_SIZE = 805,696
  → 로컬 캐시에 chunks[1]이 있으면 캐시에서 반환(네트워크 0)
    없으면 GET_CHUNK로 받아 캐시에 저장 후 반환
  → 요청이 청크 경계를 넘으면 다음 청크도 처리
```
**캐시 설계**
- 위치: `~/.cldpi/cache/blobs/` — **서버와 똑같은 CAS 레이아웃**.
- **해시가 이름이므로 캐시 무효화가 필요 없다.** 내용이 바뀌면 해시가 바뀌어 곧 다른 파일이다.
- LRU로 용량 상한(예: 5GB) 관리.
- **선반입(readahead):** 순차 읽기가 감지되면 다음 청크를 미리 받는다(없으면 동영상 재생이 끊긴다).

**메타데이터 캐시**
- `Getattr`/`Readdir` 결과를 TTL(예: 30초)로 캐시. 없으면 `ls` 한 번에 네트워크 요청이 수십 번 나간다.
- **negative 캐시**도 필요하다. macOS는 `.DS_Store`, `._*` 같은 없는 파일을 끝없이 찾는다.

### 6-4. 쓰기 경로의 진짜 함정
1. **rename-후-저장 패턴** — vim·워드 등 대부분의 에디터는 `tmp 생성 → 쓰기 → fsync → rename(tmp→원본)`으로 저장한다. `Rename`이 **기존 파일 덮어쓰기**를 지원해야 한다(CAS에선 `nodes.cur_version`만 옮기면 자연스레 새 버전이 생김). **저장 안 되는 버그의 90%가 여기서 나온다** — 통합 테스트로 고정.
2. **오래 열린 파일** — 워드/엑셀은 문서를 몇 시간씩 열어둔다. `Release`까지 기다리면 그동안 서버엔 아무것도 없다. → `Flush`에서도 업로드하고, "dirty가 N초 지속되면 자동 업로드"를 넣는다.
3. **캐시 용량 초과** — dirty 파일은 아직 원본이 서버에 없으므로 **절대 eviction 대상이 아니다**(pin). 공간이 정말 부족하면 `-ENOSPC` 반환.
4. **업로드 실패** — `Release`에서 업로드가 실패해도 유저는 이미 파일을 닫았다. → **업로드 큐 + 재시도**가 필요하다. dirty 캐시는 성공할 때까지 남기고, 재시작해도 복구되도록 큐 상태를 디스크에 기록한다.

### 6-5. 운영 주의
- 데몬은 **시그널 핸들러에서 `Unmount()`** 필수. 없으면 Ctrl+C마다 좀비 마운트가 커널 마운트 테이블에 쌓인다.
  ```go
  c := make(chan os.Signal, 1)
  signal.Notify(c, os.Interrupt, syscall.SIGTERM)
  go func() { <-c; host.Unmount(); os.Exit(0) }()
  ```
- 서버엔 유닉스 uid 개념이 없으므로(단일 사용자), uid/gid는 마운트한 사용자로 직접 채운다(`fuse.Getcontext()`).
- macOS에서 터미널에 Full Disk Access가 없으면 마운트는 성공해도 터미널의 `ls`가 `Operation not permitted`가 된다(FUSE 버그 아님, macOS TCC 문제).

### 6-6. OS별 준비
| OS | 필요 | 주의 |
|---|---|---|
| macOS | macFUSE | 시스템 확장 승인 + 재부팅 필요. Apple Silicon은 복구 모드에서 보안 정책 완화가 필요할 수 있음 |
| Windows | WinFsp | 설치만 하면 순조로움. 드라이브 문자로 마운트(`Z:`) |
| Linux(파이) | 커널 내장 | `fusermount3` 필요 |

`cgofuse`는 **cgo를 쓰므로 크로스 컴파일이 불가**하다. 맥용은 맥에서, 윈도우용은 윈도우에서 빌드해야 한다. → CLI(`cldpi`)는 순수 Go로 유지하고, **FUSE 바이너리(`cldpifs`)만 별도 모듈**로 분리한다.

## 7. 프로젝트 구조

```
cldpi/
├── go.mod
├── cmd/
│   ├── cldpi/            # CLI (순수 Go, 크로스컴파일 O)
│   └── cldpi-server/     # 서버 (파이에서 실행)
├── internal/
│   ├── proto/            # 프레임 코덱, opcode, 에러코드
│   ├── blob/             # CAS blob store
│   ├── meta/             # SQLite 메타데이터
│   ├── chunk/            # Chunker 인터페이스 + 고정크기 구현
│   ├── auth/             # Ed25519, 페어링, 세션
│   ├── server/           # 핸들러, 라우팅, 경로격리
│   └── client/           # 공유 클라이언트 라이브러리
└── fs/                   # ★ 별도 모듈 (cgo 의존)
    ├── go.mod
    └── cmd/cldpifs/      # FUSE 마운트 (cgofuse)
```
`fs/`를 별도 모듈로 빼는 이유: cgo 의존이 메인 모듈을 오염시키면 CLI의 크로스 컴파일이 깨진다.

## 8. 향후 확장

계층 분리(§2) 덕분에 아래 항목은 protocol 코드를 거의 건드리지 않고 얹을 수 있다: HTTP/REST 게이트웨이·웹 UI(모바일 접근의 전제) · FastCDC 청킹(`Chunker` 교체) · 포트포워딩 공개 노출(fail2ban·rate limit 강화) · 오프라인 pin · 충돌 처리(충돌 사본 방식) · E2E 암호화(클라이언트 측 청크 암호화).
