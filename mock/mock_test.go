package mock

import (
	"fmt"
	"reflect"
	"testing"
)

type myType struct {
	data string
}

type myInterface interface {
	doSmth(int) (*myType, error)
	doSmth2() myType
}

type myInterface2 interface {
	doSmth3(int) (*myType, error)
}

type myObj struct {
	Core
}

func (obj *myObj) doSmth(arg int) (v *myType, err error) {
	Call(obj, myInterface.doSmth, arg).Return(&v, &err)
	return
}

func (obj *myObj) doSmth2() (v myType) {
	Call(obj, myInterface.doSmth2).Return(&v)
	return
}

type tmock struct {
	fail bool
}

func (t *tmock) Fatalf(format string, args ...interface{}) {
	t.fail = true
}

func TestOnCallBase(t *testing.T) {
	tm := new(tmock)
	m := New(tm)
	obj := myInterface(&myObj{m})

	OnCall(obj, myInterface.doSmth)

	v, err := obj.doSmth(123)

	if tm.fail {
		t.Fatal("tm.fail")
	}

	if v != nil {
		t.Fatal("v != nil")
	}

	if err != nil {
		t.Fatal("err != nil")
	}
}

func TestOnCallWithArgsAndReturn(t *testing.T) {
	tm := new(tmock)
	m := New(tm)
	obj := &myObj{m}

	out1 := &myType{"data"}
	out2 := fmt.Errorf("error")
	OnCall(obj, myInterface.doSmth, 123).Return(out1, out2)

	v, err := obj.doSmth(123)

	if tm.fail {
		t.Fatal("tm.fail")
	}

	if !reflect.DeepEqual(v, out1) {
		t.Fatal("!reflect.DeepEqual(v, out1)")
	}

	if !reflect.DeepEqual(err, out2) {
		t.Fatal("!reflect.DeepEqual(err, out2)")
	}
}

func TestOnCallWithInvalidNonDefinedMethod(t *testing.T) {
	tm := new(tmock)
	m := New(tm)
	obj := &myObj{m}

	OnCall(obj, myInterface.doSmth, 123)

	obj.doSmth2()

	if !tm.fail {
		t.Fatal("!tm.fail")
	}
}

func TestOnCallWithInvalidArgs(t *testing.T) {
	tm := new(tmock)
	m := New(tm)
	obj := &myObj{m}

	OnCall(obj, myInterface.doSmth, 123)

	v, err := obj.doSmth(321)

	if !tm.fail {
		t.Fatal("!tm.fail")
	}

	if v != nil {
		t.Fatal("v != nil")
	}

	if err != nil {
		t.Fatal("err != nil")
	}
}

func TestOnCallWithNillReturn(t *testing.T) {
	tm := new(tmock)
	m := New(tm)
	obj := &myObj{m}

	OnCall(obj, myInterface.doSmth).Return(nil, nil)

	v, err := obj.doSmth(123)

	if tm.fail {
		t.Fatal("tm.fail")
	}

	if v != nil {
		t.Fatal("v != nil")
	}

	if err != nil {
		t.Fatal("err != nil")
	}
}

func TestOnCallWithNonPtrReturn(t *testing.T) {
	tm := new(tmock)
	m := New(tm)
	obj := &myObj{m}

	OnCall(obj, myInterface.doSmth2).Return(myType{"data"})

	v := obj.doSmth2()

	if tm.fail {
		t.Fatal("tm.fail")
	}

	if !reflect.DeepEqual(v, myType{"data"}) {
		t.Fatal(`!reflect.DeepEqual(v, myType{"data"})`)
	}
}

func TestExpectCall(t *testing.T) {
	tm := new(tmock)
	m := New(tm)
	obj := &myObj{m}

	ExpectCall(obj, myInterface.doSmth2).Return(myType{"data"})

	v := obj.doSmth2()

	m.CheckExpectations()

	if tm.fail {
		t.Fatal("tm.fail")
	}

	if !reflect.DeepEqual(v, myType{"data"}) {
		t.Fatal(`!reflect.DeepEqual(v, myType{"data"})`)
	}
}

func TestExpectedCallsOrder(t *testing.T) {
	tm := new(tmock)
	m := New(tm)
	obj := &myObj{m}

	ExpectCall(obj, myInterface.doSmth, 1)
	ExpectCall(obj, myInterface.doSmth, 2)

	obj.doSmth(2)

	if !tm.fail {
		t.Fatal("!tm.fail")
	}
}

func TestCheckExpectations(t *testing.T) {
	tm := new(tmock)
	m := New(tm)
	obj := &myObj{m}

	OnCall(obj, myInterface.doSmth)
	ExpectCall(obj, myInterface.doSmth, 123)

	obj.doSmth(321)

	m.CheckExpectations()

	if !tm.fail {
		t.Fatal("!tm.fail")
	}
}

func TestInvalidOutParamsCountDeclaration(t *testing.T) {
	defer expectPanic(t)

	tm := new(tmock)
	m := New(tm)
	obj := &myObj{m}

	OnCall(obj, myInterface.doSmth2).Return(myType{}, myType{})
}

func TestInvalidOutParamTypeDeclaration(t *testing.T) {
	defer expectPanic(t)

	tm := new(tmock)
	m := New(tm)
	obj := &myObj{m}

	OnCall(obj, myInterface.doSmth2).Return(&myType{})
}

func TestInvalidNilOutParamTypeDeclaration(t *testing.T) {
	defer expectPanic(t)

	tm := new(tmock)
	m := New(tm)
	obj := &myObj{m}

	OnCall(obj, myInterface.doSmth2).Return(nil)
}

func TestInvalidOutParamCountCall(t *testing.T) {
	defer expectPanic(t)

	tm := new(tmock)
	m := New(tm)
	obj := &myObj{m}
	OnCall(obj, myInterface.doSmth2).Return(myType{})
	Call(obj, myInterface.doSmth2).Return(myType{}, myType{})
}

func TestInvalidInParamsCountDeclaration(t *testing.T) {
	defer expectPanic(t)

	tm := new(tmock)
	m := New(tm)
	obj := &myObj{m}

	OnCall(obj, myInterface.doSmth, 1, 2)
}

func TestInvalidInParamTypeDeclaration(t *testing.T) {
	defer expectPanic(t)

	tm := new(tmock)
	m := New(tm)
	obj := &myObj{m}

	OnCall(obj, myInterface.doSmth, "int expected")
}

func TestInvalidObjDeclaration(t *testing.T) {
	defer expectPanic(t)

	OnCall(myInterface.doSmth, myInterface.doSmth)
}

func TestInvalidObjFuncDeclaration(t *testing.T) {
	defer expectPanic(t)

	tm := new(tmock)
	m := New(tm)
	obj := &myObj{m}

	OnCall(obj, obj)
}

func TestNonObjFuncDeclaration(t *testing.T) {
	defer expectPanic(t)

	tm := new(tmock)
	m := New(tm)
	obj := &myObj{m}

	OnCall(obj, myInterface2.doSmth3)
}

func expectPanic(t *testing.T) {
	if recover() == nil {
		t.Fatal("panic expected")
	}
}
