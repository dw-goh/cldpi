package collections

import (
	"reflect"
	"testing"
)

func TestFilter(t *testing.T) {
	isEven := func(n int) bool { return n%2 == 0 }

	got := Filter([]int{1, 2, 3, 4, 5, 6}, isEven)
	want := []int{2, 4, 6}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Filter(1..6, isEven) = %v; want %v", got, want)
	}

	// 남는 게 없을 때 — nil이든 빈 슬라이스든 len이 0이면 통과.
	if got := Filter([]int{1, 3, 5}, isEven); len(got) != 0 {
		t.Errorf("Filter(홀수만, isEven) = %v; want 빈 슬라이스", got)
	}
	if got := Filter(nil, isEven); len(got) != 0 {
		t.Errorf("Filter(nil, isEven) = %v; want 빈 슬라이스", got)
	}
}

func TestModifyFirst(t *testing.T) {
	s := []int{1, 2, 3}
	ModifyFirst(s) // 포인터를 넘기지 않았는데도...
	if s[0] != 100 {
		t.Errorf("ModifyFirst 후 s = %v; want [100 2 3] (슬라이스는 내부 배열을 공유합니다)", s)
	}
	ModifyFirst(nil)     // panic이 나면 안 됩니다
	ModifyFirst([]int{}) // 빈 슬라이스도 마찬가지
}

func TestSafeAppend(t *testing.T) {
	// base의 cap은 10. head는 len=2지만 cap=10을 물려받는다.
	base := make([]int, 3, 10)
	base[0], base[1], base[2] = 1, 2, 777
	head := base[:2]

	got := SafeAppend(head, 99)

	if want := []int{1, 2, 99}; !reflect.DeepEqual(got, want) {
		t.Errorf("SafeAppend([1 2], 99) = %v; want %v", got, want)
	}
	if base[2] != 777 {
		t.Errorf("base[2] = %d; want 777 — 원본 배열을 덮어썼습니다. "+
			"그냥 append(dst, v)를 쓰면 여유 cap에 써버립니다. 3-인덱스 슬라이싱을 보세요", base[2])
	}
	// 호출한 쪽의 head는 그대로여야 한다.
	if len(head) != 2 {
		t.Errorf("len(head) = %d; want 2", len(head))
	}

	// cap에 여유가 없는 평범한 경우도 동작해야 한다.
	if got := SafeAppend([]int{1}, 2); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Errorf("SafeAppend([1], 2) = %v; want [1 2]", got)
	}
	if got := SafeAppend(nil, 1); !reflect.DeepEqual(got, []int{1}) {
		t.Errorf("SafeAppend(nil, 1) = %v; want [1]", got)
	}
}

func TestClone(t *testing.T) {
	src := []int{1, 2, 3}
	dst := Clone(src)

	if !reflect.DeepEqual(dst, src) {
		t.Fatalf("Clone(%v) = %v; 내용이 같아야 합니다", src, dst)
	}
	dst[0] = 999
	if src[0] != 1 {
		t.Errorf("복사본을 고쳤더니 원본도 바뀌었습니다: src = %v. "+
			"dst := src 는 복사가 아닙니다 — copy를 쓰세요", src)
	}
	if got := Clone(nil); got != nil {
		t.Errorf("Clone(nil) = %v; want nil", got)
	}
}

func TestLookup(t *testing.T) {
	m := map[string]int{"a": 1, "zero": 0}

	if v, ok := Lookup(m, "a"); v != 1 || !ok {
		t.Errorf(`Lookup(m, "a") = %d, %v; want 1, true`, v, ok)
	}
	// 값이 0으로 저장된 키 — "없음"과 구별되어야 한다.
	if v, ok := Lookup(m, "zero"); v != 0 || !ok {
		t.Errorf(`Lookup(m, "zero") = %d, %v; want 0, true`, v, ok)
	}
	if v, ok := Lookup(m, "없음"); v != 0 || ok {
		t.Errorf(`Lookup(m, "없음") = %d, %v; want 0, false`, v, ok)
	}
	if _, ok := Lookup(nil, "a"); ok {
		t.Error("Lookup(nil, ...) ok = true; want false (nil 맵 읽기는 panic이 아닙니다)")
	}
}

func TestWordCount(t *testing.T) {
	got := WordCount("go go python go")
	want := map[string]int{"go": 3, "python": 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("WordCount = %v; want %v", got, want)
	}
	// 공백이 여러 개거나 앞뒤에 붙어도 빈 단어가 생기면 안 된다.
	got = WordCount("  a   b  ")
	want = map[string]int{"a": 1, "b": 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf(`WordCount("  a   b  ") = %v; want %v (strings.Split이 아니라 strings.Fields)`, got, want)
	}
	if got := WordCount(""); len(got) != 0 {
		t.Errorf(`WordCount("") = %v; want 빈 맵`, got)
	}
}

func TestSortedKeys(t *testing.T) {
	m := map[string]int{"banana": 2, "apple": 1, "cherry": 3}
	want := []string{"apple", "banana", "cherry"}

	// 맵 순회 순서는 실행마다 달라집니다. 여러 번 돌려서 항상 같은지 확인합니다.
	for i := 0; i < 20; i++ {
		if got := SortedKeys(m); !reflect.DeepEqual(got, want) {
			t.Fatalf("SortedKeys = %v; want %v (정렬하지 않으면 매번 순서가 바뀝니다)", got, want)
		}
	}
	if got := SortedKeys(map[string]int{}); len(got) != 0 {
		t.Errorf("SortedKeys(빈 맵) = %v; want 빈 슬라이스", got)
	}
}

func TestCountRunes(t *testing.T) {
	tests := []struct {
		s    string
		want int
	}{
		{"", 0},
		{"hi", 2},
		{"안녕", 2},    // len은 6
		{"안녕hi", 4},  // len은 8
		{"héllo", 5}, // é는 2바이트, len은 6
		{"🇰🇷", 2},    // 국기는 코드포인트 2개로 이루어져 있습니다
	}
	for _, tt := range tests {
		if got := CountRunes(tt.s); got != tt.want {
			t.Errorf("CountRunes(%q) = %d; want %d (len(%q) = %d — 바이트 수와 다릅니다)",
				tt.s, got, tt.want, tt.s, len(tt.s))
		}
	}
}
