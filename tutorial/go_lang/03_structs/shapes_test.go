package shapes

import (
	"reflect"
	"testing"
)

func TestPointAdd(t *testing.T) {
	p := Point{X: 1, Y: 2}
	q := Point{X: 10, Y: 20}

	if got, want := p.Add(q), (Point{X: 11, Y: 22}); got != want {
		t.Errorf("p.Add(q) = %v; want %v", got, want)
	}
	// 값 리시버이므로 p 자신은 그대로여야 한다.
	if p != (Point{X: 1, Y: 2}) {
		t.Errorf("Add 후 p = %v; want {1 2} (Add는 p를 바꾸면 안 됩니다)", p)
	}
}

func TestPointScale(t *testing.T) {
	p := Point{X: 3, Y: 4}
	p.Scale(2) // Go가 알아서 (&p).Scale(2)로 바꿔준다

	if want := (Point{X: 6, Y: 8}); p != want {
		t.Errorf("Scale(2) 후 p = %v; want %v (리시버가 *Point 인지 확인하세요)", p, want)
	}

	// 포인터로 직접 불러도 같아야 한다.
	q := &Point{X: 1, Y: 1}
	q.Scale(5)
	if want := (Point{X: 5, Y: 5}); *q != want {
		t.Errorf("Scale(5) 후 *q = %v; want %v", *q, want)
	}
}

func TestNewRect(t *testing.T) {
	r := NewRect(3, 4)
	if r == nil {
		t.Fatal("NewRect(3, 4) = nil; want *Rect")
	}
	if r.W != 3 || r.H != 4 {
		t.Errorf("NewRect(3, 4) = %+v; want {W:3 H:4}", *r)
	}
	// 음수는 0으로 보정.
	bad := NewRect(-1, -2)
	if bad.W != 0 || bad.H != 0 {
		t.Errorf("NewRect(-1, -2) = %+v; want {W:0 H:0}", *bad)
	}
}

func TestRectArea(t *testing.T) {
	tests := []struct {
		r    Rect
		want float64
	}{
		{Rect{W: 3, H: 4}, 12},
		{Rect{W: 0, H: 5}, 0},
		{Rect{W: 2.5, H: 2}, 5},
	}
	for _, tt := range tests {
		if got := tt.r.Area(); got != tt.want {
			t.Errorf("%+v.Area() = %v; want %v", tt.r, got, tt.want)
		}
	}
}

func TestCounter(t *testing.T) {
	// make도 New도 없이 선언만 한다. 이게 바로 동작해야 한다.
	var c Counter

	if got := c.Value(); got != 0 {
		t.Errorf("제로값 Counter.Value() = %d; want 0", got)
	}
	for i := 0; i < 3; i++ {
		c.Inc()
	}
	if got := c.Value(); got != 3 {
		t.Errorf("Inc 3번 후 Value() = %d; want 3 (Inc의 리시버가 *Counter 인지 확인하세요)", got)
	}

	// 구조체 안에 들어있어도 동작해야 한다.
	s := struct{ C Counter }{}
	s.C.Inc()
	if got := s.C.Value(); got != 1 {
		t.Errorf("중첩된 Counter.Value() = %d; want 1", got)
	}
}

func TestDogSpeak(t *testing.T) {
	d := Dog{Animal: Animal{Name: "Rex"}, Breed: "Shiba"}

	// 승격된 필드 — d.Animal.Name 이라고 안 써도 된다.
	if d.Name != "Rex" {
		t.Errorf("d.Name = %q; want %q", d.Name, "Rex")
	}
	if got, want := d.Speak(), "Rex says woof"; got != want {
		t.Errorf("d.Speak() = %q; want %q", got, want)
	}
	// 가려진 원래 메서드는 여전히 부를 수 있다.
	if got, want := d.Animal.Speak(), "Rex makes a sound"; got != want {
		t.Errorf("d.Animal.Speak() = %q; want %q", got, want)
	}
}

func TestCountPoints(t *testing.T) {
	got := CountPoints([]Point{{1, 2}, {3, 4}, {1, 2}, {1, 2}})
	want := map[Point]int{{1, 2}: 3, {3, 4}: 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("CountPoints = %v; want %v", got, want)
	}
	if got := CountPoints(nil); len(got) != 0 {
		t.Errorf("CountPoints(nil) = %v; want 빈 맵", got)
	}
}
