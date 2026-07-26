// Package collections — 슬라이스, 맵, 문자열.
//
//	go test ./02_slices_maps/
//
// Go에서 가장 많이 헷갈리는 부분입니다. 천천히 하세요.
package collections

import (
    "strings"
    "sort"
)

// 문제 1 — 슬라이스 순회와 append
//
// keep(v)가 true인 원소만 모아 반환하세요.
// s가 nil이거나 남는 게 없으면 길이 0인 슬라이스를 반환합니다.
//
// nil 슬라이스에 append하는 것은 완전히 합법입니다.
// var out []int 로 시작해서 바로 append하면 됩니다. make가 필요 없습니다.
func Filter(s []int, keep func(int) bool) []int {
	// TODO
        var out []int
        for _, v := range s {
            if keep(v) {
                out = append(out, v)
            }
        }
        return out	
}
// i := range s -> 인덱스
// i, v := range s -> 인덱스, 값


// 문제 2 — 슬라이스는 참조처럼 동작한다
//
// s가 비어있지 않으면 첫 원소를 100으로 바꾸세요.
//
// 슬라이스는 값 타입이지만 내부에 배열 포인터를 들고 있습니다.
// 함수에 넘겨도 "복사"되는 건 (포인터, len, cap) 세 필드뿐이고,
// 원소는 호출한 쪽과 공유됩니다. 그래서 포인터를 넘기지 않아도
// 호출한 쪽의 데이터가 바뀝니다.
func ModifyFirst(s []int) {
	// TODO
        if len(s) != 0 {
            s[0] = 100
        }
}

// 문제 3 — append의 함정 ★
//
// dst에 v를 덧붙인 새 슬라이스를 반환하되,
// **dst가 공유하는 원본 배열은 절대 건드리지 않아야** 합니다.
//
// 왜 이게 문제인가:
//
//	base := make([]int, 3, 10)   // len=3, cap=10
//	head := base[:2]             // len=2, cap=10  ← cap을 그대로 물려받는다!
//	head = append(head, 99)      // 여유 공간이 있으니 재할당 없이
//	                             // base[2]에 99를 덮어쓴다. base가 오염된다.
//
// 힌트: 3-인덱스 슬라이싱 s[low:high:max]는 cap을 잘라냅니다.
// cap을 len과 같게 만들면 append가 반드시 새 배열을 할당합니다.
func SafeAppend(dst []int, v int) []int {
	// TODO
        tmp := dst[:len(dst):len(dst)]
	return append(tmp, v)
}

// 문제 4 — copy
//
// s와 내용은 같지만 배열을 공유하지 않는 새 슬라이스를 반환하세요.
// s가 nil이면 nil을 반환합니다.
//
// 힌트: make로 같은 길이의 슬라이스를 만들고 copy(dst, src)를 씁니다.
// copy는 복사한 원소 개수를 반환하고, 짧은 쪽 길이만큼만 복사합니다.
func Clone(s []int) []int {
	// TODO
        if s == nil { return nil }
        ret := make([]int, len(s), cap(s))
        copy(ret, s)

        return ret
}

// 문제 5 — 맵 조회 (comma-ok)
//
// m에서 k를 찾아 (값, 존재여부)를 반환하세요.
//
// 주의: 없는 키를 조회해도 panic이 나지 않습니다. 값 타입의 제로값이 나옵니다.
// m["없는키"]는 0입니다. "0이 저장된 것"과 "없는 것"을 구분하려면
// 반드시 두 번째 반환값 ok를 봐야 합니다.
func Lookup(m map[string]int, k string) (int, bool) {
	// TODO
        v, ok := m[k]
	return v, ok
}

// 문제 6 — 맵 만들기
//
// 공백으로 나뉜 단어들의 등장 횟수를 세세요.
//
//	WordCount("go go python") → map[string]int{"go": 2, "python": 1}
//
// 빈 문자열이면 길이 0인 맵을 반환합니다(nil 맵도 통과합니다만,
// nil 맵에 쓰기를 하면 panic이라는 걸 기억해두세요 — 읽기는 됩니다).
//
// 힌트: strings.Fields는 임의의 공백으로 잘라주고 빈 조각을 만들지 않습니다.
// m[w]++ 는 키가 없어도 동작합니다 (0에서 시작).
func WordCount(s string) map[string]int {
	// TODO
        m := make(map[string]int)
        f := strings.Fields(s)
        for _, v := range f {
            m[v] += 1
        }
	return m
}

// 문제 7 — 맵 순회 순서는 무작위다 ★
//
// m의 키를 사전순으로 정렬해 반환하세요.
// m이 비었으면 길이 0인 슬라이스를 반환합니다.
//
// Go는 맵 순회 순서를 **의도적으로 무작위화**합니다. 순서에 의존하는 코드를
// 짜지 못하게 하려는 설계입니다. 실행할 때마다 순서가 달라집니다.
// 정해진 순서가 필요하면 키를 뽑아서 직접 정렬해야 합니다.
//
// 힌트: sort.Strings(keys) 또는 slices.Sort(keys)
func SortedKeys(m map[string]int) []string {
	// TODO
        keys := make([]string, 0, len(m))
        for k := range m {
            keys = append(keys, k)
        }
        sort.Strings(keys)
	return keys 
}

// 문제 8 — 문자열은 바이트다 ★
//
// s에 들어있는 문자(유니코드 코드포인트)의 개수를 반환하세요.
//
//	CountRunes("hi")     → 2   (len도 2)
//	CountRunes("안녕hi")  → 4   (len은 8! 한글 한 글자가 UTF-8로 3바이트)
//
// Go 문자열은 읽기 전용 바이트 슬라이스입니다. len(s)는 바이트 수입니다.
// s[0]은 문자가 아니라 byte(uint8)입니다.
//
// 힌트: for range로 문자열을 돌면 바이트가 아니라 rune 단위로 나옵니다.
// utf8.RuneCountInString(s)를 써도 됩니다.
func CountRunes(s string) int {
	// TODO
        count := 0
        for range s {
            count += 1
        }
	return count 
}
