// Package conc — goroutine과 channel.
//
//	go test ./06_goroutines/
//	go test -race ./06_goroutines/     ← 동시성 코드는 항상 -race로도 돌려보세요
//
// Go의 간판 기능입니다. M0의 "goroutine per connection"이 여기서 나옵니다.
package conc

// 문제 1 — goroutine과 채널 기초
//
// goroutine을 하나 띄워서 채널로 "pong"을 보내고, 그걸 받아서 반환하세요.
//
//	ch := make(chan string)      // 버퍼 없는 채널
//	go func() { ch <- "pong" }() // go 키워드로 goroutine 시작
//	return <-ch                  // 받을 때까지 여기서 멈춘다
//
// 버퍼 없는 채널은 "만나는 지점(rendezvous)"입니다.
// 보내는 쪽은 받는 쪽이 준비될 때까지, 받는 쪽은 보내는 쪽이 올 때까지 멈춥니다.
// 이 대기가 곧 동기화입니다 — 뮤텍스를 안 써도 됩니다.
func Ping() string {
	// TODO
	return ""
}

// 문제 2 — 채널을 반환하는 함수
//
// 0부터 n-1까지를 순서대로 보내고 채널을 닫으세요.
// 이 함수는 goroutine이 값을 다 보낼 때까지 기다리지 않고 즉시 반환해야 합니다.
//
//	for v := range Generator(3) { ... }   // 0, 1, 2를 받고 루프가 끝난다
//
// 반환 타입 <-chan int는 "받기 전용 채널"입니다.
// 호출한 쪽이 실수로 여기에 값을 보내거나 채널을 닫는 걸 컴파일러가 막아줍니다.
//
// ★ close는 **보내는 쪽이** 합니다. 받는 쪽이 닫으면 안 됩니다.
// 닫은 채널에 보내면 panic입니다. 닫힌 채널에서 받으면 제로값이 무한정 나옵니다.
// range는 채널이 닫히면 자동으로 끝납니다 — 안 닫으면 영원히 멈춥니다(deadlock).
func Generator(n int) <-chan int {
	// TODO
	return nil
}

// 문제 3 — 파이프라인
//
// in에서 받은 값을 제곱해서 내보내는 채널을 반환하세요.
// in이 닫히면 출력 채널도 닫아야 합니다.
//
//	Square(Generator(4))  → 0, 1, 4, 9
//
// 이렇게 채널로 단계를 잇는 걸 파이프라인이라고 합니다.
// 유닉스의 `cat x | grep y | wc -l` 과 같은 구조입니다.
func Square(in <-chan int) <-chan int {
	// TODO
	return nil
}

// 문제 4 — 여러 goroutine의 결과 모으기
//
// nums의 합을 workers개의 goroutine으로 나눠서 계산하세요.
// workers가 1보다 작으면 1로 취급합니다.
//
// 각 goroutine이 자기 몫의 부분합을 채널로 보내고,
// 메인이 정확히 workers개를 받아 더하면 됩니다.
//
// ★ 여러 goroutine이 같은 변수(예: total += x)를 건드리면 데이터 레이스입니다.
// 채널로 결과를 모으면 그런 문제가 없습니다.
// "메모리를 공유해서 통신하지 말고, 통신해서 메모리를 공유하라"
func SumConcurrent(nums []int, workers int) int {
	// TODO
	return 0
}

// 문제 5 — 워커 풀 ★
//
// jobs의 각 값 j에 대해 j*j를 계산해서 map[j]j*j를 반환하세요.
// 계산은 workers개의 goroutine이 나눠서 합니다. workers < 1이면 1로 취급합니다.
//
// 구조:
//
//	jobsCh    ← 메인이 일감을 넣고 닫는다
//	  ↓ (워커 여러 개가 range로 꺼내 간다 — 하나의 일감은 한 워커만 가져간다)
//	resultsCh ← 워커들이 결과를 넣는다
//	  ↓
//	메인이 모아서 맵을 만든다
//
// ★ 워커들이 맵에 직접 쓰면 안 됩니다. Go의 맵은 동시 쓰기에 안전하지 않고,
// 런타임이 감지하면 "concurrent map writes"로 프로그램을 죽입니다.
// 결과는 반드시 채널로 모아서 한 goroutine이 맵에 넣어야 합니다.
//
// 힌트: 결과 개수를 알고 있으니(len(jobs)) 그만큼만 받으면 됩니다.
// resultsCh를 닫을 필요조차 없습니다.
//
// 이게 8장의 TCP 서버와 같은 구조입니다 — 일감이 "접속"으로 바뀔 뿐입니다.
func WorkerPool(jobs []int, workers int) map[int]int {
	// TODO
	return nil
}

// 문제 6 — 닫힌 채널 구별하기
//
// ch에서 값 하나를 받아 (값, 열려있었는지)를 반환하세요.
// 채널이 이미 닫혔으면 (0, false)를 반환합니다.
//
//	v, ok := <-ch
//
// ok는 "이 값이 진짜 보내진 값인가"를 알려줍니다.
// 닫힌 채널에서 받으면 즉시 (제로값, false)가 나옵니다 — 멈추지 않습니다.
// 이걸 모르면 "왜 0이 무한히 나오지?" 하는 무한 루프에 빠집니다.
func RecvOne(ch <-chan int) (int, bool) {
	// TODO
	return 0, false
}
