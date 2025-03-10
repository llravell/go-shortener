// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/llravell/go-shortener/internal/usecase (interfaces: URLDeleteWorkerPool)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	usecase "github.com/llravell/go-shortener/internal/usecase"
)

// MockURLDeleteWorkerPool is a mock of URLDeleteWorkerPool interface.
type MockURLDeleteWorkerPool struct {
	ctrl     *gomock.Controller
	recorder *MockURLDeleteWorkerPoolMockRecorder
}

// MockURLDeleteWorkerPoolMockRecorder is the mock recorder for MockURLDeleteWorkerPool.
type MockURLDeleteWorkerPoolMockRecorder struct {
	mock *MockURLDeleteWorkerPool
}

// NewMockURLDeleteWorkerPool creates a new mock instance.
func NewMockURLDeleteWorkerPool(ctrl *gomock.Controller) *MockURLDeleteWorkerPool {
	mock := &MockURLDeleteWorkerPool{ctrl: ctrl}
	mock.recorder = &MockURLDeleteWorkerPoolMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockURLDeleteWorkerPool) EXPECT() *MockURLDeleteWorkerPoolMockRecorder {
	return m.recorder
}

// QueueWork mocks base method.
func (m *MockURLDeleteWorkerPool) QueueWork(arg0 *usecase.URLDeleteWork) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueueWork", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// QueueWork indicates an expected call of QueueWork.
func (mr *MockURLDeleteWorkerPoolMockRecorder) QueueWork(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueueWork", reflect.TypeOf((*MockURLDeleteWorkerPool)(nil).QueueWork), arg0)
}
