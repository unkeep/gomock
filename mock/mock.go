package mock

import (
	"fmt"
	"reflect"
	"runtime"
)

func OnCall(obj interface{}, f interface{}, args ...interface{}) Returner {
	return getCore(obj).onCall(obj, f, args...)
}

func ExpectCall(obj interface{}, f interface{}, args ...interface{}) Returner {
	return getCore(obj).expectCall(obj, f, args...)
}

func Call(obj interface{}, f interface{}, args ...interface{}) Returner {
	return getCore(obj).call(obj, f, args...)
}

type Returner interface {
	Return(out ...interface{})
}

type Core interface {
	CheckExpectations()
}

func New(t TestingT) Core {
	return &core{t: t}
}

type TestingT interface {
	Fatalf(format string, args ...interface{})
}

func getCore(obj interface{}) *core {
	objVal := reflect.ValueOf(obj)
	if objVal.Kind() != reflect.Ptr && objVal.Kind() != reflect.Interface {
		panic("obj must be kind of valid ptr or interface")
	}
	return objVal.Elem().Field(0).Interface().(*core)
}

type core struct {
	t        TestingT
	calls    []*callDeclaration
	expCalls []*expectedCallDeclaration
	counter  int
}

type callDeclaration struct {
	obj  interface{}
	fID  funcIdentity
	args []interface{}
	out  []interface{}
}

type expectedCallDeclaration struct {
	callDeclaration
	used bool
}

func (cd *callDeclaration) satisfied(obj interface{}, fID funcIdentity, args []interface{}) bool {
	return cd.obj == obj &&
		reflect.DeepEqual(cd.fID, fID) &&
		(cd.args == nil || reflect.DeepEqual(cd.args, args))
}

type call struct {
	decl *callDeclaration
}

func (m *core) onCall(obj interface{}, f interface{}, args ...interface{}) Returner {
	validateCall(obj, f, args, true)
	c := &callDeclaration{obj: obj, fID: getFuncID(f), args: args}
	m.calls = append(m.calls, c)
	return c
}

func (m *core) expectCall(obj interface{}, f interface{}, args ...interface{}) Returner {
	validateCall(obj, f, args, true)
	c := &expectedCallDeclaration{
		callDeclaration{obj: obj, fID: getFuncID(f), args: args},
		false,
	}
	m.expCalls = append(m.expCalls, c)
	return c
}

func (m *core) call(obj interface{}, f interface{}, args ...interface{}) Returner {
	validateCall(obj, f, args, false)

	fID := getFuncID(f)
	for i, exp := range m.expCalls {
		if exp.used || !exp.satisfied(obj, fID, args) {
			continue
		}

		if i != 0 && !m.expCalls[i-1].used {
			for _, exp := range m.expCalls {
				if !exp.used {
					m.t.Fatalf("'%s' must be called before '%s'", exp.fID, fID)
					return &call{nil}
				}
			}
		}

		exp.used = true
		return &call{&exp.callDeclaration}
	}

	for _, c := range m.calls {
		if c.satisfied(obj, fID, args) {
			return &call{c}
		}
	}

	m.t.Fatalf("'%s' is called but not defined", fID)
	return &call{nil}
}

func (m *core) CheckExpectations() {
	for _, exp := range m.expCalls {
		if !exp.used {
			m.t.Fatalf("'%s' is expected but not called", exp.fID)
			return
		}
	}
}

func (c *callDeclaration) Return(out ...interface{}) {
	if len(out) != c.fID.fType.NumOut() {
		panic(fmt.Sprintf("Invalid '%s' call declaration: out parameters count must be %d", c.fID, c.fID.fType.NumOut()))
	}

	for i, gotOut := range out {
		if err := validateFuncParam(c.fID.fType.Out(i), gotOut); err != nil {
			panic(fmt.Sprintf("Invalid '%s' call out parameters declaration: %s", c.fID, err.Error()))
		}
	}

	c.out = append(c.out, out...)
}

func validateFuncParam(paramType reflect.Type, paramValue interface{}) error {
	if paramValue == nil {
		if paramType.Kind() == reflect.Ptr || paramType.Kind() == reflect.Interface {
			return nil
		}

		return fmt.Errorf("nil is invalid value for type %s", paramType)
	}

	gotParamType := reflect.TypeOf(paramValue)
	if !gotParamType.AssignableTo(paramType) {
		return fmt.Errorf("Type %s is not assignable to type %s", gotParamType, paramType)
	}

	return nil
}

func (c *call) Return(out ...interface{}) {
	if c.decl == nil {
		return // for the tests
	}

	if len(out) != c.decl.fID.fType.NumOut() {
		panic(fmt.Sprintf("Invalid '%s' call: out parameters ptrs count must be %d", c.decl.fID, c.decl.fID.fType.NumOut()))
	}

	if c.decl.out == nil {
		return
	}

	for i, r := range out {
		if c.decl.out[i] == nil {
			continue
		}

		retVal := reflect.ValueOf(r).Elem()
		retVal.Set(reflect.ValueOf(c.decl.out[i]))
	}
}

func validateCall(obj interface{}, f interface{}, args []interface{}, optionalArgs bool) {
	objVal := reflect.ValueOf(obj)

	fType := reflect.TypeOf(f)
	if fType.Kind() != reflect.Func {
		panic("f must be kind of function")
	}

	if fType.NumIn() < 1 || !objVal.Type().AssignableTo(fType.In(0)) {
		panic("f must an interface method of obj")
	}

	if optionalArgs && args == nil {
		return
	}

	if fType.NumIn()-1 != len(args) {
		panic(fmt.Sprintf("Invalid '%s' call declaration: in parameters count must be %d", fName(f), fType.NumIn()-1))
	}

	for i, arg := range args {
		if err := validateFuncParam(fType.In(i+1), arg); err != nil {
			panic(fmt.Sprintf("Invalid '%s' call in parameters declaration: %s", fName(f), err.Error()))
		}
	}
}

type funcIdentity struct {
	name  string
	fType reflect.Type
}

func (fID funcIdentity) String() string {
	return fID.name
}

func getFuncID(f interface{}) funcIdentity {
	return funcIdentity{
		name:  fName(f),
		fType: reflect.TypeOf(f),
	}
}

func fName(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}
