package mock_example

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/unkeep/gomock/mock"
)

func TestIncrementValue(t *testing.T) {
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
			st := &storageMock{mockCore}
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
