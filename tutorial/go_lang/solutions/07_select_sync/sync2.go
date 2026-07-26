// Package conc2 — 07_select_sync 정답.
package conc2

import (
	"sync"
	"time"
)

// 문제 1 — WaitGroup
func RunAll(fns []func()) {
	var wg sync.WaitGroup
	for _, f := range fns {
		// Add는 반드시 go 하기 전에. goroutine 안에서 부르면
		// Wait이 먼저 실행돼 카운트가 0인 채로 통과할 수 있다.
		wg.Add(1)
		go func() {
			// defer로 걸어야 f()가 panic해도 카운트가 줄어든다.
			// 안 그러면 Wait이 영원히 안 끝난다.
			defer wg.Done()
			f()
		}()
	}
	wg.Wait()
	// wg는 절대 복사하면 안 된다. 함수에 넘길 일이 있으면 *sync.WaitGroup으로 넘긴다.
	// go vet이 "passes lock by value"로 잡아준다.
}

type SafeCounter struct {
	mu sync.Mutex
	n  int
}

// 문제 2 — Mutex ★
func (c *SafeCounter) Inc() {
	c.mu.Lock()
	// defer로 풀면 중간에 return하거나 panic해도 락이 풀린다.
	// 락 구간이 아주 짧고 성능이 중요한 경우에만 defer 없이 직접 Unlock한다.
	defer c.mu.Unlock()

	// n++ 는 원자적이지 않다: 읽기 → 더하기 → 쓰기 세 단계다.
	// 락이 없으면 두 goroutine이 같은 값을 읽고 같은 값을 써서 증가 하나가 사라진다.
	c.n++
}

func (c *SafeCounter) Value() int {
	// 읽기에도 락이 필요하다. 다른 goroutine의 쓰기와 겹치는 것 자체가
	// 데이터 레이스이고, -race가 잡아낸다.
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}

// 문제 3 — select
func First(a, b <-chan string) string {
	select {
	case v := <-a:
		return v
	case v := <-b:
		return v
	}
	// 둘 다 준비돼 있으면 무작위로 하나를 고른다.
	// 위에 쓴 case가 우선순위를 갖지 않는다는 점에 주의.
}

// 문제 4 — 타임아웃 ★
func RecvTimeout(ch <-chan int, d time.Duration) (int, bool) {
	select {
	case v := <-ch:
		return v, true
	case <-time.After(d):
		// time.After는 d 후에 값 하나가 나오는 채널을 반환한다.
		// 타임아웃이 특별한 문법이 아니라 그냥 또 하나의 채널이라는 게 핵심이다.
		return 0, false
	}

	// 실무 주의: 자주 도는 루프 안에서 time.After를 쓰면 타이머가 매번 새로 만들어져
	// d가 지날 때까지 GC되지 않는다. 그럴 땐 time.NewTimer를 만들어 Reset하거나
	// context.WithTimeout을 쓴다.
}

// 문제 5 — 논블로킹 수신
func TryRecv(ch <-chan int) (int, bool) {
	select {
	case v := <-ch:
		return v, true
	default:
		// default가 있으면 준비된 case가 없을 때 즉시 여기로 온다.
		// 절대 멈추지 않는다.
		return 0, false
	}
}

// 문제 6 — 팬인(fan-in)
func Merge(a, b <-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup

	// 입력 채널 하나를 out으로 흘려보내는 공통 함수.
	forward := func(in <-chan int) {
		defer wg.Done()
		for v := range in {
			out <- v
		}
	}

	wg.Add(2)
	go forward(a)
	go forward(b)

	// 닫는 일만 하는 goroutine을 따로 둔다.
	// forward 안에서 close(out)을 부르면 두 번 닫혀서 panic이고,
	// 여기서 wg.Wait()을 동기적으로 부르면 out을 아무도 안 읽어서 데드락이다.
	go func() {
		wg.Wait()
		close(out)
	}()

	return out

	// select로도 풀 수 있지만, 한쪽이 먼저 닫히면 그 case가 영원히 준비 상태가 되어
	// select가 그것만 무한히 고른다. 그때는 변수에 nil을 넣어 case를 꺼야 한다:
	//
	//	case v, ok := <-a:
	//	    if !ok { a = nil; continue }   // nil 채널은 영원히 준비되지 않는다
}

type Registry struct {
	mu      sync.Mutex
	waiters map[int]chan string
}

// 문제 7 — 생성자
func NewRegistry() *Registry {
	// nil 맵에 쓰면 panic이므로 make가 반드시 필요하다.
	// 이게 "쓸모있는 제로값"을 만들 수 없어서 생성자를 두는 전형적인 경우다.
	return &Registry{
		waiters: make(map[int]chan string),
	}
}

// 문제 7 (계속) — Wait
func (r *Registry) Wait(id int) <-chan string {
	// 버퍼 1이 핵심이다. Deliver가 받는 사람을 기다리지 않고
	// 값을 넣고 바로 빠져나올 수 있다.
	ch := make(chan string, 1)

	r.mu.Lock()
	defer r.mu.Unlock()
	r.waiters[id] = ch

	// chan string을 <-chan string으로 반환하면 호출한 쪽은 받기만 할 수 있다.
	return ch
}

// 문제 7 (계속) — Deliver
func (r *Registry) Deliver(id int, msg string) bool {
	r.mu.Lock()
	ch, ok := r.waiters[id]
	if ok {
		// 지워야 같은 id로 두 번 배달되지 않는다.
		delete(r.waiters, id)
	}
	r.mu.Unlock() // ★ 채널에 보내기 전에 락을 푼다

	if !ok {
		return false
	}

	// 락 바깥에서 보낸다. 버퍼가 1이고 id당 한 번만 배달되므로 여기서 멈추지 않는다.
	//
	// 락을 쥔 채로 보냈다면, 받는 사람이 늦을 때 락이 잡힌 채로 굳어서
	// 다른 모든 Wait/Deliver가 멈춘다. 락을 쥐고 멈출 수 있는 연산
	// (채널, I/O, 다른 락)을 하지 않는 것이 일반 규칙이다.
	ch <- msg
	return true
}
