# 07 — select와 sync

6장이 "goroutine을 만드는 법"이었다면 이 장은 **"제어하는 법"**이다. 기다리기, 타임아웃, 포기하기, 공유 상태 지키기.

```bash
go test -v ./07_select_sync/
go test -race ./07_select_sync/    # ★ 이 장은 반드시
```

---

## select

```go
select {
case v := <-a:
    // a에서 받았다
case b <- 1:
    // b로 보냈다
case <-time.After(time.Second):
    // 1초가 지났다
default:
    // 위 어느 것도 지금 당장 준비되지 않았다
}
```

`switch`처럼 생겼지만 **채널 연산 전용**이다. 동작 규칙:

1. 준비된 case가 하나면 그걸 실행한다
2. 여러 개면 **무작위로** 하나를 고른다 (한쪽만 계속 먹는 걸 막으려고)
3. 하나도 없으면 → `default`가 있으면 그리로, 없으면 **하나가 준비될 때까지 기다린다**

### 타임아웃이 별도 개념이 아니다

```go
select {
case v := <-ch:
    return v, true
case <-time.After(2 * time.Second):
    return 0, false
}
```

`time.After(d)`는 "d 후에 값 하나가 나오는 채널"을 반환한다. 타임아웃도 그냥 또 하나의 채널이다. 이 발상이 Go 동시성의 핵심 미학이다.

> 실무 팁: 루프 안에서 `time.After`를 쓰면 매번 타이머가 새로 만들어져 GC될 때까지 남는다. 자주 도는 루프에서는 `time.NewTimer`를 만들어 재사용하거나 `context.WithTimeout`을 쓴다. (M1에서 다시 만난다.)

### 논블로킹 연산

```go
select {
case v := <-ch:
    // 받았다
default:
    // 지금은 값이 없다. 기다리지 않고 넘어간다
}
```

`default`가 있으면 절대 멈추지 않는다. 보내기도 마찬가지다 — 버퍼가 차면 그냥 버리는 로그 채널 같은 데 쓴다.

### 함정: 닫힌 채널은 영원히 "준비 상태"다

```go
for {
    select {
    case v, ok := <-a:
        if !ok { a = nil }   // ← 이렇게 막는다
        ...
    case v, ok := <-b:
        ...
    }
}
```

닫힌 채널은 계속 즉시 값(제로값)을 내주므로, select가 그 case만 무한히 고른다. **nil 채널은 영원히 준비되지 않으므로** 변수에 `nil`을 넣으면 그 case가 사실상 꺼진다. 관용구로 알아두자. (문제 6은 WaitGroup으로 더 쉽게 풀 수 있다.)

---

## sync

### WaitGroup — "다 끝날 때까지 기다려"

```go
var wg sync.WaitGroup
for _, item := range items {
    wg.Add(1)                    // ★ go 하기 전에
    go func() {
        defer wg.Done()          // ★ defer로
        process(item)
    }()
}
wg.Wait()
```

세 가지만 지키면 된다.

1. **`Add`는 `go` 하기 전에.** goroutine 안에서 `Add`하면 `Wait`이 먼저 실행돼 그냥 통과해버릴 수 있다.
2. **`Done`은 `defer`로.** panic이 나도 카운트가 줄어야 한다. 안 그러면 `Wait`이 영원히 안 끝난다.
3. **복사하지 않는다.** 항상 `&wg`로 넘긴다. `go vet`이 잡아준다.

### Mutex — "한 번에 하나만"

```go
type SafeCounter struct {
    mu sync.Mutex
    n  int
}

func (c *SafeCounter) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.n++
}
```

**읽기에도 락이 필요하다.** "읽기만 하는데 뭐 어때"가 아니다. 다른 goroutine의 쓰기와 겹치면 그 자체로 데이터 레이스이고, 컴파일러 최적화나 CPU 재배치 때문에 실제로 이상한 값을 볼 수 있다.

읽기가 압도적으로 많으면 `sync.RWMutex`를 쓴다 (`RLock`은 여러 개가 동시에 잡을 수 있다).

**뮤텍스를 필드로 가진 구조체는 복사하면 안 된다.** 그래서 메서드 리시버가 전부 포인터여야 한다. `go vet`이 "passes lock by value"로 잡아준다.

### 채널이냐 뮤텍스냐

| 쓰는 상황 | 도구 |
|---|---|
| 값을 넘겨주는 흐름, 파이프라인, 작업 분배 | **채널** |
| 그냥 상태 하나를 여럿이 건드림 (카운터, 캐시, 맵) | **뮤텍스** |

"채널이 Go스러우니까 무조건 채널"은 틀렸다. 공식 위키에도 *"Use whichever is most expressive and/or most simple"*이라고 적혀 있다. 캐시에 뮤텍스를 쓰는 게 채널로 감싸는 것보다 훨씬 짧고 명확하다.

---

## `-race`가 진짜로 중요한 이유

레이스가 있는 코드는 **평소엔 잘 돌아간다.** 부하가 걸리거나, 다른 하드웨어에서, 몇 주 뒤에 터진다. 파이(ARM)는 맥북(x86/ARM)과 메모리 모델이 미묘하게 달라서 "내 노트북에선 되는데" 상황이 실제로 나온다.

```bash
go test -race ./...
```

레이스가 잡히면 이런 리포트가 나온다 — **두 goroutine이 어디서 같은 주소를 건드렸는지** 스택 트레이스까지 다 알려준다.

```
WARNING: DATA RACE
Write at 0x00c000112233 by goroutine 8:
  cldpi/gotour/07_select_sync.(*SafeCounter).Inc()
      sync2.go:42 +0x3c
Previous read at 0x00c000112233 by goroutine 7:
  ...
```

---

## 문제

| # | 대상 | 배우는 것 |
|---|---|---|
| 1 | `RunAll` | `sync.WaitGroup` |
| 2 | `SafeCounter` | ★ `sync.Mutex`, `-race` |
| 3 | `First` | `select` |
| 4 | `RecvTimeout` | ★ `select` + `time.After` |
| 5 | `TryRecv` | `select` + `default` |
| 6 | `Merge` | 팬인, "둘 다 끝나면 닫기" |
| 7 | `Registry` | ★★ **M1 요청 다중화의 원형** |

---

## 문제 7이 M1에서 하는 일

M1의 DoD는 이것이다.

> `PING` 100개를 동시에 던져도 응답이 정확히 짝지어진다.

TCP 연결 하나로 요청 여러 개를 동시에 보내면 응답은 **보낸 순서대로 오지 않는다.** 서버가 3번 요청을 먼저 처리했으면 3번 응답이 먼저 온다. 그래서 각 요청에 `request_id`를 붙이고, 응답이 오면 그 id의 주인에게 배달해야 한다.

```
클라이언트                              서버
  │  req(id=1) ─────────────────────────▶
  │  req(id=2) ─────────────────────────▶
  │  req(id=3) ─────────────────────────▶
  │                                       (2번이 제일 빨리 끝남)
  │  ◀──────────────────────── resp(id=2)
  │  ◀──────────────────────── resp(id=3)
  │  ◀──────────────────────── resp(id=1)
  │
  └─ 읽기 goroutine 하나가 전부 받아서
     waiters[id]로 각자에게 배달한다        ← 이게 Registry
```

`Registry.Wait`가 요청을 보내기 전에 부르는 것, `Deliver`가 읽기 goroutine이 응답을 받아 부르는 것이다. 문제 7을 풀면 M1의 절반은 이미 이해한 셈이다.

---

## 걸려 넘어지기 쉬운 곳

**문제 7 — 락을 잡은 채로 채널에 보내지 마라.**

```go
r.mu.Lock()
defer r.mu.Unlock()
r.waiters[id] <- msg     // ✗ 받는 사람이 없으면 락을 쥔 채 굳는다
```

맵에서 꺼내고 지우는 것까지만 락 안에서 하고, **락을 푼 뒤에** 보낸다. 그리고 채널을 버퍼 1로 만들면 받는 사람이 아직 안 왔어도 보내기가 즉시 끝난다.

일반 규칙: **락을 쥔 상태에서 멈출 수 있는 연산(채널, I/O, 다른 락)을 하지 않는다.**

**문제 1 — `wg.Add(1)`을 goroutine 안에 넣으면 안 된다.**

```go
for _, f := range fns {
    go func() {
        wg.Add(1)      // ✗ Wait이 먼저 실행되면 0이라 그냥 통과한다
        defer wg.Done()
        f()
    }()
}
wg.Wait()
```

**문제 2 — `Value()`에도 락을 걸어야 한다.** 안 걸어도 테스트는 통과하지만 `-race`가 잡는다.

---

## 더 볼 것

- [Go Concurrency Patterns: Context](https://go.dev/blog/context) — 취소와 타임아웃의 정석. M1~M3에서 쓴다
- `go doc sync` / `go doc sync/atomic`
- `golang.org/x/sync/errgroup` — WaitGroup + 에러 전파. M3의 청크 병렬 전송에서 쓴다
