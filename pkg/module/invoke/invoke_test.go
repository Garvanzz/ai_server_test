package invoke

import (
	"errors"
	"testing"
)

func funcNoReturn() {}

func funcReturnInt() int { return 42 }

func funcReturnError() error { return errors.New("err") }

func funcReturnIntError() (int, error) { return 1, errors.New("func error") }

func funcReturnIntNilError() (int, error) { return 99, nil }

func funcWithArgs(a int, b string) string { return b }

func funcWithPtrArg(p *int) int {
	if p == nil {
		return -1
	}
	return *p
}

func TestInvoke(t *testing.T) {
	iv := NewInvoker()
	iv.Register("noReturn", funcNoReturn)
	iv.Register("retInt", funcReturnInt)
	iv.Register("retErr", funcReturnError)
	iv.Register("retIntErr", funcReturnIntError)
	iv.Register("retIntNil", funcReturnIntNilError)
	iv.Register("withArgs", funcWithArgs)
	iv.Register("withPtr", funcWithPtrArg)

	if r, err := iv.Invoke("noReturn"); r != nil || err != nil {
		t.Fatalf("noReturn: got (%v, %v)", r, err)
	}

	if r, err := iv.Invoke("retInt"); r != 42 || err != nil {
		t.Fatalf("retInt: got (%v, %v)", r, err)
	}

	if r, err := iv.Invoke("retErr"); r != nil || err == nil {
		t.Fatalf("retErr: got (%v, %v)", r, err)
	}

	if r, err := iv.Invoke("retIntErr"); r != 1 || err == nil {
		t.Fatalf("retIntErr: got (%v, %v)", r, err)
	}

	if r, err := iv.Invoke("retIntNil"); r != 99 || err != nil {
		t.Fatalf("retIntNil: got (%v, %v)", r, err)
	}

	if r, err := iv.Invoke("withArgs", 10, "hello"); r != "hello" || err != nil {
		t.Fatalf("withArgs: got (%v, %v)", r, err)
	}

	if r, err := iv.Invoke("withPtr", nil); r != -1 || err != nil {
		t.Fatalf("withPtr(nil): got (%v, %v)", r, err)
	}

	v := 7
	if r, err := iv.Invoke("withPtr", &v); r != 7 || err != nil {
		t.Fatalf("withPtr(&7): got (%v, %v)", r, err)
	}
}

func TestInvokeArgCountMismatch(t *testing.T) {
	iv := NewInvoker()
	iv.Register("withArgs", funcWithArgs)
	if _, err := iv.Invoke("withArgs", 1); err == nil {
		t.Fatal("expected error for arg count mismatch")
	}
}

func TestInvokeArgTypeMismatch(t *testing.T) {
	iv := NewInvoker()
	iv.Register("withArgs", funcWithArgs)
	if _, err := iv.Invoke("withArgs", "wrong", "hello"); err == nil {
		t.Fatal("expected error for arg type mismatch")
	}
}

func TestInvokeNotFound(t *testing.T) {
	iv := NewInvoker()
	if _, err := iv.Invoke("missing"); err == nil {
		t.Fatal("expected error for missing function")
	}
}

func TestRegisterNotAFunction(t *testing.T) {
	iv := NewInvoker()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for non-function")
		}
	}()
	iv.Register("bad", 123)
}

func TestRegisterTooManyReturns(t *testing.T) {
	iv := NewInvoker()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for too many return values")
		}
	}()
	iv.Register("bad", func() (int, int, int) { return 0, 0, 0 })
}

func TestRegisterSecondReturnNotError(t *testing.T) {
	iv := NewInvoker()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for non-error second return")
		}
	}()
	iv.Register("bad", func() (int, int) { return 0, 0 })
}

func TestRegisterDuplicate(t *testing.T) {
	iv := NewInvoker()
	iv.Register("fn", funcNoReturn)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for duplicate registration")
		}
	}()
	iv.Register("fn", funcNoReturn)
}

func TestInvokeNilOnNonPointer(t *testing.T) {
	iv := NewInvoker()
	iv.Register("withArgs", funcWithArgs)
	if _, err := iv.Invoke("withArgs", nil, "hello"); err == nil {
		t.Fatal("expected error for nil on non-pointer param")
	}
}

