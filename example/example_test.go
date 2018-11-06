package example

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/unkeep/gomock/mock"
)

// Storage the interface we will mock
type Storage interface {
	GetValue(key string) (int, error)
	SetValue(key string, value int) error
}

// IncrementValue the function we will test
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

// mockStorage the Storage mock. You can implement it yourself or using gomock help tool
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

// testIncrementValue a test of incrementValue function with using mocked storage interface
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
