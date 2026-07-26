package basics

import "testing"

func TestSwap(t *testing.T) {
	a, b := Swap(1, 2)
	if a != 2 || b != 1 {
		t.Errorf("Swap(1, 2) = %d, %d; want 2, 1", a, b)
	}
	c, d := Swap(-5, 0)
	if c != 0 || d != -5 {
		t.Errorf("Swap(-5, 0) = %d, %d; want 0, -5", c, d)
	}
}

// 테이블 주도 테스트(table-driven test) — Go에서 가장 흔한 테스트 작성 방식입니다.
// 케이스를 슬라이스로 늘어놓고 루프를 돕니다. t.Run으로 각 케이스에 이름을 붙이면
// 실패했을 때 "TestSumTo/음수" 처럼 어느 케이스가 깨졌는지 바로 보입니다.
func TestSumTo(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want int
	}{
		{"0", 0, 0},
		{"음수", -5, 0},
		{"1", 1, 1},
		{"10", 10, 55},
		{"100", 100, 5050},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SumTo(tt.n); got != tt.want {
				t.Errorf("SumTo(%d) = %d; want %d", tt.n, got, tt.want)
			}
		})
	}
}

func TestGrade(t *testing.T) {
	tests := []struct {
		score int
		want  string
	}{
		{100, "A"},
		{90, "A"},
		{89, "B"},
		{80, "B"},
		{79, "C"},
		{70, "C"},
		{69, "F"},
		{0, "F"},
	}
	for _, tt := range tests {
		if got := Grade(tt.score); got != tt.want {
			t.Errorf("Grade(%d) = %q; want %q", tt.score, got, tt.want)
		}
	}
}

func TestMax(t *testing.T) {
	if _, ok := Max(); ok {
		t.Error("Max() ok = true; want false (인자가 없으면 값이 없다)")
	}
	if got, ok := Max(42); got != 42 || !ok {
		t.Errorf("Max(42) = %d, %v; want 42, true", got, ok)
	}
	if got, ok := Max(3, 1, 4, 1, 5, 9, 2, 6); got != 9 || !ok {
		t.Errorf("Max(3,1,4,1,5,9,2,6) = %d, %v; want 9, true", got, ok)
	}
	if got, ok := Max(-3, -1, -7); got != -1 || !ok {
		t.Errorf("Max(-3,-1,-7) = %d, %v; want -1, true", got, ok)
	}
	// 슬라이스를 가변 인자로 펼쳐 넘기는 문법: nums...
	nums := []int{10, 20, 5}
	if got, _ := Max(nums...); got != 20 {
		t.Errorf("Max(nums...) = %d; want 20", got)
	}
}

func TestDouble(t *testing.T) {
	n := 21
	Double(&n)
	if n != 42 {
		t.Errorf("Double(&n): n = %d; want 42", n)
	}
	Double(nil) // panic이 나면 안 됩니다
}

func TestCounter(t *testing.T) {
	c := Counter()
	for want := 1; want <= 3; want++ {
		if got := c(); got != want {
			t.Fatalf("%d번째 호출 = %d; want %d", want, got, want)
		}
	}
	// 두 카운터는 서로의 상태를 공유하면 안 됩니다.
	c2 := Counter()
	if got := c2(); got != 1 {
		t.Errorf("새 카운터의 첫 호출 = %d; want 1 (카운터끼리 상태를 공유하고 있습니다)", got)
	}
	if got := c(); got != 4 {
		t.Errorf("원래 카운터의 4번째 호출 = %d; want 4", got)
	}
}

func TestFizzBuzz(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{1, "1"},
		{2, "2"},
		{3, "Fizz"},
		{5, "Buzz"},
		{9, "Fizz"},
		{10, "Buzz"},
		{15, "FizzBuzz"},
		{30, "FizzBuzz"},
		{7, "7"},
		{100, "Buzz"},
	}
	for _, tt := range tests {
		if got := FizzBuzz(tt.n); got != tt.want {
			t.Errorf("FizzBuzz(%d) = %q; want %q", tt.n, got, tt.want)
		}
	}
}
