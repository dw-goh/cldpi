// Package errs — 에러 처리와 defer.

//	go test ./05_errors_defer/
//
// Go에는 예외(exception)가 없습니다. 에러는 그냥 반환값입니다.
// 이 장은 프로젝트 전체에서 계속 쓰게 될 패턴들입니다.
package errs

import "errors"
import "fmt"
import "strconv"

// 문제 1 — 기본 에러 반환
//
// a를 b로 나눈 몫을 반환하세요. b가 0이면 (0, 에러)를 반환합니다.
//
// error는 인터페이스입니다:
//
//	type error interface { Error() string }
//
// 가장 간단한 에러 만들기: errors.New("메시지")
// 값을 끼워 넣고 싶으면: fmt.Errorf("cannot divide %d by zero", a)
//
// 관례: 에러 메시지는 소문자로 시작하고 마침표를 찍지 않습니다.
// 호출한 쪽이 앞뒤에 문맥을 덧붙이기 때문입니다.
func Divide(a, b int) (int, error) {
	// TODO
        if b==0 { return 0, fmt.Errorf("cannot divide %d by zero", a) }
        return a / b, nil
}

// ParseError는 파싱 실패를 나타내는 커스텀 에러 타입입니다.
type ParseError struct {
	Input  string
	Reason string
}

// 문제 2 — 커스텀 에러 타입
//
// Error()는 `parse "abc": not a number` 형태를 반환해야 합니다.
//
//	(&ParseError{Input: "abc", Reason: "not a number"}).Error()
//	  → `parse "abc": not a number`
//
// 힌트: %q는 문자열을 따옴표로 감싸줍니다. fmt.Sprintf("parse %q: %s", ...)
//
// Error() string 메서드를 만들면 이 타입은 error 인터페이스를 만족합니다.
// 리시버가 *ParseError인 점에 주의 — 에러 타입은 포인터로 쓰는 게 관례입니다.
func (e *ParseError) Error() string {
	// TODO
	return fmt.Sprintf("parse %q: %s", e.Input, e.Reason)
}

// 문제 2 (계속) — ParseAge
//
// s를 정수로 파싱해서 반환하세요.
//   - 숫자가 아니면 &ParseError{Input: s, Reason: "not a number"}
//   - 음수면 &ParseError{Input: s, Reason: "negative"}
//   - 성공하면 (값, nil)
//
// 힌트: strconv.Atoi(s)가 (int, error)를 반환합니다.
func ParseAge(s string) (int, error) {
	// TODO
        v, e := strconv.Atoi(s)
        if e != nil { return 0, &ParseError{Input: s, Reason: "not a number"} }
        if v < 0 { return 0, &ParseError{Input: s, Reason: "negative"} }
	return v, e
}

// ErrNotFound는 센티널(sentinel) 에러입니다.
//
// 패키지 밖에서 errors.Is로 비교할 수 있도록 미리 만들어 둔 에러 값입니다.
// 표준 라이브러리의 io.EOF, sql.ErrNoRows가 같은 패턴입니다.
var ErrNotFound = errors.New("not found")

// 문제 3 — 에러 감싸기(wrapping) ★
//
// db에서 name을 찾아 반환하세요. 없으면 ErrNotFound를 감싼 에러를 반환합니다.
//
//	FindUser(db, "kim") → ("", `find user "kim": not found`)
//
// fmt.Errorf의 %w 동사가 에러를 "감쌉니다".
// %v로 쓰면 메시지만 남고 원래 에러 값은 사라집니다.
// %w로 쓰면 문맥을 덧붙이면서도 errors.Is로 원본을 찾아낼 수 있습니다.
//
//	fmt.Errorf("find user %q: %w", name, ErrNotFound)
func FindUser(db map[string]string, name string) (string, error) {
	// TODO
        v, ok := db[name]
        if !ok { return "", fmt.Errorf("find user %q: %w", name, ErrNotFound) }
	return v, nil
}

// 문제 3 (계속) — errors.Is
//
// err가 (몇 겹으로 감싸여 있든) ErrNotFound에서 비롯된 것이면 true.
//
// err == ErrNotFound 로 비교하면 안 됩니다. 감싸는 순간 다른 값이 됩니다.
// errors.Is는 감싼 사슬을 끝까지 따라가며 비교합니다.
func IsNotFound(err error) bool {
	// TODO
        if errors.Is(err, ErrNotFound) { return true }
	return false
}

// 문제 3 (계속) — errors.As
//
// err 사슬 어딘가에 *ParseError가 있으면 그 Reason을, 없으면 ""를 반환하세요.
//
// errors.Is는 "이 에러 값인가?"를 묻고,
// errors.As는 "이 타입이 들어있나? 있으면 꺼내줘"를 묻습니다.
//
//	var pe *ParseError
//	if errors.As(err, &pe) {
//	    // pe에 꺼낸 값이 들어있다
//	}
func ReasonOf(err error) string {
	// TODO
        var pe *ParseError
        if errors.As(err, &pe) { return pe.Reason }
	return ""
}

// 문제 4 — defer의 실행 순서 ★
//
// out이 정확히 ["c", "b", "a"] 가 되도록 defer 3개를 작성하세요.
// 각 defer는 out에 "a", "b", "c"를 순서대로 append 하는 코드여야 합니다.
//
// defer는 스택(LIFO)에 쌓입니다. 마지막에 등록한 것이 먼저 실행됩니다.
// 그리고 return 문이 실행된 **뒤에** 돌기 때문에,
// 이름 붙은 반환값(named return value)을 수정할 수 있습니다.
//
//	func DeferOrder() (out []string) {   // out이 이름 붙은 반환값
//	    defer func() { out = append(out, "a") }()
//	    ...
//	    return nil    // 여기서 out = nil 이 되고, 그 다음 defer들이 돈다
//	}
func DeferOrder() (out []string) {
	// TODO
        defer func() { out = append(out, "a") }()
        defer func() { out = append(out, "b") }()
        defer func() { out = append(out, "c") }()
	return nil
}

// Resource는 반드시 닫아야 하는 자원입니다 (파일, 소켓, DB 커넥션이라고 생각하세요).
type Resource struct {
	Closed bool
}

func (r *Resource) Close() error {
	r.Closed = true
	return nil
}

// 문제 5 — defer로 반드시 정리하기
//
// fail이 true면 에러를 반환하고, false면 nil을 반환하세요.
// **어느 쪽이든 r.Close()가 반드시 호출되어야 합니다.**
//
// 매 return마다 r.Close()를 쓰는 대신 defer를 씁니다.
// 함수 첫머리에 한 번 써두면 어떤 경로로 나가든(panic으로 나가도) 실행됩니다.
//
//	f, err := os.Open(path)
//	if err != nil { return err }
//	defer f.Close()          ← 여는 것 바로 다음 줄. 이게 Go 관례다
//
// C의 goto cleanup 패턴을 언어 기능으로 만든 것이라고 보면 됩니다.
func Process(r *Resource, fail bool) error {
	// TODO
        defer r.Close()
        if fail { return errors.New("error") }
	return nil
}

// 문제 6 — panic과 recover
//
// a/b를 반환하되, b가 0이라 panic이 나면 그걸 잡아서 에러로 바꾸세요.
// (이번엔 b를 미리 검사하지 말고, 진짜로 panic을 잡아야 합니다)
//
//	SafeDiv(10, 2) → (5, nil)
//	SafeDiv(10, 0) → (0, "recovered: runtime error: integer divide by zero" 같은 에러)
//
// recover()는 **defer 안에서만** 의미가 있습니다.
// panic 중이면 panic에 넘겨진 값을 반환하고 panic을 멈춥니다. 아니면 nil입니다.
//
//	defer func() {
//	    if r := recover(); r != nil {
//	        err = fmt.Errorf("recovered: %v", r)
//	    }
//	}()
//
// 주의: 이건 예외 처리 흉내가 아닙니다. Go에서 panic은 "여기까지 왔으면
// 프로그램이 잘못된 것"에만 씁니다. recover는 서버가 요청 하나 때문에
// 통째로 죽는 걸 막을 때 정도만 씁니다.
func SafeDiv(a, b int) (result int, err error) {
	// TODO — defer + recover를 추가하고, 마지막 줄을 return a / b, nil 로 바꾸세요.
	// (지금 상태로 두면 b가 0일 때 panic이 테스트 프로세스까지 죽입니다)
        defer func() {
            if r := recover(); r != nil {
                err = fmt.Errorf("recovered: %v", r)
                result = 0
            }
        }()
	return a/b, nil
}

// MyErr는 문제 7용 커스텀 에러입니다.
type MyErr struct{}

func (e *MyErr) Error() string { return "my error" }

// 문제 7 — typed nil 함정 ★★
//
// 아래 코드는 버그가 있습니다. fail이 false여도 반환된 error가 nil이 아닙니다.
// 고치세요.
//
// 왜 그런가:
// 인터페이스 값은 (타입, 값) 쌍입니다. nil인 *MyErr를 error에 넣으면
//
//	(타입=*MyErr, 값=nil)
//
// 이 되는데, 이건 (타입=nil, 값=nil)인 진짜 nil 인터페이스와 다릅니다.
// 그래서 err != nil 이 true가 됩니다.
//
// 이건 Go에서 가장 유명한 함정이고, 실무에서도 종종 물립니다.
// 규칙: **에러 반환값의 타입은 항상 error로 쓰고, 성공 경로에서는 nil 리터럴을 반환한다.**
func MaybeErr(fail bool) error {
	if fail {
		return &MyErr{}
	}
	return nil // TODO: 여기를 고치세요
}
