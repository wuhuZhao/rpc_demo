package internal

import (
	"testing"
)

func Sum(a, b int) int {
	return a + b
}
func TestRegister(t *testing.T) {
	err := Register(Sum)
	if err != nil {
		t.Fatalf("err: %v\n", err)
	}
	t.Logf("test success\n")
}

func TestCall(t *testing.T) {
	TestRegister(t)
	result, err := Call("Sum", 1, 2)
	if err != nil {
		t.Fatalf("err: %v\n", err)
	}
	if len(result) != 1 {
		t.Fatalf("len(result) is not equal to 1\n")
	}
	t.Logf("Sum(1,2) = %d\n", result[0].(int))
	if err := recover(); err != nil {
		t.Fatalf("%v\n", err)
	}
}
