package errs

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestDivide(t *testing.T) {
	if got, err := Divide(10, 2); got != 5 || err != nil {
		t.Errorf("Divide(10, 2) = %d, %v; want 5, nil", got, err)
	}
	got, err := Divide(10, 0)
	if err == nil {
		t.Fatal("Divide(10, 0) err = nil; want 에러")
	}
	if got != 0 {
		t.Errorf("Divide(10, 0) = %d; want 0 (에러일 때는 제로값을 반환합니다)", got)
	}
	if msg := err.Error(); msg == "" {
		t.Error("에러 메시지가 비어 있습니다")
	}
}

func TestParseErrorMessage(t *testing.T) {
	e := &ParseError{Input: "abc", Reason: "not a number"}
	want := `parse "abc": not a number`
	if got := e.Error(); got != want {
		t.Errorf("Error() = %q; want %q", got, want)
	}
	// error 인터페이스를 만족해야 한다.
	var _ error = e
}

func TestParseAge(t *testing.T) {
	if got, err := ParseAge("42"); got != 42 || err != nil {
		t.Errorf(`ParseAge("42") = %d, %v; want 42, nil`, got, err)
	}

	tests := []struct {
		in         string
		wantReason string
	}{
		{"abc", "not a number"},
		{"", "not a number"},
		{"3.5", "not a number"},
		{"-1", "negative"},
	}
	for _, tt := range tests {
		_, err := ParseAge(tt.in)
		if err == nil {
			t.Errorf("ParseAge(%q) err = nil; want 에러", tt.in)
			continue
		}
		var pe *ParseError
		if !errors.As(err, &pe) {
			t.Errorf("ParseAge(%q) 에러 타입 = %T; want *ParseError", tt.in, err)
			continue
		}
		if pe.Reason != tt.wantReason {
			t.Errorf("ParseAge(%q) Reason = %q; want %q", tt.in, pe.Reason, tt.wantReason)
		}
		if pe.Input != tt.in {
			t.Errorf("ParseAge(%q) Input = %q; want %q", tt.in, pe.Input, tt.in)
		}
	}
}

func TestFindUser(t *testing.T) {
	db := map[string]string{"kim": "김철수"}

	if got, err := FindUser(db, "kim"); got != "김철수" || err != nil {
		t.Errorf(`FindUser(db, "kim") = %q, %v; want "김철수", nil`, got, err)
	}

	_, err := FindUser(db, "lee")
	if err == nil {
		t.Fatal(`FindUser(db, "lee") err = nil; want 에러`)
	}
	// 감싼 에러는 원본과 "같지 않다". 그래서 == 비교는 실패해야 정상이다.
	if err == ErrNotFound {
		t.Error("에러를 감싸지 않았습니다. fmt.Errorf(..., %w)를 쓰세요")
	}
	// 하지만 errors.Is로는 찾아져야 한다.
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("errors.Is(err, ErrNotFound) = false; want true. err = %v (%%w 대신 %%v를 쓰지 않았는지 확인하세요)", err)
	}
	// 문맥이 메시지에 남아야 한다.
	if !strings.Contains(err.Error(), "lee") {
		t.Errorf("에러 메시지 = %q; 찾으려던 이름이 들어가야 합니다", err.Error())
	}
}

func TestIsNotFound(t *testing.T) {
	db := map[string]string{}
	_, err := FindUser(db, "nobody")

	if !IsNotFound(err) {
		t.Error("IsNotFound(감싼 에러) = false; want true")
	}
	// 여러 겹 감싸도 찾아야 한다.
	deep := fmt.Errorf("layer2: %w", fmt.Errorf("layer1: %w", err))
	if !IsNotFound(deep) {
		t.Error("IsNotFound(여러 겹 감싼 에러) = false; want true")
	}
	if IsNotFound(errors.New("다른 에러")) {
		t.Error("IsNotFound(무관한 에러) = true; want false")
	}
	if IsNotFound(nil) {
		t.Error("IsNotFound(nil) = true; want false")
	}
}

func TestReasonOf(t *testing.T) {
	_, err := ParseAge("abc")
	if got := ReasonOf(err); got != "not a number" {
		t.Errorf("ReasonOf = %q; want %q", got, "not a number")
	}
	// 감싸도 찾아내야 한다.
	wrapped := fmt.Errorf("outer: %w", err)
	if got := ReasonOf(wrapped); got != "not a number" {
		t.Errorf("ReasonOf(감싼 에러) = %q; want %q", got, "not a number")
	}
	if got := ReasonOf(errors.New("무관")); got != "" {
		t.Errorf("ReasonOf(무관한 에러) = %q; want \"\"", got)
	}
	if got := ReasonOf(nil); got != "" {
		t.Errorf("ReasonOf(nil) = %q; want \"\"", got)
	}
}

func TestDeferOrder(t *testing.T) {
	got := DeferOrder()
	want := []string{"c", "b", "a"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("DeferOrder() = %v; want %v (defer는 LIFO로 실행되고, "+
			"이름 붙은 반환값을 수정할 수 있습니다)", got, want)
	}
}

func TestProcess(t *testing.T) {
	r := &Resource{}
	if err := Process(r, false); err != nil {
		t.Errorf("Process(r, false) = %v; want nil", err)
	}
	if !r.Closed {
		t.Error("성공 경로에서 Close가 호출되지 않았습니다")
	}

	r2 := &Resource{}
	if err := Process(r2, true); err == nil {
		t.Error("Process(r, true) = nil; want 에러")
	}
	if !r2.Closed {
		t.Error("에러 경로에서 Close가 호출되지 않았습니다. defer를 쓰세요")
	}
}

func TestSafeDiv(t *testing.T) {
	if got, err := SafeDiv(10, 2); got != 5 || err != nil {
		t.Errorf("SafeDiv(10, 2) = %d, %v; want 5, nil", got, err)
	}

	// panic이 함수 밖으로 새어나가면 이 테스트 자체가 죽는다.
	got, err := SafeDiv(10, 0)
	if err == nil {
		t.Fatal("SafeDiv(10, 0) err = nil; want 에러")
	}
	if got != 0 {
		t.Errorf("SafeDiv(10, 0) = %d; want 0", got)
	}
	if !strings.Contains(err.Error(), "divide by zero") {
		t.Errorf("에러 메시지 = %q; panic 내용이 담겨야 합니다", err.Error())
	}
}

func TestMaybeErr(t *testing.T) {
	if err := MaybeErr(false); err != nil {
		t.Errorf("MaybeErr(false) = %v (타입 %T); want nil\n"+
			"  → nil인 *MyErr를 error에 넣으면 (타입=*MyErr, 값=nil)이 되어\n"+
			"    진짜 nil 인터페이스와 달라집니다.", err, err)
	}
	if err := MaybeErr(true); err == nil {
		t.Error("MaybeErr(true) = nil; want 에러")
	} else if err.Error() != "my error" {
		t.Errorf("MaybeErr(true).Error() = %q; want %q", err.Error(), "my error")
	}
}
