// Code generated by mockery v1.0.0
package automock

import mock "github.com/stretchr/testify/mock"

// MinioClient is an autogenerated mock type for the MinioClient type
type MinioClient struct {
	mock.Mock
}

// BucketExists provides a mock function with given fields: bucketName
func (_m *MinioClient) BucketExists(bucketName string) (bool, error) {
	ret := _m.Called(bucketName)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(bucketName)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(bucketName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetBucketPolicy provides a mock function with given fields: bucketName
func (_m *MinioClient) GetBucketPolicy(bucketName string) (string, error) {
	ret := _m.Called(bucketName)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(bucketName)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(bucketName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MakeBucket provides a mock function with given fields: bucketName, location
func (_m *MinioClient) MakeBucket(bucketName string, location string) error {
	ret := _m.Called(bucketName, location)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(bucketName, location)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveBucket provides a mock function with given fields: bucketName
func (_m *MinioClient) RemoveBucket(bucketName string) error {
	ret := _m.Called(bucketName)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(bucketName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetBucketPolicy provides a mock function with given fields: bucketName, policy
func (_m *MinioClient) SetBucketPolicy(bucketName string, policy string) error {
	ret := _m.Called(bucketName, policy)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(bucketName, policy)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
