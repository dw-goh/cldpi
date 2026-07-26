# 04 — 인터페이스

Go의 인터페이스는 다른 언어와 결정적으로 다르다. **구현한다고 선언하지 않는다.**

```bash
go test -v ./04_interfaces/
```

---

## 암묵적 만족 (implicit satisfaction)

```java
class Rect implements Shape { ... }   // Java — 명시적
```
```go
type Rect struct{ W, H float64 }
func (r Rect) Area() float64 { ... }  // Go — 이걸로 끝. Shape를 언급조차 안 한다
```

`Area() float64` 메서드가 있으면 `Rect`는 **이미** `Shape`다. 인터페이스가 나중에 정의돼도 상관없고, 남의 패키지에 있어도 상관없다.

이게 왜 중요한가:

- **인터페이스를 쓰는 쪽에서 정의한다.** 라이브러리가 "이 인터페이스를 구현하세요"라고 강요하지 않는다. 내가 쓸 때 필요한 만큼만 인터페이스를 선언하면 된다.
- **기존 타입을 나중에 끼워 맞출 수 있다.** 표준 라이브러리 타입이 내가 방금 만든 인터페이스를 만족하는 일이 흔하다.

대신 실수로 만족시키지 못하는 경우(오타, 시그니처 불일치)를 컴파일러가 늦게 잡는다. 그래서 이런 관용구를 쓴다:

```go
var _ Shape = Rect{}                    // Rect가 Shape가 아니면 여기서 컴파일 에러
var _ io.Writer = (*CountingWriter)(nil) // 포인터 리시버라면 이 형태
```

`_`는 "버리는 변수"다. 값을 만들지 않으려고 `nil`을 타입 변환해 넣는다.

---

## 인터페이스는 작게

표준 라이브러리의 핵심 인터페이스는 전부 메서드 1개다.

```go
type error interface     { Error() string }
type Stringer interface  { String() string }
type Reader interface    { Read(p []byte) (n int, err error) }
type Writer interface    { Write(p []byte) (n int, err error) }
```

> "The bigger the interface, the weaker the abstraction." — Rob Pike

`io.Reader` / `io.Writer`가 특히 중요하다. 이 두 개 덕분에 파일, 네트워크 연결, 메모리 버퍼, gzip 스트림, 암호화 스트림이 전부 같은 함수에 들어간다.

```go
io.Copy(dst, src)   // dst가 파일이든 소켓이든 상관없다
```

**`net.Conn`도 `io.Reader`이자 `io.Writer`다.** 8장에서 TCP 연결에 `fmt.Fprintf`를 쓰게 되는 이유가 이것이다.

---

## 인터페이스 값은 (타입, 값) 쌍이다

```
var s Shape = Rect{3, 4}

s ─┬─ 타입: Rect
   └─ 값 : {3, 4}
```

여기서 두 가지 연산이 나온다.

### 타입 단언 (type assertion)

```go
c := s.(Circle)        // 실패하면 panic
c, ok := s.(Circle)    // 실패하면 (제로값, false) — 이걸 쓴다
```

### 타입 스위치 (type switch)

```go
switch x := v.(type) {
case nil:            // v가 nil 인터페이스
case int:            // x는 int
case string:         // x는 string
case Shape:          // x는 Shape
default:             // x는 원래 타입 그대로
}
```

**case 순서가 중요하다.** 위에서부터 검사하므로 구체적인 걸 먼저 둔다. `Rect`는 `Shape`이기도 하므로 `case Shape`를 `case Rect`보다 위에 두면 `Rect`는 절대 안 걸린다.

타입 스위치를 자주 쓰고 있다면 대개 설계가 잘못된 신호다. 메서드로 풀 수 있는지 먼저 본다.

---

## `any`와 `interface{}`

```go
var v any          // Go 1.18+
var v interface{}  // 같은 것. any는 별칭이다
```

메서드를 하나도 요구하지 않는 인터페이스이므로 모든 타입이 만족한다. C의 `void*`와 비슷하지만 **타입 정보를 같이 들고 다녀서** 런타임에 뭐가 들었는지 알 수 있다.

---

## 문제

| # | 대상 | 배우는 것 |
|---|---|---|
| 1 | `Rect.Area`, `Circle.Area` | 암묵적 인터페이스 만족 |
| 2 | `TotalArea` | 인터페이스를 받는 함수 = 다형성 |
| 3 | `Temp.String` | `fmt.Stringer` |
| 4 | `Describe` | 타입 스위치 |
| 5 | `AsCircle` | 타입 단언 comma-ok |
| 6 | `CountingWriter.Write` | ★ `io.Writer` 구현 |

---

## 걸려 넘어지기 쉬운 곳

**문제 3 — `String()` 안에서 자기 자신을 `%v`로 출력하면 무한 재귀다.**

```go
func (t Temp) String() string {
    return fmt.Sprintf("%v°C", t)          // ✗ String()이 String()을 부른다 → 스택 오버플로
    return fmt.Sprintf("%.1f°C", float64(t)) // ○ 기반 타입으로 변환
}
```

`go vet`이 이 실수를 잡아준다. `go vet ./...`를 습관적으로 돌리자.

**문제 6 — 리시버가 `*CountingWriter`다.**
`N`을 수정해야 하므로 포인터 리시버여야 한다. 그래서 `fmt.Fprintf(&w, ...)`처럼 **주소를 넘긴다.** `fmt.Fprintf(w, ...)`라고 쓰면 "CountingWriter는 io.Writer를 구현하지 않는다"는 컴파일 에러가 난다.

포인터 리시버로 메서드를 정의하면 **`*T`만 인터페이스를 만족하고 `T`는 만족하지 않는다.** 처음 만나면 헷갈리는 규칙이니 에러 메시지를 잘 읽자:

```
cannot use w (variable of type CountingWriter) as io.Writer value:
    CountingWriter does not implement io.Writer (method Write has pointer receiver)
```

---

## 더 볼 것

- [Effective Go — Interfaces](https://go.dev/doc/effective_go#interfaces)
- `go doc io.Writer` / `go doc io.Reader`
- [Go Proverbs](https://go-proverbs.github.io/) — 짧은 격언 모음. 인터페이스 관련이 많다
