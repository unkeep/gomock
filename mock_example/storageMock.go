package mock_example

import (
	"github.com/unkeep/gomock/mock"
)

type storageMock struct {
	mock.Core
}

func (st *storageMock) GetValue(key string) (val int, err error) {
	mock.Call(st, storage.GetValue, key).Return(&val, &err)
	return
}

func (st *storageMock) SetValue(key string, value int) (err error) {
	mock.Call(st, storage.SetValue, key, value).Return(&err)
	return
}
