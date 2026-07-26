package iface

import (
	"fmt"
	"math"
	"testing"
)

func TestArea(t *testing.T) {
	if got := (Rect{W: 3, H: 4}).Area(); got != 12 {
		t.Errorf("Rect{3,4}.Area() = %v; want 12", got)
	}
	got := Circle{R: 2}.Area()
	want := math.Pi * 4
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("Circle{2}.Area() = %v; want %v", got, want)
	}
}

func TestTotalArea(t *testing.T) {
	// 서로 다른 구체 타입을 같은 슬라이스에 담을 수 있다.
	shapes := []Shape{
		Rect{W: 3, H: 4}, // 12
		Rect{W: 1, H: 1}, // 1
		Circle{R: 1},     // π
	}
	want := 13 + math.Pi
	if got := TotalArea(shapes); math.Abs(got-want) > 1e-9 {
		t.Errorf("TotalArea = %v; want %v", got, want)
	}
	if got := TotalArea(nil); got != 0 {
		t.Errorf("TotalArea(nil) = %v; want 0", got)
	}
}

func TestTempString(t *testing.T) {
	tests := []struct {
		t    Temp
		want string
	}{
		{21.5, "21.5°C"},
		{0, "0.0°C"},
		{-3, "-3.0°C"},
		{100.04, "100.0°C"},
	}
	for _, tt := range tests {
		if got := tt.t.String(); got != tt.want {
			t.Errorf("Temp(%v).String() = %q; want %q", float64(tt.t), got, tt.want)
		}
	}

	// fmt가 알아서 String()을 불러준다 — 이게 인터페이스를 만족시킨 보상이다.
	if got := fmt.Sprintf("%v", Temp(36.5)); got != "36.5°C" {
		t.Errorf(`fmt.Sprintf("%%v", Temp(36.5)) = %q; want "36.5°C"`, got)
	}
}

func TestDescribe(t *testing.T) {
	tests := []struct {
		name string
		v    any
		want string
	}{
		{"nil", nil, "nil"},
		{"int", 42, "int: 42"},
		{"string", "hi", "string: hi"},
		{"Rect", Rect{W: 3, H: 4}, "shape: 12.00"},
		{"Circle", Circle{R: 0}, "shape: 0.00"},
		{"float64", 3.14, "unknown"},
		{"bool", true, "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Describe(tt.v); got != tt.want {
				t.Errorf("Describe(%#v) = %q; want %q", tt.v, got, tt.want)
			}
		})
	}
}

func TestAsCircle(t *testing.T) {
	if c, ok := AsCircle(Circle{R: 5}); !ok || c.R != 5 {
		t.Errorf("AsCircle(Circle{5}) = %v, %v; want {5}, true", c, ok)
	}
	if c, ok := AsCircle(Rect{W: 1, H: 1}); ok {
		t.Errorf("AsCircle(Rect{...}) = %v, %v; want {0}, false", c, ok)
	}
	// nil 인터페이스에 단언해도 panic이 나면 안 된다 (comma-ok 형태를 썼다면).
	if _, ok := AsCircle(nil); ok {
		t.Error("AsCircle(nil) ok = true; want false")
	}
}

func TestCountingWriter(t *testing.T) {
	var w CountingWriter // 제로값으로 바로 쓸 수 있어야 한다

	n, err := w.Write([]byte("hello"))
	if n != 5 || err != nil {
		t.Errorf("Write(\"hello\") = %d, %v; want 5, nil", n, err)
	}
	if w.N != 5 {
		t.Errorf("w.N = %d; want 5", w.N)
	}

	// 진짜 보상은 이것. 표준 라이브러리가 내 타입에 써준다.
	fmt.Fprintf(&w, "%s-%d", "abc", 42) // "abc-42" = 6바이트
	if w.N != 11 {
		t.Errorf("Fprintf 후 w.N = %d; want 11", w.N)
	}

	// io.Writer로서 넘길 수 있어야 한다.
	if _, err := fmt.Fprint(&w, ""); err != nil {
		t.Errorf("Fprint 실패: %v", err)
	}
}
