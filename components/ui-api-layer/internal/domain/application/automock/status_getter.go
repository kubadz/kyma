// Code generated by mockery v1.0.0
package automock

import gateway "github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/application/gateway"
import mock "github.com/stretchr/testify/mock"

// statusGetter is an autogenerated mock type for the statusGetter type
type statusGetter struct {
	mock.Mock
}

// GetStatus provides a mock function with given fields: reName
func (_m *statusGetter) GetStatus(reName string) gateway.Status {
	ret := _m.Called(reName)

	var r0 gateway.Status
	if rf, ok := ret.Get(0).(func(string) gateway.Status); ok {
		r0 = rf(reName)
	} else {
		r0 = ret.Get(0).(gateway.Status)
	}

	return r0
}
