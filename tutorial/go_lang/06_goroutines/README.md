# 06 — goroutine과 channel

Go를 쓰는 가장 큰 이유. M0의 "goroutine per connection"이 이 장의 내용이다.

```bash
go test -v ./06_goroutines/
go test -race ./06_goroutines/    # ★ 동시성 코드는 항상 이것도 돌린다
```

---

## goroutine

```go
go doSomething()
go func() { ... }()
```

`go`를 붙이면 새 goroutine에서 실행되고, 호출한 쪽은 기다리지 않고 다음 줄로 간다.

OS 스레드가 아니다. **런타임이 관리하는 초경량 스레드**다.

| | OS 스레드 | goroutine |
|---|---|---|
| 초기 스택 | 1~8MB (고정) | 2KB (필요하면 자동으로 늘어남) |
| 생성 비용 | 수십 μs | 수백 ns |
| 만들 수 있는 개수 | 수천 개 | **수십만 개** |
| 전환 | 커널 스케줄러 | Go 런타임 (커널 안 거침) |

그래서 Go에서는 "접속마다 스레드 하나"가 사치가 아니다. C로 서버를 짤 때 `epoll`로 이벤트 루프를 짜야 했던 이유가 스레드가 비쌌기 때문인데, Go는 그냥 goroutine을 띄우면 된다. **런타임이 내부적으로 epoll/kqueue를 써주고, 우리는 동기 코드처럼 짜면 된다.**

```go
for {
    conn, err := ln.Accept()
    if err != nil { return err }
    go handleConn(conn)      // ← 8장에서 이렇게 쓴다
}
```

### 두 가지 주의

**1. main이 끝나면 모든 goroutine이 즉시 죽는다.** 기다려주지 않는다.

**2. goroutine 안의 panic은 프로세스 전체를 죽인다.** 다른 goroutine이 잡을 수 없다. 5장에서 본 `defer recover()` 패턴이 필요한 이유다.

---

## 채널

```go
ch := make(chan int)       // 버퍼 없음 (unbuffered)
ch := make(chan int, 10)   // 버퍼 10개

ch <- 42                   // 보내기
v := <-ch                  // 받기
v, ok := <-ch              // 받기 + 채널이 닫혔는지
close(ch)                  // 닫기
for v := range ch { }      // 닫힐 때까지 계속 받기
```

**버퍼 없는 채널**은 만나는 지점이다. 보내는 쪽은 받는 쪽이 준비될 때까지 멈추고, 받는 쪽도 마찬가지다. 이 "만남"이 곧 동기화다.

**버퍼 있는 채널**은 버퍼가 찰 때까지는 안 멈춘다. 가득 차면 그때부터 멈춘다.

### 채널 규칙 ★

| 상황 | 결과 |
|---|---|
| 닫힌 채널에서 받기 | 즉시 (제로값, false). **멈추지 않는다** |
| 닫힌 채널에 보내기 | **panic** |
| 닫힌 채널을 또 닫기 | **panic** |
| nil 채널에 보내기/받기 | **영원히 멈춤** |
| nil 채널을 닫기 | **panic** |

**close는 보내는 쪽이 한다.** 받는 쪽은 언제 더 안 올지 알 수 없다. 그리고 close는 "더 보낼 게 없다"는 신호이지 자원 해제가 아니다. 받는 쪽이 `range`로 도는 게 아니라면 닫지 않아도 GC가 알아서 회수한다.

### 방향 있는 채널 타입

```go
func Generator(n int) <-chan int    // 받기 전용을 반환
func consume(ch <-chan int)         // 이 함수는 받기만 한다
func produce(ch chan<- int)         // 이 함수는 보내기만 한다
```

화살표가 `chan` 왼쪽이면 받기, 오른쪽이면 보내기. 함수 시그니처에 쓰면 **실수로 반대 방향 연산을 하는 걸 컴파일러가 막아준다.** 문서 역할도 한다.

---

## "통신해서 메모리를 공유하라"

> Do not communicate by sharing memory; instead, share memory by communicating.

여러 goroutine이 같은 변수를 건드리면 **데이터 레이스**다.

```go
total := 0
for _, n := range nums {
    go func() { total += n }()   // ✗ 레이스. 결과가 매번 다르다
}
```

`total += n`은 원자적이지 않다(읽기 → 더하기 → 쓰기). C에서 pthread로 겪던 것과 똑같은 문제다.

채널로 결과를 모으면 이 문제가 없다:

```go
results := make(chan int)
for _, n := range nums {
    go func() { results <- n }()  // 각자 자기 값만 보낸다
}
total := 0
for range nums {
    total += <-results            // 더하는 건 이 goroutine 혼자
}
```

**맵은 특히 위험하다.** 여러 goroutine이 동시에 맵에 쓰면 Go 런타임이 감지해서 `fatal error: concurrent map writes`로 프로그램을 죽인다. (뮤텍스는 다음 장에서 다룬다.)

### `-race` 플래그

```bash
go test -race ./...
go run -race main.go
```

실행하면서 실제로 일어난 메모리 접근을 추적해 레이스를 잡아준다. **켜고 안 걸렸다고 없는 게 아니라, 그 실행 경로에서 안 걸린 것뿐**이지만, 그래도 안 쓰는 것보다 압도적으로 낫다. 느려지므로(2~10배) 테스트에서만 켠다.

---

## 문제

| # | 함수 | 배우는 것 |
|---|---|---|
| 1 | `Ping` | `go` + 버퍼 없는 채널 |
| 2 | `Generator` | 채널 반환, `close`, 방향 있는 채널 |
| 3 | `Square` | 파이프라인 |
| 4 | `SumConcurrent` | 결과를 채널로 모으기 |
| 5 | `WorkerPool` | ★ jobs/results 패턴 — 8장 서버의 원형 |
| 6 | `RecvOne` | `v, ok := <-ch` |

---

## 걸려 넘어지기 쉬운 곳

**테스트가 안 끝나고 멈춰 있다면** 거의 항상 둘 중 하나다.

1. **`close`를 안 했다.** `range ch`는 채널이 닫혀야 끝난다.
2. **보낸 개수와 받는 개수가 안 맞는다.** 버퍼 없는 채널에 3개 보내는데 2개만 받으면 세 번째 보내기에서 영원히 멈춘다.

Ctrl-C로 죽이거나 기다리면 Go가 **모든 goroutine의 스택을 덤프**해준다. `chan send`나 `chan receive`에서 멈춰 있는 줄 번호를 보면 어디가 문제인지 바로 나온다.

```
goroutine 1 [chan receive]:
	cldpi/gotour/06_goroutines.SumConcurrent(...)
		/Users/.../conc.go:87
```

> 이 튜토리얼의 테스트에는 2초 타임아웃이 걸려 있어서, 대부분은 10분을 기다리지 않고 친절한 메시지로 실패한다.

**문제 2 — `Generator`는 즉시 반환해야 한다.**

```go
func Generator(n int) <-chan int {
    ch := make(chan int)
    for i := 0; i < n; i++ { ch <- i }   // ✗ 첫 번째 보내기에서 영원히 멈춤
    close(ch)
    return ch                            //   여기 도달 못 한다
}
```

버퍼 없는 채널이라 받는 사람이 없으면 첫 `ch <- 0`에서 막힌다. 그런데 받는 사람은 이 함수가 반환해야 생긴다. 전형적인 self-deadlock이다. 보내는 부분을 `go func() { ... }()`로 감싸야 한다.

**문제 5 — 맵에 여러 goroutine이 쓰면 안 된다.**
워커들은 결과를 채널로 보내고, 맵에 넣는 건 메인 goroutine 하나만 해야 한다. `-race`로 돌리면 잡힌다.

**루프 변수 캡처 — 이제는 안전하다.**
예전 Go 튜토리얼에서 이런 경고를 봤을 것이다.

```go
for _, v := range items {
    go func() { use(v) }()   // 옛날엔 모든 goroutine이 마지막 v를 봤다
}
```

**Go 1.22부터 루프 변수는 반복마다 새로 만들어진다.** 이 버그는 이제 안 난다. 인터넷의 오래된 자료에서 `go func(v int){...}(v)` 같은 코드를 보면 그 시절 회피책이라고 이해하면 된다. (`go.mod`의 `go` 버전이 1.22 이상이어야 적용된다 — 이 튜토리얼은 1.26이다.)

---

## 더 볼 것

- [Go Concurrency Patterns](https://go.dev/blog/pipelines) — 파이프라인과 취소
- [Effective Go — Concurrency](https://go.dev/doc/effective_go#concurrency)
- [Go Memory Model](https://go.dev/ref/mem) — 나중에. 뭐가 보장되는지 정확히 알고 싶을 때
