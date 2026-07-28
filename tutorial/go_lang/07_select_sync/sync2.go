// Package conc2 — select와 sync.
//
//	go test ./07_select_sync/
//	go test -race ./07_select_sync/    ← 이 장은 -race가 특히 중요합니다
//
// 마지막 문제(Registry)는 M1의 "요청 다중화"를 미리 만들어보는 것입니다.
package conc2

import (
	"sync"
	"time"
)

// 문제 1 — WaitGroup
//
// fns의 함수들을 각각 goroutine으로 실행하고, **전부 끝난 뒤에** 반환하세요.
//
//	var wg sync.WaitGroup
//	wg.Add(1)              // 기다릴 개수를 늘린다
//	go func() {
//	    defer wg.Done()    // 끝나면 하나 줄인다. defer로 쓰는 게 안전하다
//	    ...
//	}()
//	wg.Wait()              // 0이 될 때까지 기다린다
//
// ★ wg.Add는 반드시 go 하기 **전에** 불러야 합니다.
// goroutine 안에서 Add를 부르면 Wait이 먼저 실행돼 그냥 지나가버릴 수 있습니다.
//
// var wg sync.WaitGroup 처럼 선언만 해도 바로 쓸 수 있습니다 (3장의 "쓸모있는 제로값").
// 그리고 WaitGroup은 절대 복사하면 안 됩니다 — 항상 &wg로 넘깁니다. go vet이 잡아줍니다.
func RunAll(fns []func()) {
	// TODO
        var wg sync.WaitGroup
        wg.Add(len(fns))
        
        for _, f := range fns {
            go func() {
                defer wg.Done()
                f()
            }()   
        }
        wg.Wait()
}

// SafeCounter는 여러 goroutine이 동시에 써도 안전한 카운터입니다.
type SafeCounter struct {
	mu sync.Mutex
	n  int
}

// 문제 2 — Mutex ★
//
// Inc는 n을 1 증가시키고, Value는 현재 값을 반환합니다.
// 둘 다 여러 goroutine이 동시에 불러도 안전해야 합니다.
//
//	c.mu.Lock()
//	defer c.mu.Unlock()
//
// n++ 는 원자적이지 않습니다 (읽기 → 더하기 → 쓰기). 락이 없으면
// 두 goroutine이 같은 값을 읽고 같은 값을 써서 증가 한 번이 사라집니다.
//
// ★ Value에도 락이 필요합니다. 읽기만 해도 다른 goroutine의 쓰기와 겹치면
// 데이터 레이스입니다. (읽기가 압도적으로 많으면 sync.RWMutex를 씁니다)
//
// 6장에서 본 "채널로 통신하라"와 뮤텍스 중 뭘 쓰나:
// 값을 넘겨주는 흐름이면 채널, 상태를 보호하는 것이면 뮤텍스가 보통 더 간단합니다.
func (c *SafeCounter) Inc() {
	// TODO
        c.mu.Lock()
        defer c.mu.Unlock()
        c.n++
}

func (c *SafeCounter) Value() int {
	// TODO
        c.mu.Lock()
        defer c.mu.Unlock()
	return c.n
}

// 문제 3 — select
//
// a와 b 중 **먼저 도착한** 값을 반환하세요.
//
//	select {
//	case v := <-a:
//	    return v
//	case v := <-b:
//	    return v
//	}
//
// select는 여러 채널 연산 중 준비된 것 하나를 고릅니다.
// 둘 다 준비돼 있으면 **무작위로** 하나를 고릅니다 (한쪽만 계속 처리하는 걸 막으려고).
// 아무것도 준비 안 됐으면 하나가 준비될 때까지 멈춥니다.
func First(a, b <-chan string) string {
	// TODO
        select {
        case v := <-a:
            return v
        case v := <-b:
            return v
        }
}

// 문제 4 — 타임아웃 ★
//
// ch에서 값을 받되, d 시간이 지나면 포기하고 (0, false)를 반환하세요.
//
//	select {
//	case v := <-ch:
//	    return v, true
//	case <-time.After(d):
//	    return 0, false
//	}
//
// time.After(d)는 d 후에 값 하나가 나오는 채널을 반환합니다.
// "타임아웃"이라는 별도 개념 없이, 그냥 또 하나의 채널로 표현됩니다.
//
// M1에서 read/write deadline을 다룰 때 같은 발상이 나옵니다.
// (실제 네트워크 코드에서는 conn.SetReadDeadline이나 context를 더 자주 씁니다)
func RecvTimeout(ch <-chan int, d time.Duration) (int, bool) {
	// TODO
        select {
        case v := <-ch:
            return v, true
        case <- time.After(d):
            return 0, false
        }
}

// 문제 5 — 논블로킹 수신
//
// ch에 지금 당장 받을 값이 있으면 (값, true), 없으면 즉시 (0, false)를 반환하세요.
// **절대 멈추면 안 됩니다.**
//
//	select {
//	case v := <-ch:
//	    return v, true
//	default:
//	    return 0, false
//	}
//
// default가 있으면 준비된 case가 없을 때 바로 default로 갑니다.
// 보내기도 같은 방식으로 논블로킹으로 만들 수 있습니다 —
// 로그를 흘려보내되 버퍼가 차면 그냥 버리는 식으로 자주 씁니다.
func TryRecv(ch <-chan int) (int, bool) {
	// TODO
        select {
        case v := <-ch:
            return v, true
        default:
            return 0, false
        }
}

// 문제 6 — 팬인(fan-in)
//
// a와 b에서 나오는 값을 하나의 채널로 합쳐서 반환하세요.
// **a와 b가 둘 다 닫히면** 반환한 채널도 닫아야 합니다.
// (순서는 상관없습니다)
//
// 힌트: a를 읽는 goroutine과 b를 읽는 goroutine을 각각 띄우고,
// WaitGroup으로 둘 다 끝나기를 기다렸다가 닫는 goroutine을 하나 더 띄웁니다.
//
//	go func() { wg.Wait(); close(out) }()
//
// select로도 풀 수 있지만 한쪽만 먼저 닫혔을 때 처리가 까다롭습니다.
// 닫힌 채널은 계속 준비 상태라 select가 그쪽만 무한히 고르기 때문에,
// 해당 변수에 nil을 넣어 그 case를 영구히 꺼버리는 기법을 써야 합니다.
func Merge(a, b <-chan int) <-chan int {
	// TODO
        var wg sync.WaitGroup
        wg.Add(2)

        ch := make(chan int)
        go func() {
            for v := range a { ch <- v }
            wg.Done()
        }()
        go func() {
            for v := range b { ch <- v }
            wg.Done()
        }()
        go func() {
            wg.Wait()
            close(ch)
        }()
	return ch
}

// Registry는 id별로 응답을 기다리는 사람들을 관리합니다. ★★
//
// M1의 "요청 다중화"가 정확히 이 구조입니다:
// 하나의 TCP 연결로 요청 여러 개를 동시에 보내고, 응답이 뒤섞여 돌아와도
// request_id를 보고 올바른 대기자에게 배달해야 합니다.
//
//	요청 3개를 보냄 → 응답이 2, 1, 3 순서로 돌아옴 → 각자 자기 것을 받아야 함
type Registry struct {
	mu      sync.Mutex
	waiters map[int]chan string
}

// 문제 7 — 생성자
//
// waiters 맵을 초기화한 Registry를 반환하세요.
// (nil 맵에 쓰면 panic이므로 반드시 make가 필요합니다)
func NewRegistry() *Registry {
        return &Registry{
            waiters: make(map[int]chan string),
        }
}

// 문제 7 (계속) — Wait
//
// id로 응답을 기다리겠다고 등록하고, 응답이 올 채널을 반환하세요.
//
// ★ 채널은 반드시 **버퍼 1**로 만드세요: make(chan string, 1)
// 그래야 Deliver가 받는 사람을 기다리지 않고 즉시 보내고 빠져나올 수 있습니다.
// 버퍼가 없으면 Deliver가 락을 잡은 채 멈춰서 전체가 굳을 수 있습니다.
func (r *Registry) Wait(id int) <-chan string {
	// TODO
        ch := make(chan string, 1)
        
        r.mu.Lock()
        defer r.mu.Unlock()
        r.waiters[id] = ch

	return ch
}

// 문제 7 (계속) — Deliver
//
// id로 기다리는 사람이 있으면 msg를 보내고 등록을 지운 뒤 true를 반환하세요.
// 없으면 아무것도 하지 않고 false를 반환합니다.
//
// 같은 id로 두 번 Deliver하면 두 번째는 false여야 합니다.
//
// ★ 락을 잡은 상태에서 채널에 보내지 마세요.
// 맵에서 꺼내고 지우는 것까지만 락 안에서 하고, 락을 푼 뒤에 보냅니다.
// (락을 잡은 채로 멈출 수 있는 연산을 하는 건 데드락의 지름길입니다)
func (r *Registry) Deliver(id int, msg string) bool {
        r.mu.Lock()
        ch, ok := r.waiters[id]
        if ok {
            delete(r.waiters, id)
        }
        r.mu.Unlock()

        if !ok {
            return false
        }

        ch <- msg
        return true
}
