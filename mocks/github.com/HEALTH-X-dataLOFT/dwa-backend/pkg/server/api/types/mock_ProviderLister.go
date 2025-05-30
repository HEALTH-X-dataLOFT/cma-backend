// Copyright 2025 HEALTH-X dataLOFT
//
// Licensed under the European Union Public Licence, Version 1.2 (the
// "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://eupl.eu/1.2/en/
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by mockery v2.52.1. DO NOT EDIT.

package types

import (
	context "context"

	types "github.com/HEALTH-X-dataLOFT/cma-backend/pkg/server/api/types"
	mock "github.com/stretchr/testify/mock"
)

// MockProviderLister is an autogenerated mock type for the ProviderLister type
type MockProviderLister struct {
	mock.Mock
}

type MockProviderLister_Expecter struct {
	mock *mock.Mock
}

func (_m *MockProviderLister) EXPECT() *MockProviderLister_Expecter {
	return &MockProviderLister_Expecter{mock: &_m.Mock}
}

// GetProvider provides a mock function with given fields: ctx, providerID
func (_m *MockProviderLister) GetProvider(ctx context.Context, providerID string) (types.Provider, error) {
	ret := _m.Called(ctx, providerID)

	if len(ret) == 0 {
		panic("no return value specified for GetProvider")
	}

	var r0 types.Provider
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (types.Provider, error)); ok {
		return rf(ctx, providerID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) types.Provider); ok {
		r0 = rf(ctx, providerID)
	} else {
		r0 = ret.Get(0).(types.Provider)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, providerID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockProviderLister_GetProvider_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetProvider'
type MockProviderLister_GetProvider_Call struct {
	*mock.Call
}

// GetProvider is a helper method to define mock.On call
//   - ctx context.Context
//   - providerID string
func (_e *MockProviderLister_Expecter) GetProvider(ctx interface{}, providerID interface{}) *MockProviderLister_GetProvider_Call {
	return &MockProviderLister_GetProvider_Call{Call: _e.mock.On("GetProvider", ctx, providerID)}
}

func (_c *MockProviderLister_GetProvider_Call) Run(run func(ctx context.Context, providerID string)) *MockProviderLister_GetProvider_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockProviderLister_GetProvider_Call) Return(_a0 types.Provider, _a1 error) *MockProviderLister_GetProvider_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockProviderLister_GetProvider_Call) RunAndReturn(run func(context.Context, string) (types.Provider, error)) *MockProviderLister_GetProvider_Call {
	_c.Call.Return(run)
	return _c
}

// GetProviderURL provides a mock function with given fields: ctx, providerID
func (_m *MockProviderLister) GetProviderURL(ctx context.Context, providerID string) (string, error) {
	ret := _m.Called(ctx, providerID)

	if len(ret) == 0 {
		panic("no return value specified for GetProviderURL")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (string, error)); ok {
		return rf(ctx, providerID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, providerID)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, providerID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockProviderLister_GetProviderURL_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetProviderURL'
type MockProviderLister_GetProviderURL_Call struct {
	*mock.Call
}

// GetProviderURL is a helper method to define mock.On call
//   - ctx context.Context
//   - providerID string
func (_e *MockProviderLister_Expecter) GetProviderURL(ctx interface{}, providerID interface{}) *MockProviderLister_GetProviderURL_Call {
	return &MockProviderLister_GetProviderURL_Call{Call: _e.mock.On("GetProviderURL", ctx, providerID)}
}

func (_c *MockProviderLister_GetProviderURL_Call) Run(run func(ctx context.Context, providerID string)) *MockProviderLister_GetProviderURL_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockProviderLister_GetProviderURL_Call) Return(_a0 string, _a1 error) *MockProviderLister_GetProviderURL_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockProviderLister_GetProviderURL_Call) RunAndReturn(run func(context.Context, string) (string, error)) *MockProviderLister_GetProviderURL_Call {
	_c.Call.Return(run)
	return _c
}

// ListProviders provides a mock function with given fields: ctx
func (_m *MockProviderLister) ListProviders(ctx context.Context) ([]types.Provider, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for ListProviders")
	}

	var r0 []types.Provider
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]types.Provider, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []types.Provider); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]types.Provider)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockProviderLister_ListProviders_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListProviders'
type MockProviderLister_ListProviders_Call struct {
	*mock.Call
}

// ListProviders is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockProviderLister_Expecter) ListProviders(ctx interface{}) *MockProviderLister_ListProviders_Call {
	return &MockProviderLister_ListProviders_Call{Call: _e.mock.On("ListProviders", ctx)}
}

func (_c *MockProviderLister_ListProviders_Call) Run(run func(ctx context.Context)) *MockProviderLister_ListProviders_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockProviderLister_ListProviders_Call) Return(_a0 []types.Provider, _a1 error) *MockProviderLister_ListProviders_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockProviderLister_ListProviders_Call) RunAndReturn(run func(context.Context) ([]types.Provider, error)) *MockProviderLister_ListProviders_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockProviderLister creates a new instance of MockProviderLister. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockProviderLister(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockProviderLister {
	mock := &MockProviderLister{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
