package mock

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

// OnCall declares that 'obj' method 'f' can be called with 'args' during the test.
// If 'args' are not specified 'f' can be called with any input parameters.
// The mocked function will return default constructed values in case of output parameters
// are not specified via Returner.Return method
func OnCall(obj interface{}, f interface{}, args ...interface{}) Returner {
	return getCore(obj).onCall(obj, f, args...)
}

// ExpectCall declares that 'obj' method 'f' must be called with 'args' during the test.
// If 'args' are not specified 'f' can be called with any input parameters.
// The mocked function will return default constructed values in case of output parameters
// are not specified via Returner.Return method
// Note that ExpectCall also sets the expected order of declared calls
func ExpectCall(obj interface{}, f interface{}, args ...interface{}) Returner {
	return getCore(obj).expectCall(obj, f, args...)
}

// Call perfoms a call of 'obj' method 'f' with input parameters 'args'
// This function should be used for mocking methods implementation.
// Pass named output parameters by reference. E.g ...Return(&out1, &out2)
func Call(obj interface{}, f interface{}, args ...interface{}) Returner {
	return getCore(obj).call(obj, f, args...)
}

// Returner specifies ouput parameters in case of OnCall/ExpectCall usage
// or assins ouput parameters in case of Call usage
type Returner interface {
	Return(out ...interface{})
}

// Core is the mocking engine. Declare it as first unnamed member of your mock structure.
// CheckExpectations should be called at the end of the test case. It checks that earlier
// declared via ExpectCall methods are realy called during the test
type Core interface {
	CheckExpectations()
}

// New creates mock.Core for the given *testing.T
func New(t TestingT) Core {
	return &core{t: t}
}

// TestingT is a ligth interface of testing.T. Is required for mock package been testable
type TestingT interface {
	Fatalf(format string, args ...interface{})
}

func getCore(obj interface{}) *core {
	objVal := reflect.ValueOf(obj)
	if objVal.Kind() != reflect.Ptr && objVal.Kind() != reflect.Interface {
		panic("obj must be kind of ptr or interface")
	}
	return objVal.Elem().Field(0).Interface().(*core)
}

type core struct {
	t        TestingT
	calls    []*callDeclaration
	expCalls []*expectedCallDeclaration
}

type callDeclaration struct {
	obj  interface{}
	fID  funcIdentity
	args []interface{}
	out  []interface{}
}

func (cd callDeclaration) String() string {
	return callToStr(cd.obj, cd.fID.name, cd.args)
}

type expectedCallDeclaration struct {
	callDeclaration
	used bool // TODO: make thread safe
}

func (cd *callDeclaration) satisfied(obj interface{}, fID funcIdentity, args []interface{}) bool {
	return cd.obj == obj &&
		reflect.DeepEqual(cd.fID, fID) &&
		(cd.args == nil || reflect.DeepEqual(cd.args, args))
}

type call struct {
	decl *callDeclaration
}

func adaptArgs(args []interface{}, f interface{}) {
	fType := reflect.TypeOf(f)

	for i, arg := range args {
		argType := reflect.TypeOf(arg)
		fArgType := fType.In(i + 1)
		if argType == fArgType {
			continue
		}

		if arg == nil {
			args[i] = reflect.Zero(fArgType).Interface()
			continue
		}

		if argType.ConvertibleTo(fArgType) {
			args[i] = reflect.ValueOf(arg).Convert(fArgType).Interface()
		}
	}
}

func (c *core) onCall(obj interface{}, f interface{}, args ...interface{}) Returner {
	validateCall(obj, f, args, true)
	if args != nil {
		adaptArgs(args, f)
	}
	cd := &callDeclaration{
		obj:  obj,
		fID:  getFuncID(f),
		args: args,
	}
	c.calls = append(c.calls, cd)
	return cd
}

func (c *core) expectCall(obj interface{}, f interface{}, args ...interface{}) Returner {
	validateCall(obj, f, args, true)
	if args != nil {
		adaptArgs(args, f)
	}
	ecd := &expectedCallDeclaration{
		callDeclaration: callDeclaration{
			obj:  obj,
			fID:  getFuncID(f),
			args: args,
		},
		used: false,
	}
	c.expCalls = append(c.expCalls, ecd)
	return ecd
}

func (c *core) call(obj interface{}, f interface{}, args ...interface{}) Returner {
	validateCall(obj, f, args, false)

	fID := getFuncID(f)
	for i, exp := range c.expCalls {
		if exp.used || !exp.satisfied(obj, fID, args) {
			continue
		}

		if i != 0 && !c.expCalls[i-1].used {
			for _, exp := range c.expCalls {
				if !exp.used {
					c.t.Fatalf(`%s must be called before %s`, exp, callToStr(obj, fName(f), args))
					return &call{nil}
				}
			}
		}

		exp.used = true
		return &call{&exp.callDeclaration}
	}

	for _, cd := range c.calls {
		if cd.satisfied(obj, fID, args) {
			return &call{cd}
		}
	}

	c.t.Fatalf(`%s called but not defined`, callToStr(obj, fName(f), args))
	return &call{nil}
}

func (c *core) CheckExpectations() {
	for _, exp := range c.expCalls {
		if !exp.used {
			c.t.Fatalf(`%s expected but not called`, exp)
			return
		}
	}
}

func (cd *callDeclaration) Return(out ...interface{}) {
	if len(out) != cd.fID.fType.NumOut() {
		panic(fmt.Sprintf(`Invalid %s return values: count must be %d`,
			cd.fID.name, cd.fID.fType.NumOut()))
	}

	for i, gotOut := range out {
		if err := validateFuncParam(cd.fID.fType.Out(i), gotOut); err != nil {
			panic(fmt.Sprintf(`Invalid %s %d-th return value: %s`, cd.fID.name, i+1, err.Error()))
		}
	}

	cd.out = append(cd.out, out...)
}

func validateFuncParam(paramType reflect.Type, paramValue interface{}) error {
	if paramValue == nil {
		if paramType.Kind() == reflect.Ptr ||
			paramType.Kind() == reflect.Interface ||
			paramType.Kind() == reflect.Slice {
			return nil
		}

		return fmt.Errorf("nil is invalid value for type %s", paramType)
	}

	gotParamType := reflect.TypeOf(paramValue)
	if !gotParamType.AssignableTo(paramType) && !gotParamType.ConvertibleTo(paramType) {
		return fmt.Errorf("%v is neither assignable nor converible to type %s", paramValue, paramType)
	}

	return nil
}

func (c *call) Return(out ...interface{}) {
	if c.decl == nil {
		return // for internal tests
	}

	if len(out) != c.decl.fID.fType.NumOut() {
		panic(fmt.Sprintf(`Invalid %s call return parameters count: got %d, expected %d`,
			c.decl, len(out), c.decl.fID.fType.NumOut()))
	}

	if c.decl.out == nil {
		return
	}

	for i, r := range out {
		if c.decl.out[i] == nil {
			continue
		}

		if reflect.TypeOf(r).Kind() != reflect.Ptr {
			panic(fmt.Sprintf(`Invalid %s call %d-th return parameter binding. Ptr expected.`, c.decl, i+1))
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
		panic("f must be an obj interface method")
	}

	if optionalArgs && args == nil {
		return
	}

	if fType.NumIn()-1 != len(args) {
		panic(fmt.Sprintf(`Invalid %s args count. Got %d, expected %d`,
			fName(f), len(args), fType.NumIn()-1))
	}

	for i, arg := range args {
		paramType := fType.In(i + 1)
		if err := validateFuncParam(paramType, arg); err != nil {
			panic(fmt.Sprintf(`Invalid %s %d-th arg: %s`, fName(f), i+1, err.Error()))
		}
	}
}

type funcIdentity struct {
	name  string
	fType reflect.Type
}

func getFuncID(f interface{}) funcIdentity {
	return funcIdentity{
		name:  fName(f),
		fType: reflect.TypeOf(f),
	}
}

func fName(f interface{}) string {
	fullName := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	tokens := strings.Split(fullName, ".")
	return tokens[len(tokens)-1]
}

func callToStr(obj interface{}, name string, args []interface{}) string {
	if args == nil {
		return name
	}

	strArgs := []string{}
	for _, arg := range args {
		strArgs = append(strArgs, fmt.Sprint(arg))
	}

	return fmt.Sprintf("%v.%s(%s)", obj, name, strings.Join(strArgs, ", "))
}
