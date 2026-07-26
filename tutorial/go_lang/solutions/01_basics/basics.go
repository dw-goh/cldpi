// Package basics — 01_basics 정답.
package basics

import "strconv"

// 문제 1 — 다중 반환값
func Swap(a, b int) (int, int) {
	return b, a
}

// 문제 2 — 반복문
func SumTo(n int) int {
	sum := 0
	for i := 1; i <= n; i++ {
		sum += i
	}
	// n <= 0이면 루프가 한 번도 안 돌고 0이 그대로 반환된다.
	// 별도의 if로 막을 필요가 없다.
	return sum
}

// 문제 3 — switch
func Grade(score int) string {
	// 조건식 없는 switch. if-else 사슬보다 읽기 좋아서 Go에서 자주 쓴다.
	// case가 끝나면 자동으로 빠져나온다 — break를 쓰지 않는다.
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	default:
		return "F"
	}
}

// 문제 4 — 가변 인자
func Max(nums ...int) (int, bool) {
	if len(nums) == 0 {
		return 0, false
	}
	// 첫 원소를 초기값으로 잡는다. math.MinInt 같은 상수를 쓰는 것보다
	// 빈 경우를 위에서 이미 걸렀으므로 이쪽이 안전하다.
	best := nums[0]
	for _, n := range nums[1:] {
		if n > best {
			best = n
		}
	}
	return best, true
}

// 문제 5 — 포인터
func Double(p *int) {
	if p == nil {
		return
	}
	*p *= 2
}

// 문제 6 — 클로저
func Counter() func() int {
	// n은 Counter가 반환된 뒤에도 살아있다. 반환된 클로저가 붙잡고 있기 때문에
	// 컴파일러가 n을 힙에 올린다(escape analysis).
	// C에서 지역변수 주소를 반환하면 dangling pointer지만 Go에서는 안전하다.
	n := 0
	return func() int {
		n++
		return n
	}
}

// 문제 7 — 조합
func FizzBuzz(n int) string {
	// 15의 배수를 먼저 검사해야 한다. 순서를 바꾸면 15가 "Fizz"로 잡힌다.
	switch {
	case n%15 == 0:
		return "FizzBuzz"
	case n%3 == 0:
		return "Fizz"
	case n%5 == 0:
		return "Buzz"
	default:
		// string(n)이 아니다. string(int) 변환은 정수를 유니코드 코드포인트로
		// 해석하므로 string(7)은 "7"이 아니라 제어문자가 된다.
		return strconv.Itoa(n)
	}
}
