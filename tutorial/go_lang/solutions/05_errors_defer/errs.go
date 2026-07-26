// Package errs — 05_errors_defer 정답.
package errs

import (
	"errors"
	"fmt"
	"strconv"
)

// 문제 1 — 기본 에러 반환
func Divide(a, b int) (int, error) {
	if b == 0 {
		// 에러일 때 첫 번째 반환값은 제로값으로 둔다.
		// 호출한 쪽이 err만 보고 판단할 수 있게 하는 관례다.
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

type ParseError struct {
	Input  string
	Reason string
}

// 문제 2 — 커스텀 에러 타입
//
// %q는 문자열을 따옴표로 감싸고 특수문자를 이스케이프한다.
// 사용자 입력을 에러 메시지에 넣을 때 %s보다 %q가 안전하고 읽기 좋다.
func (e *ParseError) Error() string {
	return fmt.Sprintf("parse %q: %s", e.Input, e.Reason)
}

func ParseAge(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		// strconv가 준 err는 여기서 버린다. 우리 타입으로 바꿔서 반환한다.
		// (원본을 남기고 싶으면 ParseError에 Err error 필드를 두고
		//  Unwrap() error 메서드를 만들면 errors.Is/As가 사슬을 따라간다)
		return 0, &ParseError{Input: s, Reason: "not a number"}
	}
	if n < 0 {
		return 0, &ParseError{Input: s, Reason: "negative"}
	}
	return n, nil
}

var ErrNotFound = errors.New("not found")

// 문제 3 — 에러 감싸기(wrapping) ★
func FindUser(db map[string]string, name string) (string, error) {
	v, ok := db[name]
	if !ok {
		// %w가 핵심이다. %v로 쓰면 메시지 문자열만 남고
		// errors.Is(err, ErrNotFound)가 false가 된다.
		return "", fmt.Errorf("find user %q: %w", name, ErrNotFound)
	}
	return v, nil
}

func IsNotFound(err error) bool {
	// errors.Is는 Unwrap() 사슬을 끝까지 따라가며 == 로 비교한다.
	// err가 nil이면 그냥 false를 반환하므로 nil 검사를 따로 안 해도 된다.
	return errors.Is(err, ErrNotFound)
}

func ReasonOf(err error) string {
	// errors.As의 두 번째 인자는 "대상 타입의 포인터"여야 한다.
	// 우리가 찾는 게 *ParseError이므로 **ParseError를 넘기게 된다.
	var pe *ParseError
	if errors.As(err, &pe) {
		return pe.Reason
	}
	return ""
}

// 문제 4 — defer의 실행 순서 ★
func DeferOrder() (out []string) {
	// defer는 스택에 쌓인다. 등록 순서 a, b, c → 실행 순서 c, b, a.
	defer func() { out = append(out, "a") }()
	defer func() { out = append(out, "b") }()
	defer func() { out = append(out, "c") }()

	// return이 먼저 out = nil 을 수행하고, 그 다음 defer들이 돈다.
	// out이 "이름 붙은 반환값"이기 때문에 defer가 손을 댈 수 있다.
	//
	// 주의: defer fmt.Println(out) 처럼 클로저 없이 쓰면
	// 인자가 defer 등록 시점에 평가되어 그때의 out(nil)이 박힌다.
	// 나중 값을 보려면 반드시 클로저로 감싸야 한다.
	return nil
}

type Resource struct {
	Closed bool
}

func (r *Resource) Close() error {
	r.Closed = true
	return nil
}

// 문제 5 — defer로 반드시 정리하기
func Process(r *Resource, fail bool) error {
	// 자원을 얻은 직후에 defer를 건다. 이게 Go의 표준 리듬이다.
	// 아래에 return이 몇 개가 생기든, panic이 나든 반드시 실행된다.
	defer r.Close()

	if fail {
		return errors.New("processing failed")
	}
	return nil

	// ⚠️ 루프 안에서는 이렇게 쓰면 안 된다.
	// defer는 함수가 끝날 때 실행되지 블록이 끝날 때가 아니다.
	// 루프 몸통을 별도 함수로 빼거나, 루프 안에서 명시적으로 Close를 부른다.
}

// 문제 6 — panic과 recover
func SafeDiv(a, b int) (result int, err error) {
	// 반환값에 이름(result, err)이 있어야 defer가 수정할 수 있다.
	// func SafeDiv(a, b int) (int, error) 였다면 recover로 잡아도
	// 반환값을 바꿀 방법이 없다.
	defer func() {
		if r := recover(); r != nil {
			// r은 any다. 0으로 나누기면 runtime.Error가 들어온다.
			err = fmt.Errorf("recovered: %v", r)
			// result는 제로값 0인 채로 남는다.
		}
	}()

	// b가 0이면 여기서 panic: "runtime error: integer divide by zero"
	return a / b, nil
}

type MyErr struct{}

func (e *MyErr) Error() string { return "my error" }

// 문제 7 — typed nil 함정 ★★
func MaybeErr(fail bool) error {
	if fail {
		return &MyErr{}
	}
	// nil 리터럴을 직접 반환한다.
	// 이렇게 하면 인터페이스가 (타입=nil, 값=nil)인 진짜 nil이 된다.
	return nil

	// 원래 버그 코드:
	//
	//	var e *MyErr        // 구체 타입 포인터
	//	if fail { e = &MyErr{} }
	//	return e            // e가 nil이어도 error는 (타입=*MyErr, 값=nil)
	//	                    // → err != nil 이 true가 된다
	//
	// 규칙: 에러를 담는 지역변수의 타입은 항상 error로 쓴다.
	// 구체 타입 포인터를 인터페이스 반환값으로 승격시키지 않는다.
}
