# 03 — 구조체와 메서드

Go에는 클래스가 없다. 구조체와, 그 구조체에 붙는 함수(메서드)가 있을 뿐이다.

```bash
go test -v ./03_structs/
```

---

## 메서드는 그냥 리시버가 붙은 함수다

```go
func (p Point) Add(q Point) Point { ... }
//   └─ 리시버 ─┘
```

C에서 이렇게 쓰던 것과 같다:

```c
Point point_add(Point p, Point q);
p = point_add(p, q);          // C
p = p.Add(q)                  // Go
```

차이는 **어떤 타입에든 붙일 수 있다는 것**이다. 구조체가 아니어도 된다.

```go
type Celsius float64
func (c Celsius) String() string { ... }   // 합법
```

단, **자기 패키지에서 선언한 타입에만** 붙일 수 있다. `int`나 남의 패키지 타입에 메서드를 추가할 수는 없다.

---

## 값 리시버 vs 포인터 리시버 ★

이 장에서 가장 중요한 구분이다.

```go
func (c Counter) Inc()  { c.n++ }   // 복사본을 증가시킨다. 아무 일도 안 일어난다
func (c *Counter) Inc() { c.n++ }   // 원본을 증가시킨다
```

**포인터 리시버를 쓰는 경우:**
1. 리시버를 수정해야 할 때
2. 구조체가 커서 복사가 아까울 때
3. `sync.Mutex`처럼 복사하면 안 되는 것을 품고 있을 때

**규칙:** 한 타입의 메서드는 값/포인터 중 **하나로 통일한다.** 하나라도 포인터 리시버가 필요하면 전부 포인터로 맞추는 게 관례다. (이 장의 `Counter`는 대비를 보여주려고 일부러 섞어놨다.)

### 호출할 때는 신경 안 써도 된다

```go
p := Point{1, 2}
p.Scale(2)      // Go가 (&p).Scale(2)로 바꿔준다

pp := &Point{1, 2}
pp.Add(q)       // Go가 (*pp).Add(q)로 바꿔준다
```

단, **주소를 얻을 수 있는 값(addressable)**이어야 한다. 맵의 원소는 주소를 얻을 수 없어서 `m["k"].Scale(2)`는 컴파일 에러다. 그래서 맵에 구조체를 넣을 때는 보통 포인터(`map[string]*Counter`)로 넣는다.

---

## 제로값이 쓸모있게 (useful zero value)

Go는 모든 변수를 제로값으로 초기화한다. 숫자는 0, 문자열은 `""`, 포인터/슬라이스/맵/채널/인터페이스는 `nil`, 구조체는 모든 필드가 제로값.

C의 초기화 안 된 스택 변수 같은 쓰레기 값은 절대 없다.

좋은 Go 타입은 **선언만 해도 쓸 수 있게** 설계한다.

```go
var mu sync.Mutex     // 바로 Lock() 가능
var buf bytes.Buffer  // 바로 Write() 가능
var s []int           // 바로 append() 가능
```

`NewXxx()` 생성자는 **꼭 필요할 때만** 만든다. 검증이 필요하거나(문제 3), 맵/채널 필드를 초기화해야 하거나 할 때다.

---

## 임베딩은 상속이 아니다

```go
type Dog struct {
    Animal        // 필드 이름이 없다 = 임베딩
    Breed string
}
```

`Animal`의 필드와 메서드가 `Dog`에서 바로 보인다(승격). 하지만 **상속이 아니다.**

```go
var a Animal = d       // ✗ 컴파일 에러. Dog는 Animal이 아니다
a = d.Animal           // ○ 명시적으로 꺼내야 한다
```

C++의 상속은 "is-a"지만, Go의 임베딩은 "has-a에 이름 붙이기 귀찮아서 생략한 것"에 가깝다. 그래서 다형성은 임베딩이 아니라 **인터페이스**로 한다 (다음 장).

임베딩을 실무에서 쓰는 대표적인 예:

```go
type LoggingConn struct {
    net.Conn      // Read/Write/Close를 전부 물려받는다
}
func (c LoggingConn) Read(p []byte) (int, error) {   // 하나만 가로챈다
    n, err := c.Conn.Read(p)
    log.Printf("read %d bytes", n)
    return n, err
}
```

---

## 문제

| # | 대상 | 배우는 것 |
|---|---|---|
| 1 | `Point.Add` | 값 리시버 |
| 2 | `Point.Scale` | ★ 포인터 리시버 |
| 3 | `NewRect`, `Rect.Area` | 생성자 패턴 |
| 4 | `Counter.Inc/Value` | 쓸모있는 제로값 |
| 5 | `Dog.Speak` | 임베딩, 메서드 가리기 |
| 6 | `CountPoints` | 구조체를 맵 키로 |

---

## 걸려 넘어지기 쉬운 곳

**문제 2, 4 — 리시버를 값으로 쓰면 컴파일은 되는데 테스트가 실패한다.**
Go 컴파일러는 이걸 잡아주지 않는다. "고쳤는데 왜 안 바뀌지?" 싶으면 리시버에 `*`가 있는지 먼저 본다.

**문제 6 — 구조체 리터럴 안에서 타입 이름 생략.**
`map[Point]int{{1,2}: 3}` 처럼 `Point{1,2}`의 타입 이름을 생략할 수 있다. 슬라이스도 마찬가지로 `[]Point{{1,2},{3,4}}`.

**`%v`와 `%+v`.**
디버깅할 때 `fmt.Printf("%+v\n", r)`을 쓰면 `{W:3 H:4}`처럼 필드 이름이 같이 나온다. `%v`는 `{3 4}`. `%#v`는 `shapes.Rect{W:3, H:4}`.

---

## 더 볼 것

- [Effective Go — Methods](https://go.dev/doc/effective_go#methods)
- [Go FAQ — 값 리시버와 포인터 리시버 중 뭘 쓰나](https://go.dev/doc/faq#methods_on_values_or_pointers)
