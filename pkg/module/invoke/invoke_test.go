package invoke

import (
	"errors"
	"fmt"
	"testing"
)

func Func1() {
	fmt.Println("func1 ========")
}

func Func2() int {
	fmt.Println("func2 ========")
	return 1
}

func Func3() (int, int) {
	fmt.Println("func3 ========")
	return 1, 2
}

func Func4() error {
	fmt.Println("func4 ========")
	return errors.New("func4 error")
}

func Func5() (int, error) {
	fmt.Println("func5 ========")
	return 1, errors.New("func5 error")
}

func TestInvoke(t *testing.T) {
	iv := NewInvoker()
	iv.Register("1", Func1)
	iv.Register("2", Func2)
	iv.Register("3", Func3)
	iv.Register("4", Func4)
	iv.Register("5", Func5)

	result, err := iv.Invoke("1")
	fmt.Println(result, err)

	result, err = iv.Invoke("2")
	fmt.Println(result, err)

	result, err = iv.Invoke("3")
	fmt.Println(result, err)

	result, err = iv.Invoke("4")
	fmt.Println(result, err)
	
	result, err = iv.Invoke("5")
	fmt.Println(result, err)
}

