# 08 — TCP echo 서버 🏁

**이 장이 M0의 DoD다.**

> 노트북에서 라즈베리파이의 Go 프로그램에 TCP로 접속해 문자열을 주고받는다.

```bash
go test -v ./08_tcp_echo/
go test -race ./08_tcp_echo/
```

---

## TCP는 바이트 스트림이다 ★

이 장에서 딱 하나만 가져간다면 이것이다. PLAN.md에도 M0의 핵심 개념으로 적혀 있다.

> **TCP는 바이트 스트림이다 — 메시지 경계가 없다.**

내가 `Write`를 몇 번 나눠서 했는지, 얼마나 큰 덩어리로 보냈는지는 **상대방에게 전혀 전달되지 않는다.**

```
보낸 쪽                              받는 쪽이 실제로 보는 것
Write("hello\n")                     Read() → "hello\nwor"       ← 합쳐짐
Write("world\n")                     Read() → "ld\n"             ← 쪼개짐
```

커널이 합치기도 하고(Nagle 알고리즘) 쪼개기도 한다(MTU, 혼잡 제어). 중간 라우터도 관여한다. **내 Write 한 번 = 상대 Read 한 번이 절대 아니다.**

그래서 "메시지"라는 개념은 **애플리케이션이 직접 만들어야 한다.** 방법은 크게 셋이다.

| 방법 | 예 | 특징 |
|---|---|---|
| 구분자 | `\n`까지가 한 메시지 | HTTP 헤더, Redis. 간단하지만 데이터에 구분자가 못 들어감 |
| **길이 접두사** | `[4바이트 길이][본문]` | 바이너리 안전. **M1에서 이걸 만든다** |
| 고정 길이 | 항상 64바이트씩 | 가장 단순. 유연성 없음 |

이 장에서는 제일 쉬운 **구분자 방식(`\n`)**을 쓴다. M1에서 길이 접두사로 갈아탄다.

`TestStreamHasNoMessageBoundaries`가 이걸 직접 증명한다. 한 번의 `Write`에 세 메시지를 넣어 보내고, 반대로 한 메시지를 세 번에 나눠 보낸다. 둘 다 서버가 올바르게 처리해야 한다.

---

## Go의 네트워크 API

```go
ln, err := net.Listen("tcp", ":9000")     // 서버: 듣기 시작
conn, err := ln.Accept()                  // 서버: 접속 하나 받기 (없으면 기다림)
conn, err := net.Dial("tcp", "host:9000") // 클라이언트: 접속하기
```

**`net.Conn`은 `io.Reader`이자 `io.Writer`다.** 4장에서 배운 게 여기서 회수된다.

```go
fmt.Fprintf(conn, "%s\n", msg)   // 파일에 쓰듯이 소켓에 쓴다
io.Copy(os.Stdout, conn)         // 소켓을 통째로 표준 출력에 흘린다
```

```go
type Conn interface {
    Read(b []byte) (n int, err error)
    Write(b []byte) (n int, err error)
    Close() error
    LocalAddr() Addr
    RemoteAddr() Addr
    SetDeadline(t time.Time) error       // ← M1의 read/write deadline
    SetReadDeadline(t time.Time) error
    SetWriteDeadline(t time.Time) error
}
```

### 왜 Go 서버가 이렇게 짧은가

```go
for {
    conn, err := ln.Accept()
    if err != nil { return err }
    go handleConn(conn)
}
```

C로 짰다면 `epoll`/`kqueue` 이벤트 루프를 만들고, 논블로킹 소켓을 관리하고, 상태 머신을 돌려야 했다. 스레드가 비싸서 "접속마다 스레드"가 불가능했기 때문이다.

Go는 goroutine이 2KB라서 접속마다 하나씩 띄워도 된다. 그리고 **런타임이 내부적으로 epoll을 써준다** — goroutine이 `Read`에서 멈추면 OS 스레드를 놓아주고 다른 goroutine이 그 스레드를 쓴다. 우리는 동기 코드처럼 짜고, 성능은 이벤트 루프급으로 나온다.

---

## bufio

날 `conn.Read([]byte)`를 직접 쓰면 부분 읽기를 매번 처리해야 한다. `bufio`가 그걸 대신해준다.

```go
r := bufio.NewReader(conn)
line, err := r.ReadString('\n')   // \n을 만날 때까지 계속 읽는다

sc := bufio.NewScanner(conn)
for sc.Scan() {                   // 줄 단위로 잘라준다
    line := sc.Text()             // 개행이 제거된 한 줄
}
```

### ⚠️ 함정 1 — Reader를 연결당 하나만 만든다

```go
for _, msg := range msgs {
    fmt.Fprintf(conn, "%s\n", msg)
    r := bufio.NewReader(conn)      // ✗ 매번 새로 만들면
    line, _ := r.ReadString('\n')   //   앞서 미리 읽어둔 데이터가 버려진다
}
```

`bufio.Reader`는 소켓에서 **한 번에 큰 덩어리를 읽어 자기 버퍼에 쌓아둔다.** 서버가 `"a\nb\nc\n"`를 한 번에 보냈다면, 첫 `ReadString`이 세 줄을 다 읽어와서 `"a\n"`만 돌려주고 `"b\nc\n"`는 버퍼에 남긴다. 여기서 Reader를 버리면 그 데이터가 **영영 사라진다.** 두 번째 응답을 영원히 기다리게 된다.

문제 2의 핵심이 이것이다.

### ⚠️ 함정 2 — Scanner의 64KB 한계

`bufio.Scanner`의 기본 최대 토큰 길이는 **64KB**다. 더 긴 줄이 오면 `Scan()`이 그냥 `false`를 반환하고 루프가 끝난다. **에러도 안 나고 조용히 끝난다.** `sc.Err()`를 확인해야 `bufio.ErrTooLong`이 보인다.

늘리려면:

```go
sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)   // 최대 1MB
```

M1에서 길이 접두사 프레이밍을 직접 만드는 이유 중 하나가 이 한계다. 그리고 **상한 없이 받으면 안 된다** — 악의적인 클라이언트가 개행 없이 10GB를 보내면 서버 메모리가 터진다. M1의 DoD에 "`PAYLOAD_LEN` 상한 검증"이 있는 이유다.

---

## 문제

| # | 대상 | 배우는 것 |
|---|---|---|
| 1 | `Echo` | `net.Dial`, `net.Conn`은 Reader/Writer |
| 2 | `EchoMany` | ★ 연결 재사용, bufio 버퍼 함정 |
| 3 | `handleConn` | ★ `bufio.Scanner`로 메시지 경계 만들기 |
| 4 | `Serve` | ★ accept 루프 + goroutine per conn |

---

## 다 풀었으면 — 진짜로 파이에 붙여보기

여기까지가 M0 DoD다. `cmd/` 아래에 실행 파일이 준비되어 있다.

### 1. 먼저 노트북 안에서

```bash
# 터미널 1
go run ./08_tcp_echo/cmd/echo-server -addr :9000

# 터미널 2
go run ./08_tcp_echo/cmd/echo-client -addr 127.0.0.1:9000 hello world
# → hello
#   world

# 대화형
go run ./08_tcp_echo/cmd/echo-client -addr 127.0.0.1:9000
```

> 아직 문제를 안 풀었어도 동작하는 걸 먼저 보고 싶다면 `./solutions/08_tcp_echo/cmd/...`를 쓰면 된다.

### 2. 파이에서 서버 띄우기

리포지터리를 파이로 가져가서 (파이에도 Go 설치 필요):

```bash
# 파이에서
go run ./08_tcp_echo/cmd/echo-server -addr :9000
```

`:9000`은 **모든 인터페이스에서 듣는다**는 뜻이다. `127.0.0.1:9000`으로 띄우면 파이 안에서만 접속되니 주의.

방화벽(`ufw`)을 켰다면 열어준다:

```bash
sudo ufw allow 9000/tcp
```

### 3. 노트북에서 접속

Tailscale을 깔았다면 파이의 tailnet 주소를 쓴다:

```bash
tailscale status                          # 파이의 100.x.y.z 확인
go run ./08_tcp_echo/cmd/echo-client -addr 100.x.y.z:9000 "안녕 파이"
```

**여기서 응답이 돌아오면 M0의 네트워크 부분은 끝이다.**

### 크로스 컴파일 — 파이에 Go를 안 깔아도 된다

```bash
GOOS=linux GOARCH=arm64 go build -o echo-server ./08_tcp_echo/cmd/echo-server
scp echo-server pi@100.x.y.z:~/
ssh pi@100.x.y.z ./echo-server -addr :9000
```

환경변수 두 개면 끝이다. 이게 순수 Go의 강력한 점이고, **cgo를 쓰는 순간 사라지는 능력**이기도 하다. PLAN.md에서 "`cgofuse`는 cgo라 크로스 컴파일 불가 → FUSE만 별도 모듈로 분리"라고 한 이유가 이것이다.

(파이가 64비트 OS면 `arm64`, 32비트면 `GOARCH=arm GOARM=7`)

---

## 걸려 넘어지기 쉬운 곳

**테스트가 5초 타임아웃으로 실패한다면** — 클라이언트(문제 1, 2)와 서버(문제 3, 4)가 **둘 다** 구현돼야 한다. 서버부터(3, 4) 짜고 클라이언트(1, 2)를 짜는 게 디버깅하기 쉽다.

**`sc.Text()`에는 개행이 없다.** 되돌려 보낼 때 `"\n"`을 다시 붙여야 한다. 안 붙이면 클라이언트의 `ReadString('\n')`이 영원히 기다린다.

**`ReadString('\n')`은 구분자를 포함해서 준다.** `strings.TrimRight(line, "\r\n")`으로 떼어낸다. `\r`까지 떼는 건 윈도우 클라이언트 대비다.

**`defer conn.Close()`를 빼먹으면** 소켓이 샌다. 서버는 결국 `too many open files`로 죽는다. 파일 디스크립터 한도는 보통 256~1024개뿐이다.

**`go handleConn(conn)`에서 `go`를 빼면** 컴파일도 되고 테스트도 대부분 통과한다. 다만 한 번에 접속 하나씩만 처리된다. `TestConcurrentClients`가 이걸 잡는다.

**주소 형식.** `":9000"`은 모든 인터페이스, `"127.0.0.1:9000"`은 로컬만, `"127.0.0.1:0"`은 "빈 포트 아무거나"(테스트에서 쓴다).

---

## 여기서 M1로

M1의 DoD는 이렇다.

> `PING` 100개를 동시에 던져도 응답이 정확히 짝지어진다.
> 랜덤 바이트를 디코더에 넣어도 panic이 안 난다.

이 장에서 만든 echo 서버와 뭐가 다른가:

| | 08장 (지금) | M1 |
|---|---|---|
| 메시지 경계 | `\n` 구분자 | 4바이트 길이 접두사 |
| 데이터 | 텍스트만 (`\n` 못 넣음) | 임의의 바이너리 |
| 크기 제한 | 64KB에서 조용히 끊김 | 명시적 상한 + 에러 |
| 동시 요청 | 보낸 순서대로 하나씩 | `request_id`로 다중화 (7장 `Registry`!) |
| 악의적 입력 | 생각 안 함 | 퍼징으로 검증 |

**7장의 `Registry`와 8장의 `Serve`를 합치면 M1의 뼈대가 된다.**

---

## 더 볼 것

- `go doc net.Conn` / `go doc bufio.Scanner`
- [Go 표준 라이브러리의 net/http 서버 소스](https://github.com/golang/go/blob/master/src/net/http/server.go) — `Serve` 함수를 찾아보면 우리가 짠 것과 구조가 똑같다 (에러 처리와 백오프가 붙었을 뿐)
