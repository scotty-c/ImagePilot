// Source: internal/docker_interface.go

// Package internal/mocks is a generated GoMock package.
package internalmocks

import (
	context "context"
	io "io"
	reflect "reflect"

	types "github.com/docker/docker/api/types"
	gomock "github.com/golang/mock/gomock"
)

// MockDockerClient is a mock of DockerClient interface.
type MockDockerClient struct {
	ctrl     *gomock.Controller
	recorder *MockDockerClientMockRecorder
}

// MockDockerClientMockRecorder is the mock recorder for MockDockerClient.
type MockDockerClientMockRecorder struct {
	mock *MockDockerClient
}

// NewMockDockerClient creates a new mock instance.
func NewMockDockerClient(ctrl *gomock.Controller) *MockDockerClient {
	mock := &MockDockerClient{ctrl: ctrl}
	mock.recorder = &MockDockerClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDockerClient) EXPECT() *MockDockerClientMockRecorder {
	return m.recorder
}

// ImageBuild mocks base method.
func (m *MockDockerClient) ImageBuild(ctx context.Context, context io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ImageBuild", ctx, context, options)
	ret0, _ := ret[0].(types.ImageBuildResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ImageBuild indicates an expected call of ImageBuild.
func (mr *MockDockerClientMockRecorder) ImageBuild(ctx, context, options interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ImageBuild", reflect.TypeOf((*MockDockerClient)(nil).ImageBuild), ctx, context, options)
}

// ImagePush mocks base method.
func (m *MockDockerClient) ImagePush(ctx context.Context, image string, options types.ImagePushOptions) (io.ReadCloser, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ImagePush", ctx, image, options)
	ret0, _ := ret[0].(io.ReadCloser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ImagePush indicates an expected call of ImagePush.
func (mr *MockDockerClientMockRecorder) ImagePush(ctx, image, options interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ImagePush", reflect.TypeOf((*MockDockerClient)(nil).ImagePush), ctx, image, options)
}
