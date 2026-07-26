// Package iface — 인터페이스.
//
//	go test ./04_interfaces/
//
// Go의 인터페이스는 "이 메서드들을 가진 아무 타입"이라는 뜻입니다.
// implements 키워드가 없습니다. 메서드가 맞으면 그냥 만족합니다.
package iface

import "io"
import "math"
import "fmt"

// Shape는 넓이를 잴 수 있는 모든 것입니다.
//
// 인터페이스는 "이 메서드 집합"을 가리키는 타입입니다.
// 인터페이스는 작게 만드는 게 Go의 관례입니다 — 표준 라이브러리의
// io.Reader, io.Writer, fmt.Stringer, error 전부 메서드가 1개입니다.
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
// Rect와 Circle에 Area()를 구현하세요. 그러면 둘 다 자동으로 Shape가 됩니다.
// "Rect implements Shape" 같은 선언을 어디에도 쓰지 않는다는 점에 주목하세요.
//
// 원의 넓이는 math.Pi * R * R 입니다.
func (r Rect) Area() float64 {
	// TODO
	return r.W * r.H
}

func (c Circle) Area() float64 {
	// TODO
	return math.Pi * c.R * c.R
}

// 컴파일 타임 검증 관용구.
// "Rect는 Shape를 만족해야 한다"를 컴파일러에게 확인시킵니다.
// 값을 만들지 않으려고 nil을 Rect로 변환해 넣습니다.
// 실무에서 인터페이스 구현이 깨지는 걸 빨리 잡으려고 자주 씁니다.
var (
	_ Shape = Rect{}
	_ Shape = Circle{}
)

// 문제 2 — 인터페이스를 받는 함수
//
// 모든 도형의 넓이 합을 반환하세요.
//
// shapes 안에 Rect가 들었는지 Circle이 들었는지 신경 쓰지 않습니다.
// Area()를 부를 수 있다는 것만 압니다. 이게 Go의 다형성입니다.
func TotalArea(shapes []Shape) float64 {
	// TODO
        totalArea := 0.0
        for _, s := range shapes {
            totalArea += s.Area()
	}
        return totalArea
}

// Temp는 섭씨 온도입니다. 구조체가 아니어도 메서드를 붙일 수 있습니다.
type Temp float64

// 문제 3 — fmt.Stringer 구현
//
// String()은 "21.5°C" 형태를 반환해야 합니다. 소수점 첫째 자리까지.
//
//	Temp(21.5).String() → "21.5°C"
//	Temp(-3).String()   → "-3.0°C"
//
// String() string 메서드가 있는 타입은 fmt.Stringer를 만족합니다.
// fmt.Println이나 %v가 알아서 이 메서드를 불러줍니다.
//
// 힌트: fmt.Sprintf("%.1f°C", float64(t))
// 주의: String() 안에서 %v로 t 자신을 출력하면 무한 재귀입니다.
// 반드시 float64(t)로 변환해서 넘기세요.
func (t Temp) String() string {
	// TODO
	return fmt.Sprintf("%.1f°C", float64(t))
}

// 문제 4 — 타입 스위치
//
// v의 종류에 따라 아래 문자열을 반환하세요.
//
//	nil        → "nil"
//	int        → "int: 42"
//	string     → "string: hi"
//	Shape      → "shape: 12.00"     (Area()를 %.2f로)
//	그 외      → "unknown"
//
// any는 interface{}의 별칭입니다 — "아무 메서드도 요구하지 않는 인터페이스",
// 즉 모든 타입이 만족합니다. C의 void*와 비슷하지만 타입 정보를 들고 다닙니다.
//
// 문법:
//
//	switch x := v.(type) {
//	case nil:
//	case int:
//	    // 여기서 x는 int 타입
//	case Shape:
//	    // 여기서 x는 Shape 타입
//	default:
//	}
//
// 주의: case 순서가 중요합니다. Rect는 Shape이기도 하므로
// 더 구체적인 case를 위에 둡니다.
func Describe(v any) string {
	// TODO
        switch x := v.(type) {
        case nil:
            return "nil"
        case int:
            return fmt.Sprintf("int: %d", int(x))
        case string:
            return fmt.Sprintf("string: %s", string(x))
        case Shape:
            return fmt.Sprintf("shape: %.2f", x.Area())
        }
	return "unknown"
}

// 문제 5 — 타입 단언 (comma-ok)
//
// s가 실제로 Circle이면 (그 Circle, true)를, 아니면 (Circle{}, false)를 반환하세요.
//
//	c, ok := s.(Circle)
//
// ok를 받지 않고 c := s.(Circle) 이라고 쓰면 실패했을 때 panic입니다.
// 확신이 없으면 항상 comma-ok 형태를 쓰세요.
func AsCircle(s Shape) (Circle, bool) {
	// TODO
        c, ok := s.(Circle)
        if ok { return c, true }
	return Circle{}, false
}

// CountingWriter는 자기에게 쓰인 바이트 수를 셉니다.
type CountingWriter struct {
	N int
}

// 문제 6 — 표준 라이브러리 인터페이스 구현하기 ★
//
// Write는 p의 바이트 수를 N에 더하고 (len(p), nil)을 반환해야 합니다.
// 데이터를 어디에 저장할 필요는 없습니다 — 세기만 합니다.
//
// 이 메서드 하나로 CountingWriter는 io.Writer가 됩니다.
// 그러면 io.Writer를 받는 표준 라이브러리 함수 전부에 넣을 수 있습니다:
// fmt.Fprintf, io.Copy, json.NewEncoder, gzip.NewWriter, ...
//
// io.Writer의 정의는 이게 전부입니다:
//
//	type Writer interface {
//	    Write(p []byte) (n int, err error)
//	}
//
// M1부터 계속 쓰게 됩니다. net.Conn도 io.Reader이자 io.Writer입니다.
func (w *CountingWriter) Write(p []byte) (int, error) {
	// TODO
        w.N += len(p)
	return len(p), nil
}

var _ io.Writer = (*CountingWriter)(nil)
