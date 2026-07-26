# 01 — 기초 문법

C/Python을 알고 있다면 Go 문법 자체는 하루면 읽힌다. 문제는 "다른 언어와 미묘하게 다른 지점"이다. 이 장은 그 지점만 모았다.

```bash
go test -v ./01_basics/
```

---

## 먼저 알아둘 것

### 변수 선언

```go
var x int = 10   // 정석
var x = 10       // 타입 추론
x := 10          // 짧은 선언 — 함수 안에서만 쓸 수 있다
```

실무에서는 `:=`가 압도적으로 많다. 다만 **선언했으면 반드시 써야 한다.** 안 쓰면 컴파일 에러다. C의 `-Wunused-variable` 경고가 아니라 에러다. 처음엔 짜증나지만 죽은 코드가 안 쌓인다.

`import`도 마찬가지다. 안 쓰는 import는 컴파일 에러다.

### 타입은 뒤에 온다

```c
int  x;              // C
int  add(int a, int b);
```
```go
var x int            // Go
func add(a, b int) int
```

C의 `int (*fp)(char*, int)` 같은 선언을 왼쪽 오른쪽으로 읽어야 했던 걸 없애려는 의도다. Go에서는 항상 왼쪽에서 오른쪽으로 읽힌다.

### 대문자 = public

```go
func Encode() {}   // 패키지 밖에서 쓸 수 있다 (export됨)
func decode() {}   // 이 패키지 안에서만
```

`public`/`private` 키워드가 없다. **이름의 첫 글자가 접근 제어다.** 그래서 테스트가 호출하는 함수는 전부 대문자로 시작한다.

### 세미콜론과 중괄호

세미콜론은 컴파일러가 자동으로 넣는다. 대신 **여는 중괄호를 다음 줄에 두면 안 된다.**

```go
func f()
{          // 컴파일 에러
}
```

포맷은 논쟁거리가 아니다. `gofmt`(= `go fmt ./...`)가 정답을 강제한다. 탭 들여쓰기, 정렬, 줄바꿈 전부 자동이다.

### if

```go
if err != nil {        // 괄호 없음, 중괄호 필수
    return err
}

if v, ok := m["key"]; ok {   // 초기화문을 붙일 수 있다. v와 ok는 if 안에서만 산다
    use(v)
}
```

### 반복문은 for 하나뿐

```go
for i := 0; i < 10; i++ {}   // C 스타일
for x < 10 {}                // while
for {}                       // 무한 루프
for i, v := range slice {}   // foreach
```

---

## 문제

`basics.go`의 TODO를 위에서부터 채운다.

| # | 함수 | 배우는 것 |
|---|---|---|
| 1 | `Swap` | 다중 반환값 |
| 2 | `SumTo` | for가 유일한 반복문 |
| 3 | `Grade` | switch (break 없음) |
| 4 | `Max` | 가변 인자, "값이 없음"을 bool로 표현 |
| 5 | `Double` | 포인터 |
| 6 | `Counter` | 클로저 — 함수는 값이다 |
| 7 | `FizzBuzz` | 조합 + `strconv.Itoa` |

---

## 걸려 넘어지기 쉬운 곳

**문제 3 — switch에 break를 쓰지 않는다.**
C와 달리 case는 자동으로 빠져나온다. 일부러 다음 case로 넘어가려면 `fallthrough`를 명시해야 하는데, 실무에서 거의 안 쓴다.

**문제 4 — `(int, bool)` 반환은 Go의 관용구다.**
C처럼 `-1`을 실패 표시로 쓰거나 Python처럼 `None`을 반환하는 대신, Go는 `(값, 성공여부)`를 반환한다. 맵 조회 `v, ok := m[k]`, 채널 수신 `v, ok := <-ch`, 타입 단언 `v, ok := x.(T)` 전부 같은 모양이다.

**문제 6 — `Counter()`가 반환하는 건 함수다.**
`func() int`는 "인자 없고 int를 반환하는 함수" 타입이다. 클로저가 붙잡은 변수는 힙으로 옮겨가므로(escape analysis) C에서처럼 스택 위 지역변수 주소를 반환하는 위험이 없다.

**문제 7 — `string(7)`은 `"7"`이 아니다.**
`string(int)` 변환은 정수를 **유니코드 코드포인트**로 해석한다. `string(65)`는 `"A"`다. 숫자를 문자열로 바꾸려면 반드시 `strconv.Itoa(7)`을 쓴다. (요즘 Go는 이 실수에 `go vet` 경고를 띄운다.)

---

## 더 볼 것

```bash
go doc strconv.Itoa      # 터미널에서 표준 라이브러리 문서 보기
go doc strconv           # 패키지 전체 목록
```

- [Tour of Go — Basics](https://go-tour-ko.appspot.com/basics/1)
