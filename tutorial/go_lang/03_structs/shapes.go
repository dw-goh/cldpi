// Package shapes — 구조체와 메서드.
//
//	go test ./03_structs/
//
// 타입 선언은 미리 제공합니다(테스트가 필드 이름에 의존하므로).
// 여러분이 채울 것은 메서드 본문입니다.
package shapes

import "fmt"

// Point는 2차원 좌표입니다.
//
// C의 struct와 거의 같습니다. 다만 Go의 구조체는 필드가 전부 비교 가능하면
// == 로 비교할 수 있고, 맵의 키로도 쓸 수 있습니다 (문제 6).
type Point struct {
	X, Y int
}

// 문제 1 — 값 리시버 메서드
//
// p와 q를 더한 새 Point를 반환하세요. p 자신은 바뀌지 않습니다.
//
// (p Point)가 리시버입니다. C에서 첫 인자로 struct를 넘기던 것과 같지만,
// p.Add(q) 형태로 부를 수 있습니다.
// 값 리시버는 복사본을 받습니다 — p를 고쳐도 호출한 쪽에 반영되지 않습니다.
func (p Point) Add(q Point) Point {
	// TODO
        n_X := p.X + q.X
        n_Y := p.Y + q.Y
	return Point{X: n_X, Y: n_Y}
}

// 문제 2 — 포인터 리시버 메서드 ★
//
// p의 X, Y를 각각 f배로 만드세요. (자기 자신을 수정합니다)
//
// (p *Point)는 포인터 리시버입니다. 이걸 (p Point)로 바꾸면
// 복사본만 바뀌고 원본은 그대로입니다 — 컴파일은 되지만 테스트가 실패합니다.
//
// 호출할 때 (&p).Scale(2)라고 쓸 필요는 없습니다.
// p가 주소를 얻을 수 있는 변수라면 p.Scale(2)로 쓰면 Go가 알아서 &p로 바꿉니다.
func (p *Point) Scale(f int) {
	// TODO
        p.X *= f
        p.Y *= f
}

// Rect는 직사각형입니다.
type Rect struct {
	W, H float64
}

// 문제 3 — 생성자 패턴
//
// Rect를 만들어 포인터로 반환하세요.
// w나 h가 음수면 0으로 보정합니다.
//
// Go에는 생성자 문법이 없습니다. NewXxx 함수를 만드는 게 관례입니다.
// 검증이나 기본값 설정이 필요할 때만 만들고, 아니면 그냥 Rect{W: 3, H: 4}를 씁니다.
func NewRect(w, h float64) *Rect {
	// TODO
        if w < 0 { w = 0 }
        if h < 0 { h = 0 }
        
	return &Rect{W: w, H: h}
}

// 문제 3 (계속) — 넓이를 반환하세요.
func (r Rect) Area() float64 {
	// TODO
	return r.W * r.H
}

// Counter는 제로값 상태로 바로 쓸 수 있어야 합니다.
//
// n이 소문자라 패키지 밖에서는 접근할 수 없습니다 (private 필드).
type Counter struct {
	n int
}

// 문제 4 — 제로값이 쓸모있게
//
// Inc는 카운터를 1 증가시킵니다.
// Value는 현재 값을 반환합니다.
//
// 핵심: var c Counter 라고만 써도 바로 동작해야 합니다.
// Go 표준 라이브러리는 이 원칙을 지킵니다 — sync.Mutex, bytes.Buffer 모두
// var mu sync.Mutex 처럼 선언만 하면 바로 쓸 수 있습니다.
// "쓸모있는 제로값(useful zero value)"이라고 부릅니다.
func (c *Counter) Inc() {
	// TODO
        c.n++
}

func (c Counter) Value() int {
	// TODO
	return c.n 
}

// Animal과 Dog — 임베딩(embedding)
type Animal struct {
	Name string
}

// Speak는 Animal의 기본 동작입니다. (이미 구현되어 있습니다)
func (a Animal) Speak() string {
	return a.Name + " makes a sound"
}

// Dog는 Animal을 임베딩합니다. 필드 이름이 없다는 점에 주목하세요.
// 이렇게 하면 Animal의 필드와 메서드가 Dog로 "승격(promote)"됩니다.
//
//	d.Name        // d.Animal.Name 과 같다
//	d.Speak()     // d.Animal.Speak() 와 같다
//
// C++의 상속과 비슷해 보이지만 상속이 아닙니다. Dog는 Animal이 "아니고",
// 그냥 Animal을 품고 있으면서 이름을 빌려 쓰는 것뿐입니다.
type Dog struct {
	Animal
	Breed string
}

// 문제 5 — 임베딩과 메서드 가리기
//
// Dog.Speak는 "<이름> says woof" 를 반환해야 합니다.
//
//	Dog{Animal{"Rex"}, "Shiba"}.Speak() → "Rex says woof"
//
// Dog에 Speak를 정의하면 승격된 Animal.Speak를 가립니다(shadow).
// 안쪽 것을 부르고 싶으면 d.Animal.Speak()라고 명시하면 됩니다.
func (d Dog) Speak() string {
	// TODO
	return fmt.Sprintf("%s says woof", d.Name)
}

// 문제 6 — 구조체를 맵의 키로
//
// 각 Point가 몇 번 나왔는지 세세요.
//
//	CountPoints([]Point{{1,2}, {1,2}, {3,4}})
//	  → map[Point]int{{1,2}: 2, {3,4}: 1}
//
// 필드가 전부 비교 가능한 구조체는 == 비교와 맵 키가 됩니다.
// (슬라이스나 맵을 필드로 가지면 비교 불가능해서 키로 못 씁니다)
func CountPoints(ps []Point) map[Point]int {
	// TODO
        m := make(map[Point]int)
        for _, p := range ps {
            m[p] += 1
        }
	return m
}
