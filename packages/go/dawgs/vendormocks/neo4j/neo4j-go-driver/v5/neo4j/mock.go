// Copyright 2023 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/neo4j/neo4j-go-driver/v5/neo4j (interfaces: Result,Transaction,Session)

// Package neo4j is a generated GoMock package.
package neo4j

import (
	reflect "reflect"

	neo4j "github.com/neo4j/neo4j-go-driver/v5/neo4j"
	db "github.com/neo4j/neo4j-go-driver/v5/neo4j/db"
	gomock "go.uber.org/mock/gomock"
)

// MockResult is a mock of Result interface.
type MockResult struct {
	ctrl     *gomock.Controller
	recorder *MockResultMockRecorder
}

// MockResultMockRecorder is the mock recorder for MockResult.
type MockResultMockRecorder struct {
	mock *MockResult
}

// NewMockResult creates a new mock instance.
func NewMockResult(ctrl *gomock.Controller) *MockResult {
	mock := &MockResult{ctrl: ctrl}
	mock.recorder = &MockResultMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockResult) EXPECT() *MockResultMockRecorder {
	return m.recorder
}

// Collect mocks base method.
func (m *MockResult) Collect() ([]*db.Record, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Collect")
	ret0, _ := ret[0].([]*db.Record)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Collect indicates an expected call of Collect.
func (mr *MockResultMockRecorder) Collect() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Collect", reflect.TypeOf((*MockResult)(nil).Collect))
}

// Consume mocks base method.
func (m *MockResult) Consume() (neo4j.ResultSummary, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Consume")
	ret0, _ := ret[0].(neo4j.ResultSummary)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Consume indicates an expected call of Consume.
func (mr *MockResultMockRecorder) Consume() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Consume", reflect.TypeOf((*MockResult)(nil).Consume))
}

// Err mocks base method.
func (m *MockResult) Err() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Err")
	ret0, _ := ret[0].(error)
	return ret0
}

// Err indicates an expected call of Err.
func (mr *MockResultMockRecorder) Err() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Err", reflect.TypeOf((*MockResult)(nil).Err))
}

// Keys mocks base method.
func (m *MockResult) Keys() ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Keys")
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Keys indicates an expected call of Keys.
func (mr *MockResultMockRecorder) Keys() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Keys", reflect.TypeOf((*MockResult)(nil).Keys))
}

// Next mocks base method.
func (m *MockResult) Next() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Next")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Next indicates an expected call of Next.
func (mr *MockResultMockRecorder) Next() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Next", reflect.TypeOf((*MockResult)(nil).Next))
}

// NextRecord mocks base method.
func (m *MockResult) NextRecord(arg0 **db.Record) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NextRecord", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// NextRecord indicates an expected call of NextRecord.
func (mr *MockResultMockRecorder) NextRecord(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NextRecord", reflect.TypeOf((*MockResult)(nil).NextRecord), arg0)
}

// PeekRecord mocks base method.
func (m *MockResult) PeekRecord(arg0 **db.Record) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PeekRecord", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// PeekRecord indicates an expected call of PeekRecord.
func (mr *MockResultMockRecorder) PeekRecord(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PeekRecord", reflect.TypeOf((*MockResult)(nil).PeekRecord), arg0)
}

// Record mocks base method.
func (m *MockResult) Record() *db.Record {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Record")
	ret0, _ := ret[0].(*db.Record)
	return ret0
}

// Record indicates an expected call of Record.
func (mr *MockResultMockRecorder) Record() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Record", reflect.TypeOf((*MockResult)(nil).Record))
}

// Single mocks base method.
func (m *MockResult) Single() (*db.Record, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Single")
	ret0, _ := ret[0].(*db.Record)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Single indicates an expected call of Single.
func (mr *MockResultMockRecorder) Single() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Single", reflect.TypeOf((*MockResult)(nil).Single))
}

// MockTransaction is a mock of Transaction interface.
type MockTransaction struct {
	ctrl     *gomock.Controller
	recorder *MockTransactionMockRecorder
}

// MockTransactionMockRecorder is the mock recorder for MockTransaction.
type MockTransactionMockRecorder struct {
	mock *MockTransaction
}

// NewMockTransaction creates a new mock instance.
func NewMockTransaction(ctrl *gomock.Controller) *MockTransaction {
	mock := &MockTransaction{ctrl: ctrl}
	mock.recorder = &MockTransactionMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTransaction) EXPECT() *MockTransactionMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockTransaction) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockTransactionMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockTransaction)(nil).Close))
}

// Commit mocks base method.
func (m *MockTransaction) Commit() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Commit")
	ret0, _ := ret[0].(error)
	return ret0
}

// Commit indicates an expected call of Commit.
func (mr *MockTransactionMockRecorder) Commit() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Commit", reflect.TypeOf((*MockTransaction)(nil).Commit))
}

// Rollback mocks base method.
func (m *MockTransaction) Rollback() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Rollback")
	ret0, _ := ret[0].(error)
	return ret0
}

// Rollback indicates an expected call of Rollback.
func (mr *MockTransactionMockRecorder) Rollback() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Rollback", reflect.TypeOf((*MockTransaction)(nil).Rollback))
}

// Run mocks base method.
func (m *MockTransaction) Run(arg0 string, arg1 map[string]interface{}) (neo4j.Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Run", arg0, arg1)
	ret0, _ := ret[0].(neo4j.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Run indicates an expected call of Run.
func (mr *MockTransactionMockRecorder) Run(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockTransaction)(nil).Run), arg0, arg1)
}

// MockSession is a mock of Session interface.
type MockSession struct {
	ctrl     *gomock.Controller
	recorder *MockSessionMockRecorder
}

// MockSessionMockRecorder is the mock recorder for MockSession.
type MockSessionMockRecorder struct {
	mock *MockSession
}

// NewMockSession creates a new mock instance.
func NewMockSession(ctrl *gomock.Controller) *MockSession {
	mock := &MockSession{ctrl: ctrl}
	mock.recorder = &MockSessionMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSession) EXPECT() *MockSessionMockRecorder {
	return m.recorder
}

// BeginTransaction mocks base method.
func (m *MockSession) BeginTransaction(arg0 ...func(*neo4j.TransactionConfig)) (neo4j.Transaction, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "BeginTransaction", varargs...)
	ret0, _ := ret[0].(neo4j.Transaction)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BeginTransaction indicates an expected call of BeginTransaction.
func (mr *MockSessionMockRecorder) BeginTransaction(arg0 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BeginTransaction", reflect.TypeOf((*MockSession)(nil).BeginTransaction), arg0...)
}

// Close mocks base method.
func (m *MockSession) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockSessionMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockSession)(nil).Close))
}

// LastBookmark mocks base method.
func (m *MockSession) LastBookmark() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LastBookmark")
	ret0, _ := ret[0].(string)
	return ret0
}

// LastBookmark indicates an expected call of LastBookmark.
func (mr *MockSessionMockRecorder) LastBookmark() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LastBookmark", reflect.TypeOf((*MockSession)(nil).LastBookmark))
}

// LastBookmarks mocks base method.
func (m *MockSession) LastBookmarks() []string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LastBookmarks")
	ret0, _ := ret[0].([]string)
	return ret0
}

// LastBookmarks indicates an expected call of LastBookmarks.
func (mr *MockSessionMockRecorder) LastBookmarks() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LastBookmarks", reflect.TypeOf((*MockSession)(nil).LastBookmarks))
}

// ReadTransaction mocks base method.
func (m *MockSession) ReadTransaction(arg0 neo4j.TransactionWork, arg1 ...func(*neo4j.TransactionConfig)) (interface{}, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ReadTransaction", varargs...)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReadTransaction indicates an expected call of ReadTransaction.
func (mr *MockSessionMockRecorder) ReadTransaction(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadTransaction", reflect.TypeOf((*MockSession)(nil).ReadTransaction), varargs...)
}

// Run mocks base method.
func (m *MockSession) Run(arg0 string, arg1 map[string]interface{}, arg2 ...func(*neo4j.TransactionConfig)) (neo4j.Result, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Run", varargs...)
	ret0, _ := ret[0].(neo4j.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Run indicates an expected call of Run.
func (mr *MockSessionMockRecorder) Run(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockSession)(nil).Run), varargs...)
}

// WriteTransaction mocks base method.
func (m *MockSession) WriteTransaction(arg0 neo4j.TransactionWork, arg1 ...func(*neo4j.TransactionConfig)) (interface{}, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "WriteTransaction", varargs...)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WriteTransaction indicates an expected call of WriteTransaction.
func (mr *MockSessionMockRecorder) WriteTransaction(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteTransaction", reflect.TypeOf((*MockSession)(nil).WriteTransaction), varargs...)
}
