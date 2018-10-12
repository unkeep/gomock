// gomock is golang mocking framework which provides a fast way of implementing light and flexible mock objects. It will help you to make your tests more useful and readable!
//
// mock - core package. Contains types and functions for implementing/setup mock objects
//
// gomock - commandline tool, generates and outputs a mock object from an interface
//
//
// Usage:
//
// Suggest you have an interface
//  type Storage interface {
//  	GetValue(key string) (int, error)
//  	SetValue(key string, value int) error
//  }
//
// And function you would like to test that depends on this interface
//  func IncrementValue(key string, st Storage) (int, error) {
//  	val, err := st.GetValue(key)
//  	if err != nil {
//  		return 0, err
//  	}
//  	val++
//  	if err := st.SetValue(key, val); err != nil {
//  		return 0, err
//  	}
//  	return val, nil
//  }
//
// You can declare a mock implementation by yourself or generate it by gomock tool
//  type mockStorage struct {
//  	mock.Core
//  }
//
//  func (m *mockStorage) GetValue(key string) (val int, err error) {
// 	 mock.Call(m, Storage.GetValue, key).Return(&val, &err)
// 	 return
//  }
//
//  func (m *mockStorage) SetValue(key string, value int) (err error) {
// 	 mock.Call(m, Storage.SetValue, key, value).Return(&err)
// 	 return
//  }
//
// The test of IncrementValue can be like that
//  func TestIncrementValue(t *testing.T) {
//  	cases := []struct {
//  		name      string
//  		setupMock func(Storage)
//  		in        string
//  		out       int
//  		outErr    error
//  	}{
//  		{
//  			name: "success_scenario",
//  			setupMock: func(st Storage) {
//  				mock.ExpectCall(st, Storage.GetValue, "key").Return(123, nil)
//  				mock.ExpectCall(st, Storage.SetValue, "key", 124)
//  			},
//  			in:  "key",
//  			out: 124,
//  		},
//  		{
//  			name: "on_GetValue_error",
//  			setupMock: func(st Storage) {
//  				mock.ExpectCall(st, Storage.GetValue).Return(0, fmt.Errorf("Error"))
//  			},
//  			outErr: fmt.Errorf("Error"),
//  		},
//  		{
//  			name: "on_SetValue_error",
//  			setupMock: func(st Storage) {
//  				mock.ExpectCall(st, Storage.GetValue)
//  				mock.ExpectCall(st, Storage.SetValue).Return(fmt.Errorf("Error"))
//  			},
//  			outErr: fmt.Errorf("Error"),
//  		},
//  	}
//
//  	for _, c := range cases {
//  		t.Run(c.name, func(t *testing.T) {
//  			mockCore := mock.New(t)
//  			st := &mockStorage{mockCore}
//  			c.setupMock(st)
//  			newVal, err := IncrementValue(c.in, st)
//
//  			if !reflect.DeepEqual(err, c.outErr) {
//  				t.Fatalf(`Error expected: "%v", got "%v"`, c.outErr, err)
//  			}
//
//  			if newVal != c.out {
//  				t.Fatalf(`Value expected: "%d", got: "%d"`, c.out, newVal)
//  			}
//
//  			mockCore.CheckExpectations()
//  		})
//  	}
//  }
package main

// blank imports help docs.
import (
	// mock package
	_ "github.com/unkeep/gomock/mock"
)
