// Package basics — Go 기초 문법.
//
// 각 함수의 TODO를 채운 뒤 실행하세요:
//
//	go test ./01_basics/
//
// import는 직접 추가해야 합니다. (strconv 등)
package basics

import (
    "strconv"
) 

// 문제 1 — 다중 반환값
//
// a와 b를 바꿔서 반환하세요.
//
// Go 함수는 값을 여러 개 반환합니다. C처럼 포인터를 넘길 필요도,
// Python처럼 튜플을 만들 필요도 없습니다.
func Swap(a, b int) (int, int) {
	// TODO
        tmp := a
        a = b
        b = tmp
	return a, b
}
// Q. 포인터가 아니라 int를 인자로 사용하면 Swap의 의미가 없는 거 아닌가?

// 문제 2 — 반복문
//
// 1부터 n까지의 합을 반환하세요. n이 0 이하면 0을 반환합니다.
//
// Go에는 while이 없습니다. 반복문은 for 하나뿐입니다.
// 조건부만 쓰면 while처럼 동작합니다: for x < 10 { ... }
func SumTo(n int) int {
	// TODO
        result := 0
        for n > 0 {
            result += n
            n -= 1
        }
	return result
}

// 문제 3 — switch
//
// 점수를 등급으로 변환하세요.
//
//	90 이상 → "A"
//	80 이상 → "B"
//	70 이상 → "C"
//	그 외   → "F"
//
// Go의 switch는 case가 끝나면 자동으로 빠져나옵니다. break를 쓰지 않습니다.
// 조건식 없는 switch는 if-else 사슬을 대신합니다:
//
//	switch {
//	case score >= 90:
//	    ...
//	}
func Grade(score int) string {
	// TODO
        switch {
        case score >= 90:
            return "A"
        case score >= 80:
            return "B"
        case score >= 70:
            return "C"
        }
	return "F"
}

// 문제 4 — 가변 인자
//
// 인자 중 가장 큰 값을 반환하세요.
// 인자가 하나도 없으면 (0, false)를 반환합니다.
//
// nums는 함수 안에서 []int(슬라이스)입니다.
// "값이 없을 수도 있다"를 Go는 두 번째 반환값 bool로 표현합니다.
// (Python의 None, C의 -1 매직넘버 대신)
func Max(nums ...int) (int, bool) {
	// TODO
        length := len(nums)
        if length == 0 {
            return 0, false
        }

        max := nums[0]
        for i:=1; i<length; i++ {
            t := nums[i]
            if max < t { max = t }
        }
        return max, true
}

// 문제 5 — 포인터
//
// p가 가리키는 값을 2배로 만드세요. p가 nil이면 아무것도 하지 않습니다.
//
// Go의 포인터는 C와 같지만 포인터 산술(p++)이 없습니다.
// nil 포인터를 역참조하면 panic이 납니다 — segfault 대신 스택 트레이스가 뜹니다.
func Double(p *int) {
	// TODO
        if p == nil { return }
        *p *= 2
}

// 문제 6 — 클로저
//
// 호출할 때마다 1, 2, 3... 을 반환하는 함수를 만들어 반환하세요.
// Counter()를 두 번 호출하면 서로 독립적인 카운터가 나와야 합니다.
//
// Go에서 함수는 값입니다. 바깥 변수를 붙잡아(capture) 둘 수 있습니다.
// Python의 nonlocal과 같고, C에는 없는 기능입니다.
func Counter() func() int {
	// TODO
        counter := 0
	return func() int {
            counter += 1
            return counter
        }
}

// 문제 7 — 조합
//
// FizzBuzz를 문자열 하나로 반환하세요.
//
//	3의 배수      → "Fizz"
//	5의 배수      → "Buzz"
//	15의 배수     → "FizzBuzz"
//	그 외         → 숫자를 문자열로 ("7")
//
// 힌트: 정수를 문자열로 바꾸려면 strconv.Itoa를 씁니다.
// string(7)은 "7"이 아니라 유니코드 코드포인트 7번 문자가 되니 주의하세요.
func FizzBuzz(n int) string {
	// TODO
        switch {
        case n % 15 == 0:
            return "FizzBuzz"
        case n % 3 == 0:
            return "Fizz"
        case n % 5 == 0:
            return "Buzz"
        }
	return strconv.Itoa(n)
}
