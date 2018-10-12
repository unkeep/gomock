package mockexample

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/unkeep/gomock/mock"
)

// storage the interface we will mock
type storage interface {
	GetValue(key string) (int, error)
	SetValue(key string, value int) error
}

// incrementValue the function we will test
func incrementValue(key string, st storage) (int, error) {
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

// mockStorage the Storage mock. You can implement it yourself or using gomock help tool
type mockStorage struct {
	mock.Core
}

func (m *mockStorage) GetValue(key string) (val int, err error) {
	mock.Call(m, storage.GetValue, key).Return(&val, &err)
	return
}

func (m *mockStorage) SetValue(key string, value int) (err error) {
	mock.Call(m, storage.SetValue, key, value).Return(&err)
	return
}

// testIncrementValue a test of incrementValue function with using mocked storage interface
func testIncrementValue(t *testing.T) {
	cases := []struct {
		name      string
		setupMock func(storage)
		in        string
		out       int
		outErr    error
	}{
		{
			name: "success_scenario",
			setupMock: func(st storage) {
				mock.ExpectCall(st, storage.GetValue, "key").Return(123, nil)
				mock.ExpectCall(st, storage.SetValue, "key", 124)
			},
			in:  "key",
			out: 124,
		},
		{
			name: "on_GetValue_error",
			setupMock: func(st storage) {
				mock.ExpectCall(st, storage.GetValue).Return(0, fmt.Errorf("Error"))
			},
			outErr: fmt.Errorf("Error"),
		},
		{
			name: "on_SetValue_error",
			setupMock: func(st storage) {
				mock.ExpectCall(st, storage.GetValue)
				mock.ExpectCall(st, storage.SetValue).Return(fmt.Errorf("Error"))
			},
			outErr: fmt.Errorf("Error"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mockCore := mock.New(t)
			st := &mockStorage{mockCore}
			c.setupMock(st)
			newVal, err := incrementValue(c.in, st)

			if !reflect.DeepEqual(err, c.outErr) {
				t.Fatalf(`Error expected: "%v", got "%v"`, c.outErr, err)
			}

			if newVal != c.out {
				t.Fatalf(`Value expected: "%d", got: "%d"`, c.out, newVal)
			}

			mockCore.CheckExpectations()
		})
	}
}

func Example() {
}
