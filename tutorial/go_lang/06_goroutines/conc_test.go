package conc

import (
	"reflect"
	"testing"
	"time"
)

// runWithTimeout은 f를 goroutine에서 실행하고, 제한 시간 안에 끝나지 않으면
// 테스트를 실패시킵니다. 동시성 코드는 틀리면 "실패"가 아니라 "영원히 멈춤"이라
// 이런 안전장치가 필요합니다.
func runWithTimeout(t *testing.T, msg string, f func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		f()
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal(msg)
	}
}

// collect는 채널이 닫힐 때까지 값을 모읍니다.
func collect(t *testing.T, ch <-chan int) []int {
	t.Helper()
	var out []int
	timeout := time.After(2 * time.Second)
	for {
		select {
		case v, ok := <-ch:
			if !ok {
				return out
			}
			out = append(out, v)
		case <-timeout:
			t.Fatalf("2초 안에 채널이 닫히지 않았습니다. 보내는 쪽이 close(ch)를 부르는지 확인하세요.\n"+
				"  지금까지 받은 값: %v", out)
			return nil
		}
	}
}

func TestPing(t *testing.T) {
	var got string
	runWithTimeout(t, "Ping이 멈췄습니다. 버퍼 없는 채널은 보내는 쪽과 받는 쪽이 둘 다 있어야 합니다 — go 키워드를 빠뜨리지 않았나요?", func() {
		got = Ping()
	})
	if got != "pong" {
		t.Errorf("Ping() = %q; want %q", got, "pong")
	}
}

func TestGenerator(t *testing.T) {
	// Generator는 값을 다 보낼 때까지 기다리지 않고 즉시 반환해야 한다.
	var ch <-chan int
	runWithTimeout(t, "Generator가 반환하지 않습니다. 값을 보내는 부분을 go func(){...}() 로 감쌌는지 확인하세요", func() {
		ch = Generator(4)
	})

	got := collect(t, ch)
	want := []int{0, 1, 2, 3}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Generator(4) = %v; want %v", got, want)
	}

	// n이 0이면 아무것도 안 보내고 바로 닫혀야 한다.
	if got := collect(t, Generator(0)); len(got) != 0 {
		t.Errorf("Generator(0) = %v; want 빈 슬라이스", got)
	}
}

func TestSquare(t *testing.T) {
	got := collect(t, Square(Generator(5)))
	want := []int{0, 1, 4, 9, 16}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Square(Generator(5)) = %v; want %v (순서도 유지되어야 합니다)", got, want)
	}

	// 입력이 닫히면 출력도 닫혀야 한다. 안 그러면 collect가 타임아웃난다.
	if got := collect(t, Square(Generator(0))); len(got) != 0 {
		t.Errorf("Square(Generator(0)) = %v; want 빈 슬라이스", got)
	}
}

func TestSumConcurrent(t *testing.T) {
	nums := make([]int, 100)
	want := 0
	for i := range nums {
		nums[i] = i
		want += i
	}

	for _, workers := range []int{0, 1, 3, 7, 100, 200} {
		var got int
		runWithTimeout(t, "SumConcurrent가 멈췄습니다. 보낸 개수와 받는 개수가 맞는지 확인하세요", func() {
			got = SumConcurrent(nums, workers)
		})
		if got != want {
			t.Errorf("SumConcurrent(0..99, workers=%d) = %d; want %d", workers, got, want)
		}
	}

	var got int
	runWithTimeout(t, "SumConcurrent(nil, ...)이 멈췄습니다", func() {
		got = SumConcurrent(nil, 4)
	})
	if got != 0 {
		t.Errorf("SumConcurrent(nil, 4) = %d; want 0", got)
	}
}

func TestWorkerPool(t *testing.T) {
	jobs := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	want := make(map[int]int, len(jobs))
	for _, j := range jobs {
		want[j] = j * j
	}

	for _, workers := range []int{0, 1, 3, 20} {
		var got map[int]int
		runWithTimeout(t, "WorkerPool이 멈췄습니다. jobs 채널을 닫았는지, 결과를 정확히 len(jobs)개 받는지 확인하세요", func() {
			got = WorkerPool(jobs, workers)
		})
		if !reflect.DeepEqual(got, want) {
			t.Errorf("WorkerPool(jobs, workers=%d) = %v; want %v", workers, got, want)
		}
	}

	var got map[int]int
	runWithTimeout(t, "WorkerPool(nil, ...)이 멈췄습니다", func() {
		got = WorkerPool(nil, 3)
	})
	if len(got) != 0 {
		t.Errorf("WorkerPool(nil, 3) = %v; want 빈 맵", got)
	}
}

// 일감이 많을 때도 채널 개수가 맞는지 확인합니다.
// 워커 수보다 일감이 훨씬 많으면, 보내는 쪽과 받는 쪽의 개수가 어긋나 있을 때
// 확실히 멈춥니다. 작은 입력에서는 우연히 통과할 수 있습니다.
func TestWorkerPoolManyJobs(t *testing.T) {
	jobs := make([]int, 1000)
	want := make(map[int]int, len(jobs))
	for i := range jobs {
		jobs[i] = i
		want[i] = i * i
	}
	var got map[int]int
	runWithTimeout(t, "일감 1000개에서 WorkerPool이 멈췄습니다. 버퍼 없는 채널에 "+
		"보내는 쪽과 받는 쪽의 개수가 안 맞으면 여기서 막힙니다", func() {
		got = WorkerPool(jobs, 8)
	})
	if !reflect.DeepEqual(got, want) {
		t.Errorf("일감 1000개 결과가 다릅니다 (len=%d, want len=%d)", len(got), len(want))
	}
}

func TestRecvOne(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 42
	if v, ok := RecvOne(ch); v != 42 || !ok {
		t.Errorf("RecvOne(값이 있는 채널) = %d, %v; want 42, true", v, ok)
	}

	closed := make(chan int)
	close(closed)
	var (
		v  int
		ok bool
	)
	runWithTimeout(t, "닫힌 채널에서 받는데 멈췄습니다 — 닫힌 채널 수신은 즉시 반환됩니다", func() {
		v, ok = RecvOne(closed)
	})
	if v != 0 || ok {
		t.Errorf("RecvOne(닫힌 채널) = %d, %v; want 0, false", v, ok)
	}

	// 닫히기 전에 버퍼에 남아있던 값은 먼저 나온다.
	buf := make(chan int, 1)
	buf <- 7
	close(buf)
	if v, ok := RecvOne(buf); v != 7 || !ok {
		t.Errorf("RecvOne(닫혔지만 버퍼에 값이 남은 채널) = %d, %v; want 7, true", v, ok)
	}
}
