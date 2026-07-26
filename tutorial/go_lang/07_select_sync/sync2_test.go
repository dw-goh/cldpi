package conc2

import (
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"
)

func runWithTimeout(t *testing.T, msg string, f func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		f()
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal(msg)
	}
}

// gen은 값들을 순서대로 보내고 닫는 채널을 만듭니다.
func gen(vals ...int) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for _, v := range vals {
			ch <- v
		}
	}()
	return ch
}

func TestRunAll(t *testing.T) {
	const n = 100
	var mu sync.Mutex
	count := 0

	fns := make([]func(), n)
	for i := 0; i < n; i++ {
		fns[i] = func() {
			mu.Lock()
			count++
			mu.Unlock()
		}
	}

	runWithTimeout(t, "RunAll이 끝나지 않습니다. wg.Add / wg.Done 개수가 맞는지 확인하세요", func() {
		RunAll(fns)
	})

	mu.Lock()
	got := count
	mu.Unlock()
	if got != n {
		t.Errorf("실행된 함수 개수 = %d; want %d (RunAll이 전부 끝나기를 기다리지 않았습니다)", got, n)
	}

	runWithTimeout(t, "RunAll(nil)이 끝나지 않습니다", func() { RunAll(nil) })
}

func TestSafeCounter(t *testing.T) {
	var c SafeCounter // 제로값으로 바로 쓸 수 있어야 한다

	const goroutines, each = 50, 200
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < each; j++ {
				c.Inc()
			}
		}()
	}
	wg.Wait()

	if want := goroutines * each; c.Value() != want {
		t.Errorf("Value() = %d; want %d — 증가가 유실됐습니다. Inc에 락이 있는지 확인하세요", c.Value(), want)
	}
}

// 읽기와 쓰기가 동시에 일어나도 레이스가 없어야 합니다.
// 이 테스트는 -race 없이는 거의 항상 통과하므로 반드시 아래로 돌리세요:
//
//	go test -race ./07_select_sync/
func TestSafeCounterRace(t *testing.T) {
	var c SafeCounter
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); c.Inc() }()
		go func() { defer wg.Done(); _ = c.Value() }()
	}
	wg.Wait()
}

func TestFirst(t *testing.T) {
	// a가 먼저 도착.
	a := make(chan string, 1)
	b := make(chan string)
	a <- "a-first"
	var got string
	runWithTimeout(t, "First가 멈췄습니다", func() { got = First(a, b) })
	if got != "a-first" {
		t.Errorf("First = %q; want %q", got, "a-first")
	}

	// b가 먼저 도착.
	a2 := make(chan string)
	b2 := make(chan string, 1)
	b2 <- "b-first"
	runWithTimeout(t, "First가 멈췄습니다", func() { got = First(a2, b2) })
	if got != "b-first" {
		t.Errorf("First = %q; want %q", got, "b-first")
	}

	// 둘 다 늦게 도착 — 멈췄다가 먼저 온 걸 받아야 한다.
	a3 := make(chan string)
	b3 := make(chan string)
	go func() {
		time.Sleep(50 * time.Millisecond)
		a3 <- "slow-a"
	}()
	runWithTimeout(t, "First가 값을 기다리지 않고 바로 반환했거나 멈췄습니다", func() { got = First(a3, b3) })
	if got != "slow-a" {
		t.Errorf("First = %q; want %q", got, "slow-a")
	}
}

func TestRecvTimeout(t *testing.T) {
	// 값이 제때 오는 경우.
	ch := make(chan int, 1)
	ch <- 42
	if v, ok := RecvTimeout(ch, time.Second); v != 42 || !ok {
		t.Errorf("RecvTimeout(값 있음) = %d, %v; want 42, true", v, ok)
	}

	// 아무도 안 보내는 경우 — 타임아웃되어야 한다.
	empty := make(chan int)
	start := time.Now()
	v, ok := RecvTimeout(empty, 100*time.Millisecond)
	elapsed := time.Since(start)

	if v != 0 || ok {
		t.Errorf("RecvTimeout(빈 채널) = %d, %v; want 0, false", v, ok)
	}
	if elapsed < 90*time.Millisecond {
		t.Errorf("%v만에 반환했습니다. 기다리지 않고 바로 포기한 것 같습니다 (default를 쓰지 않았나요?)", elapsed)
	}
	if elapsed > time.Second {
		t.Errorf("%v나 걸렸습니다. 타임아웃이 동작하지 않습니다", elapsed)
	}

	// 타임아웃 직전에 값이 오는 경우.
	late := make(chan int)
	go func() {
		time.Sleep(50 * time.Millisecond)
		late <- 7
	}()
	if v, ok := RecvTimeout(late, time.Second); v != 7 || !ok {
		t.Errorf("RecvTimeout(늦게 옴) = %d, %v; want 7, true", v, ok)
	}
}

func TestTryRecv(t *testing.T) {
	ch := make(chan int, 1)

	// 비어 있으면 즉시 false.
	start := time.Now()
	if v, ok := TryRecv(ch); v != 0 || ok {
		t.Errorf("TryRecv(빈 채널) = %d, %v; want 0, false", v, ok)
	}
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Errorf("TryRecv가 %v 멈췄습니다. default가 필요합니다", elapsed)
	}

	// 값이 있으면 그걸 준다.
	ch <- 9
	if v, ok := TryRecv(ch); v != 9 || !ok {
		t.Errorf("TryRecv(값 있음) = %d, %v; want 9, true", v, ok)
	}
	// 꺼냈으니 다시 비어야 한다.
	if _, ok := TryRecv(ch); ok {
		t.Error("TryRecv를 두 번 불렀는데 두 번 다 값이 나왔습니다")
	}
}

func TestMerge(t *testing.T) {
	out := Merge(gen(1, 2, 3), gen(4, 5, 6))

	var got []int
	done := make(chan struct{})
	go func() {
		defer close(done)
		for v := range out {
			got = append(got, v)
		}
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Merge가 반환한 채널이 닫히지 않습니다. 두 입력이 모두 닫힌 뒤에 close해야 합니다")
	}

	sort.Ints(got)
	want := []int{1, 2, 3, 4, 5, 6}
	if len(got) != len(want) {
		t.Fatalf("Merge = %v; want %v (순서는 상관없음)", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Merge = %v; want %v (순서는 상관없음)", got, want)
		}
	}

	// 한쪽이 비어 있어도 동작해야 한다.
	out2 := Merge(gen(), gen(1))
	got2 := []int{}
	done2 := make(chan struct{})
	go func() {
		defer close(done2)
		for v := range out2 {
			got2 = append(got2, v)
		}
	}()
	select {
	case <-done2:
	case <-time.After(3 * time.Second):
		t.Fatal("한쪽이 빈 경우 Merge 채널이 닫히지 않습니다")
	}
	if len(got2) != 1 || got2[0] != 1 {
		t.Errorf("Merge(빈 채널, [1]) = %v; want [1]", got2)
	}
}

func TestRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("NewRegistry() = nil")
	}

	ch1 := r.Wait(1)
	ch2 := r.Wait(2)

	// 등록 순서와 다르게 배달해도 각자 자기 것을 받아야 한다 — 이게 다중화다.
	if !r.Deliver(2, "resp-2") {
		t.Error("Deliver(2, ...) = false; want true")
	}
	if !r.Deliver(1, "resp-1") {
		t.Error("Deliver(1, ...) = false; want true")
	}

	if got := <-ch1; got != "resp-1" {
		t.Errorf("ch1 = %q; want %q", got, "resp-1")
	}
	if got := <-ch2; got != "resp-2" {
		t.Errorf("ch2 = %q; want %q", got, "resp-2")
	}

	// 없는 id.
	if r.Deliver(99, "nobody") {
		t.Error("Deliver(등록 안 된 id) = true; want false")
	}
	// 이미 배달된 id — 등록이 지워졌어야 한다.
	if r.Deliver(1, "again") {
		t.Error("Deliver(이미 배달된 id) = true; want false")
	}
}

// 동시에 등록하고 동시에 배달해도 뒤섞이지 않아야 합니다.
// 반드시 -race로 돌려보세요.
func TestRegistryConcurrent(t *testing.T) {
	r := NewRegistry()
	const n = 200

	chans := make([]<-chan string, n)
	for i := 0; i < n; i++ {
		chans[i] = r.Wait(i)
	}

	// 배달하는 쪽을 여러 goroutine으로 흩뿌린다.
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if !r.Deliver(i, fmt.Sprintf("resp-%d", i)) {
				t.Errorf("Deliver(%d) = false; want true", i)
			}
		}()
	}
	wg.Wait()

	// 각자 자기 응답을 정확히 받아야 한다.
	for i := 0; i < n; i++ {
		select {
		case got := <-chans[i]:
			if want := fmt.Sprintf("resp-%d", i); got != want {
				t.Errorf("id %d의 채널에서 %q를 받았습니다; want %q — 응답이 엉뚱한 대기자에게 갔습니다", i, got, want)
			}
		case <-time.After(3 * time.Second):
			t.Fatalf("id %d가 응답을 받지 못했습니다", i)
		}
	}
}
