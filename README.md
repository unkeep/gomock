# gomock
[![GoDoc](https://godoc.org/github.com/unkeep/gomock?status.svg)](https://godoc.org/github.com/unkeep/gomock)
[![Build Status](https://travis-ci.org/unkeep/gomock.svg?branch=master)](https://travis-ci.org/unkeep/gomock)
[![codecov.io Code Coverage](https://img.shields.io/codecov/c/github/unkeep/gomock.svg?maxAge=2592000)](https://codecov.io/github/unkeep/gomock?branch=master)

*gomock* is golang mocking framework which provides a fast way of implementing light and flexible mock objects. It will help you to make your tests more useful and readable!

* _mock_ - core package. Contains types and functions for implementing/setup mock objects
* _gomock_ - commandline tool, generates mock objects from interface

Download:
```shell
go get github.com/unkeep/gomock
```

If you do not have the go command on your system, you need to [Install Go](http://golang.org/doc/install) first

* * *
Usage:

Suppose you have an interface:

```golang
type Storage interface {
	GetValue(key string) (int, error)
	SetValue(key string, value int) error
}
```

And function you would like to test that depends on this interface:

```golang
func IncrementValue(key string, st Storage) (int, error) {
	val, err := st.GetValue(key)
	if err != nil {
		return 0, err
	}
	val++
	if err := st.SetValue(key, val); err != nil {
		return 0, err
	}
	return val, nil
}
```

You can declare a mock implementation by yourself or generate it by gomock tool:

```golang
 type mockStorage struct {
 	mock.M
 }

 func (m *mockStorage) GetValue(key string) (val int, err error) {
	 mock.Call(m, Storage.GetValue, key).Return(&val, &err)
	 return
 }

 func (m *mockStorage) SetValue(key string, value int) (err error) {
	 mock.Call(m, Storage.SetValue, key, value).Return(&err)
	 return
 }
```

The test of IncrementValue can look like that:

```golang
func TestIncrementValue(t *testing.T) {
	cases := []struct {
		name      string
		setupMock func(Storage)
		in        string
		out       int
		outErr    error
	}{
		{
			name: "success_scenario",
			setupMock: func(st Storage) {
				mock.ExpectCall(st, Storage.GetValue, "key").Return(123, nil)
				mock.ExpectCall(st, Storage.SetValue, "key", 124)
			},
			in:  "key",
			out: 124,
		},
		{
			name: "on_GetValue_error",
			setupMock: func(st Storage) {
				mock.ExpectCall(st, Storage.GetValue).Return(0, fmt.Errorf("Error"))
			},
			outErr: fmt.Errorf("Error"),
		},
		{
			name: "on_SetValue_error",
			setupMock: func(st Storage) {
				mock.ExpectCall(st, Storage.GetValue)
				mock.ExpectCall(st, Storage.SetValue).Return(fmt.Errorf("Error"))
			},
			outErr: fmt.Errorf("Error"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := mock.New(t)
			st := &mockStorage{m}
			c.setupMock(st)
			newVal, err := IncrementValue(c.in, st)

			if !reflect.DeepEqual(err, c.outErr) {
				t.Fatalf(`Error expected: "%v", got "%v"`, c.outErr, err)
			}

			if newVal != c.out {
				t.Fatalf(`Value expected: "%d", got: "%d"`, c.out, newVal)
			}

			m.CheckExpectations()
		})
	}
}
```

