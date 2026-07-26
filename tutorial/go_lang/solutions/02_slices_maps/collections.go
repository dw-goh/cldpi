// Package collections — 02_slices_maps 정답.
package collections

import (
	"sort"
	"strings"
	"unicode/utf8"
)

// 문제 1 — 슬라이스 순회와 append
func Filter(s []int, keep func(int) bool) []int {
	// var out []int 는 nil 슬라이스다. len도 cap도 0이지만 append는 잘 동작한다.
	// make([]int, 0, len(s))로 시작하면 재할당을 줄일 수 있다 —
	// 결과가 클 것으로 예상되면 그렇게 하는 게 낫다.
	out := make([]int, 0, len(s))
	for _, v := range s {
		if keep(v) {
			out = append(out, v)
		}
	}
	// append의 반환값을 반드시 다시 대입해야 한다.
	// 재할당이 일어나면 out이 가리키는 배열 자체가 바뀌기 때문.
	return out
}

// 문제 2 — 슬라이스는 참조처럼 동작한다
func ModifyFirst(s []int) {
	if len(s) == 0 {
		return
	}
	// s는 값으로 복사되었지만 내부 배열 포인터는 호출한 쪽과 같다.
	// 그래서 원소를 고치면 호출한 쪽에 보인다.
	//
	// 반대로 s = append(s, 1) 은 호출한 쪽에 보이지 않는다.
	// len은 복사본의 필드이기 때문이다.
	s[0] = 100
}

// 문제 3 — append의 함정 ★
func SafeAppend(dst []int, v int) []int {
	// s[low:high:max] — 세 번째 인덱스가 cap을 정한다.
	// cap을 len과 같게 잘라두면 append가 여유 공간을 찾지 못하고
	// 반드시 새 배열을 할당한다. 원본 배열은 안전하다.
	//
	// 이걸 빼먹으면:
	//   base := make([]int, 3, 10); head := base[:2]
	//   append(head, 99) → base[2]를 덮어쓴다
	return append(dst[:len(dst):len(dst)], v)
}

// 문제 4 — copy
func Clone(s []int) []int {
	if s == nil {
		return nil
	}
	out := make([]int, len(s))
	// copy(dst, src)는 min(len(dst), len(src))개를 복사하고 그 수를 반환한다.
	// dst의 길이가 0이면 아무것도 복사되지 않으니 make의 길이를 잘 줘야 한다.
	copy(out, s)
	return out
	// 참고: Go 1.21+에는 slices.Clone(s)가 있다. 실무에서는 그걸 쓰면 된다.
}

// 문제 5 — 맵 조회 (comma-ok)
func Lookup(m map[string]int, k string) (int, bool) {
	// nil 맵을 읽는 것은 안전하다 (제로값 + false).
	// 쓰는 것만 panic이다.
	v, ok := m[k]
	return v, ok
}

// 문제 6 — 맵 만들기
func WordCount(s string) map[string]int {
	m := make(map[string]int)
	// strings.Fields는 연속된 공백을 하나로 취급하고 앞뒤 공백을 무시한다.
	// strings.Split(s, " ")를 쓰면 빈 문자열 조각이 섞여 들어온다.
	for _, w := range strings.Fields(s) {
		// 키가 없어도 된다. 없으면 제로값 0에서 시작한다.
		m[w]++
	}
	return m
}

// 문제 7 — 맵 순회 순서는 무작위다 ★
func SortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		// 값이 필요 없으면 키만 받는다. for k, _ := range m 이라고 쓰지 않는다.
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
	// Go 1.21+: slices.Sort(keys) 또는 slices.Sorted(maps.Keys(m))
}

// 문제 8 — 문자열은 바이트다 ★
func CountRunes(s string) int {
	// 방법 1 — 표준 라이브러리. 가장 빠르고 명확하다.
	return utf8.RuneCountInString(s)

	// 방법 2 — for range는 문자열을 rune 단위로 순회한다.
	// (i는 바이트 오프셋이므로 1씩 늘지 않는다는 점에 주의)
	//
	//	n := 0
	//	for range s {
	//	    n++
	//	}
	//	return n
	//
	// 방법 3 — len([]rune(s)). 동작하지만 슬라이스를 새로 할당하므로 낭비다.
	//
	// 하면 안 되는 것 — len(s). 그건 바이트 수다.
}
