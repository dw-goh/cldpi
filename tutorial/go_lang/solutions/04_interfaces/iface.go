// Package iface — 04_interfaces 정답.
package iface

import (
	"fmt"
	"io"
	"math"
)

type Shape interface {
	Area() float64
}

type Rect struct {
	W, H float64
}

type Circle struct {
	R float64
}

// 문제 1 — 인터페이스를 암묵적으로 만족시키기
//
// Shape를 언급하는 곳이 어디에도 없다는 점에 주목.
// 메서드 시그니처가 맞으면 그걸로 끝이다.
func (r Rect) Area() float64 {
	return r.W * r.H
}

func (c Circle) Area() float64 {
	return math.Pi * c.R * c.R
}

var (
	_ Shape = Rect{}
	_ Shape = Circle{}
)

// 문제 2 — 인터페이스를 받는 함수
func TotalArea(shapes []Shape) float64 {
	total := 0.0
	for _, s := range shapes {
		// s의 구체 타입이 뭔지 모른다. 알 필요도 없다.
		// 런타임에 s 안의 타입 정보를 보고 올바른 Area가 호출된다
		// (C++의 vtable과 비슷한 방식이다).
		total += s.Area()
	}
	return total
}

type Temp float64

// 문제 3 — fmt.Stringer 구현
func (t Temp) String() string {
	// float64(t)로 변환하는 게 핵심이다.
	// fmt.Sprintf("%v°C", t)라고 쓰면 fmt가 t의 String()을 다시 불러서
	// 무한 재귀 → 스택 오버플로가 난다. go vet이 이걸 잡아준다.
	return fmt.Sprintf("%.1f°C", float64(t))
}

// 문제 4 — 타입 스위치
func Describe(v any) string {
	switch x := v.(type) {
	case nil:
		// v 자체가 nil 인터페이스일 때만 걸린다.
		return "nil"
	case int:
		// 이 블록 안에서 x의 타입은 int다.
		return fmt.Sprintf("int: %d", x)
	case string:
		return fmt.Sprintf("string: %s", x)
	case Shape:
		// Rect도 Circle도 여기 걸린다.
		// 이 case를 case Rect 위에 두면 Rect는 절대 안 걸리므로
		// 구체적인 타입을 먼저 쓰는 게 원칙이다.
		return fmt.Sprintf("shape: %.2f", x.Area())
	default:
		// x는 v와 같은 값, 타입은 any 그대로다.
		return "unknown"
	}
}

// 문제 5 — 타입 단언 (comma-ok)
func AsCircle(s Shape) (Circle, bool) {
	// comma-ok 없이 s.(Circle)이라고 쓰면 s가 Rect일 때 panic이다.
	c, ok := s.(Circle)
	return c, ok
}

type CountingWriter struct {
	N int
}

// 문제 6 — 표준 라이브러리 인터페이스 구현하기 ★
//
// io.Writer의 계약: 읽은 만큼의 n을 반환하고, n < len(p)면 반드시 err != nil.
// 우리는 항상 전부 "쓰므로" (len(p), nil)을 반환한다.
//
// N을 수정해야 하니 포인터 리시버다. 그래서 *CountingWriter만 io.Writer이고
// CountingWriter는 아니다 — 호출할 때 &w를 넘겨야 하는 이유.
func (w *CountingWriter) Write(p []byte) (int, error) {
	w.N += len(p)
	return len(p), nil
}

var _ io.Writer = (*CountingWriter)(nil)
