// Package conc — 06_goroutines 정답.
package conc

// 문제 1 — goroutine과 채널 기초
func Ping() string {
	ch := make(chan string)
	go func() {
		ch <- "pong"
	}()
	// 버퍼 없는 채널이라 goroutine의 보내기와 여기의 받기가 만나야 둘 다 진행된다.
	// 이 "만남" 자체가 동기화다. 뮤텍스도, WaitGroup도 필요 없다.
	return <-ch
}

// 문제 2 — 채널을 반환하는 함수
func Generator(n int) <-chan int {
	ch := make(chan int)

	// go로 감싸는 게 핵심이다. 감싸지 않으면 첫 ch <- 0 에서 영원히 멈춘다
	// (받을 사람은 이 함수가 반환한 뒤에야 생기므로).
	go func() {
		// defer close(ch)로 써도 된다. 어느 경로로 끝나든 반드시 닫힌다.
		defer close(ch)
		for i := 0; i < n; i++ {
			ch <- i
		}
	}()

	// 반환 타입이 <-chan int라 호출한 쪽은 받기만 할 수 있다.
	// chan int를 <-chan int로 넘기는 변환은 자동이다.
	return ch
}

// 문제 3 — 파이프라인
func Square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		// in이 닫히면 range가 끝나고, defer가 out을 닫는다.
		// 이 "입력이 닫히면 출력도 닫는다" 규칙 덕분에 단계를 몇 개든 이을 수 있다.
		for v := range in {
			out <- v * v
		}
	}()
	return out
}

// 문제 4 — 여러 goroutine의 결과 모으기
func SumConcurrent(nums []int, workers int) int {
	if workers < 1 {
		workers = 1
	}
	if len(nums) == 0 {
		return 0
	}
	if workers > len(nums) {
		// 일감보다 워커가 많으면 빈 워커가 생긴다.
		// 빈 워커도 0을 보내면 되지만, 아예 줄이는 게 깔끔하다.
		workers = len(nums)
	}

	partials := make(chan int, workers) // 버퍼를 주면 워커가 보내고 바로 끝날 수 있다
	chunk := (len(nums) + workers - 1) / workers

	for w := 0; w < workers; w++ {
		lo := w * chunk
		hi := min(lo+chunk, len(nums))
		// Go 1.22부터 루프 변수 w는 반복마다 새로 만들어지므로
		// 예전처럼 go func(lo, hi int){...}(lo, hi) 로 넘길 필요가 없다.
		go func() {
			sum := 0
			for _, n := range nums[lo:hi] {
				sum += n
			}
			// total += sum 처럼 공유 변수를 건드리면 데이터 레이스다.
			// 각자 자기 결과만 채널로 보내고, 합치는 건 한 goroutine이 한다.
			partials <- sum
		}()
	}

	total := 0
	for i := 0; i < workers; i++ {
		// 정확히 workers개를 받는다. 개수가 안 맞으면 여기서 영원히 멈춘다.
		total += <-partials
	}
	return total
}

// 문제 5 — 워커 풀 ★
func WorkerPool(jobs []int, workers int) map[int]int {
	if workers < 1 {
		workers = 1
	}

	type result struct{ job, val int }

	jobsCh := make(chan int)
	resultsCh := make(chan result)

	// 워커들. 전부 같은 jobsCh에서 꺼내 간다.
	// 하나의 일감은 정확히 한 워커에게만 간다 — 채널이 알아서 보장해준다.
	for w := 0; w < workers; w++ {
		go func() {
			for j := range jobsCh {
				resultsCh <- result{job: j, val: j * j}
			}
			// jobsCh가 닫히면 range가 끝나고 이 goroutine도 끝난다.
		}()
	}

	// 일감을 넣는 것도 goroutine이어야 한다.
	// 여기서 동기적으로 넣으면, 결과를 받는 아래 루프에 도달하기 전에
	// resultsCh가 막혀서 전체가 멈춘다.
	go func() {
		defer close(jobsCh) // 닫아야 워커들의 range가 끝난다
		for _, j := range jobs {
			jobsCh <- j
		}
	}()

	// 결과 개수를 미리 안다(len(jobs)). 그래서 resultsCh는 닫을 필요도 없다.
	// 맵에 쓰는 건 이 goroutine 하나뿐이므로 안전하다.
	out := make(map[int]int, len(jobs))
	for i := 0; i < len(jobs); i++ {
		r := <-resultsCh
		out[r.job] = r.val
	}
	return out

	// 8장의 TCP 서버가 이것과 같은 구조다.
	// 일감이 "접속"으로 바뀌고, 워커 대신 접속마다 goroutine을 새로 띄울 뿐이다.
}

// 문제 6 — 닫힌 채널 구별하기
func RecvOne(ch <-chan int) (int, bool) {
	// 닫힌 채널에서 받으면 멈추지 않고 즉시 (제로값, false)가 나온다.
	// ok를 안 보면 "0이 무한히 나오는" 루프에 빠지기 쉽다.
	v, ok := <-ch
	return v, ok
}
