# 05 — 에러 처리와 defer

**Go에는 예외가 없다.** try/catch도, `throws`도 없다. 에러는 그냥 반환값이다.

이 장의 패턴은 앞으로 프로젝트 전체에서 매일 쓰게 된다.

```bash
go test -v ./05_errors_defer/
```

---

## error는 인터페이스일 뿐이다

```go
type error interface {
    Error() string
}
```

이게 전부다. `Error() string` 메서드가 있으면 error다.

```go
f, err := os.Open("config.json")
if err != nil {
    return err
}
```

Python이라면 `try/except`, C라면 `if (fd < 0) { perror(...) }`로 쓰던 자리다. Go는 이걸 **모든 함수 호출마다** 명시적으로 쓴다.

처음엔 장황해 보인다. 실제로 장황하다. 대신:
- 어떤 줄에서 에러가 날 수 있는지 코드만 봐도 안다
- 조용히 무시되는 에러가 없다 (무시하려면 `_`라고 명시해야 한다)
- 스택 되감기(unwinding)가 없으니 제어 흐름이 단순하다

### 에러 만들기

```go
errors.New("connection closed")                  // 고정 메시지
fmt.Errorf("read chunk %d: %w", id, err)         // 값 끼우기 + 감싸기
```

**관례:** 메시지는 소문자로 시작하고 마침표를 안 찍는다. 호출한 쪽이 계속 앞에 문맥을 덧붙이기 때문이다.

```
open meta.db: read chunk 42: unexpected EOF
```

---

## 센티널 에러와 감싸기(wrapping) ★

```go
var ErrNotFound = errors.New("not found")     // 센티널 — 미리 만들어둔 에러 값
```

호출한 쪽에서 "이게 그 에러인가?"를 판단할 수 있게 해준다. `io.EOF`, `sql.ErrNoRows`, `os.ErrNotExist`가 전부 이 패턴이다.

문맥을 덧붙이면서도 원본을 잃지 않으려면 `%w`를 쓴다.

```go
return fmt.Errorf("find user %q: %w", name, ErrNotFound)
//                                    ↑ wrap. %v로 쓰면 원본 값이 사라진다
```

그러면 호출한 쪽에서:

```go
if errors.Is(err, ErrNotFound) { ... }   // 감싼 사슬을 끝까지 따라가며 == 비교

var pe *ParseError
if errors.As(err, &pe) { ... }           // 사슬에서 특정 타입을 꺼낸다
```

| | 묻는 것 |
|---|---|
| `err == ErrNotFound` | 정확히 이 값인가? **감싸면 실패한다. 쓰지 말 것** |
| `errors.Is(err, ErrNotFound)` | 사슬 어딘가에 이 값이 있는가? |
| `errors.As(err, &pe)` | 사슬 어딘가에 이 타입이 있는가? 있으면 꺼내줘 |

---

## defer

```go
defer f.Close()
```

함수가 끝날 때 실행된다. **어떻게 끝나든** — 정상 return, 에러 return, panic 전부.

세 가지 규칙만 기억하면 된다.

**1. 인자는 defer를 쓰는 순간 평가된다.**

```go
i := 0
defer fmt.Println(i)   // 0을 출력한다. 아래에서 i를 바꿔도 소용없다
i = 42
```

나중 값을 쓰고 싶으면 클로저로 감싼다: `defer func() { fmt.Println(i) }()`

**2. LIFO 순서로 실행된다.**

```go
for i := 0; i < 3; i++ {
    defer fmt.Println(i)
}
// 출력: 2 1 0
```

**3. return 문 뒤에 실행되므로, 이름 붙은 반환값을 수정할 수 있다.**

```go
func f() (err error) {
    defer func() { err = fmt.Errorf("wrapped: %w", err) }()
    return errors.New("original")
}
// 반환값: "wrapped: original"
```

이 세 번째 성질이 문제 4와 문제 6의 핵심이다.

### 관용구: 여는 줄 바로 다음에 defer

```go
f, err := os.Open(path)
if err != nil {
    return err
}
defer f.Close()      // ← 여기. 아래에 return이 몇 개든 신경 안 써도 된다
```

C의 `goto cleanup` 패턴을 언어 기능으로 만든 것이다.

> ⚠️ **루프 안에서 defer를 쓰면 안 된다.** defer는 *함수*가 끝날 때 실행되지 블록이 끝날 때가 아니다. 파일 1000개를 여는 루프에서 `defer f.Close()`를 쓰면 1000개가 전부 열린 채로 쌓인다. 루프 몸통을 별도 함수로 빼야 한다.

---

## panic / recover

**Go에서 panic은 예외가 아니다.** "여기까지 왔다면 프로그램이 잘못된 것"에만 쓴다. nil 역참조, 슬라이스 범위 초과, 0으로 나누기 같은 것들이 panic을 낸다.

`recover()`는 defer 안에서만 의미가 있고, panic을 멈추고 넘겨진 값을 반환한다.

실무에서 쓰는 곳은 사실상 하나다: **서버가 요청 하나 때문에 통째로 죽는 걸 막을 때.**

```go
go func() {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("connection handler panic: %v", r)
        }
    }()
    handleConn(c)
}()
```

M0의 8장에서 접속마다 goroutine을 띄우게 되는데, 그중 하나가 panic하면 **프로세스 전체가 죽는다.** goroutine의 panic은 다른 goroutine이 잡을 수 없다. 그래서 이 패턴이 필요하다.

---

## 문제

| # | 대상 | 배우는 것 |
|---|---|---|
| 1 | `Divide` | `errors.New` |
| 2 | `ParseError`, `ParseAge` | 커스텀 에러 타입 |
| 3 | `FindUser`, `IsNotFound`, `ReasonOf` | ★ `%w`, `errors.Is`, `errors.As` |
| 4 | `DeferOrder` | ★ defer LIFO + 이름 붙은 반환값 |
| 5 | `Process` | defer로 반드시 정리 |
| 6 | `SafeDiv` | panic/recover |
| 7 | `MaybeErr` | ★★ typed nil 함정 |

---

## 걸려 넘어지기 쉬운 곳

**문제 7이 이 장의 하이라이트다.** Go에서 가장 유명한 함정이고, 경력자도 물린다.

```go
func bad() error {
    var e *MyErr = nil
    return e          // err != nil 이 된다!
}
```

인터페이스 값은 **(타입, 값) 쌍**이다. `nil`인 `*MyErr`를 `error`에 넣으면 `(타입=*MyErr, 값=nil)`이 되는데, 진짜 nil 인터페이스는 `(타입=nil, 값=nil)`이다. 둘은 다르다.

**규칙:** 에러를 담는 변수의 타입은 **항상 `error`**로 쓴다. 구체 타입 포인터를 중간에 두지 않는다.

```go
func good() error {
    if fail {
        return &MyErr{}
    }
    return nil          // nil 리터럴을 직접 반환
}
```

**문제 4 — `return nil`을 지우면 안 된다.** `return nil`이 먼저 `out = nil`을 하고, 그 다음에 defer들이 돌면서 append한다. 순서를 이해하는 게 문제의 핵심이다.

**에러를 무시할 때는 명시적으로.**

```go
_ = f.Close()   // 무시한다는 걸 코드로 남긴다
f.Close()       // 반환값을 안 쓰는 건 컴파일 에러가 아니지만, 린터가 잡는다
```

---

## 더 볼 것

- [Go 블로그 — Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors) — `%w`, `Is`, `As`가 왜 생겼는지
- [Effective Go — Errors](https://go.dev/doc/effective_go#errors)
- `go doc errors`
