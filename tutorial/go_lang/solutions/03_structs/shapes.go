// Package shapes — 03_structs 정답.
package shapes

type Point struct {
	X, Y int
}

// 문제 1 — 값 리시버 메서드
//
// 값 리시버라 p는 복사본이다. 새 Point를 만들어 반환하므로
// 호출한 쪽의 p는 그대로 남는다.
func (p Point) Add(q Point) Point {
	return Point{X: p.X + q.X, Y: p.Y + q.Y}
}

// 문제 2 — 포인터 리시버 메서드 ★
//
// 리시버가 *Point이므로 p는 호출한 쪽의 변수를 가리킨다.
// p.X처럼 써도 된다 — Go는 포인터 리시버의 필드 접근에
// (*p).X 를 자동으로 넣어준다. C의 -> 연산자가 필요 없다.
func (p *Point) Scale(f int) {
	p.X *= f
	p.Y *= f
}

type Rect struct {
	W, H float64
}

// 문제 3 — 생성자 패턴
//
// 포인터를 반환한다고 해서 C처럼 malloc/free를 신경 쓸 필요는 없다.
// 지역변수의 주소를 반환해도 안전하다 — escape analysis가 힙으로 옮긴다.
func NewRect(w, h float64) *Rect {
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	return &Rect{W: w, H: h}
}

// Area는 값 리시버로 충분하다. 아무것도 수정하지 않고 구조체도 작다.
func (r Rect) Area() float64 {
	return r.W * r.H
}

type Counter struct {
	n int
}

// 문제 4 — 제로값이 쓸모있게
//
// Inc는 상태를 바꾸므로 반드시 포인터 리시버여야 한다.
// (c Counter)로 쓰면 복사본의 n만 증가하고 조용히 아무 일도 안 일어난다.
func (c *Counter) Inc() {
	c.n++
}

// Value는 읽기만 하므로 값 리시버로도 된다.
//
// 다만 실무에서는 한 타입의 메서드를 값/포인터 중 하나로 통일하는 게 관례다.
// 특히 나중에 sync.Mutex를 필드로 넣게 되면 값 리시버는 뮤텍스를 복사해버려서
// 위험해진다 (go vet이 경고한다). 여기서는 대비를 보여주려고 일부러 섞었다.
func (c Counter) Value() int {
	return c.n
}

type Animal struct {
	Name string
}

func (a Animal) Speak() string {
	return a.Name + " makes a sound"
}

type Dog struct {
	Animal
	Breed string
}

// 문제 5 — 임베딩과 메서드 가리기
//
// Dog에 Speak를 정의하면 승격된 Animal.Speak를 가린다.
// d.Name은 d.Animal.Name의 줄임말이다.
func (d Dog) Speak() string {
	return d.Name + " says woof"
	// 안쪽 구현을 재사용하고 싶다면 d.Animal.Speak()를 명시적으로 부른다.
	// C++의 Base::method()와 비슷하지만, 이건 상속이 아니라
	// 그냥 필드에 들어있는 값의 메서드를 부르는 것뿐이다.
}

// 문제 6 — 구조체를 맵의 키로
func CountPoints(ps []Point) map[Point]int {
	m := make(map[Point]int)
	for _, p := range ps {
		m[p]++
	}
	return m
	// Point는 int 필드만 있어서 비교 가능(comparable)하다.
	// 슬라이스, 맵, 함수를 필드로 가지면 비교 불가능해져서
	// == 도 맵 키도 쓸 수 없다 (컴파일 에러).
}
