package internal_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/golang/mock/gomock"
	"github.com/schollz/progressbar/v3"
	"github.com/scotty-c/ImagePilot/internal"
	internalmocks "github.com/scotty-c/ImagePilot/internal/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCreateDockerfileTemplate(t *testing.T) {
	// Cleanup Dockerfile after test
	defer os.Remove("Dockerfile")

	err := internal.CreateDockerfileTemplate("golang:1.18-alpine", []string{"RUN apk add --no-cache git", "CMD [\"go\", \"run\"]"})
	assert.NoError(t, err)

	// Verify the Dockerfile contents
	dockerfileContents, err := os.ReadFile("Dockerfile")
	assert.NoError(t, err)
	expected := `FROM golang:1.18-alpine
RUN apk add --no-cache git
CMD ["go", "run"]
`
	assert.Equal(t, expected, string(dockerfileContents))
}

func TestEncodeAuthConfig(t *testing.T) {
	authConfig := types.AuthConfig{
		Username: "testuser",
		Password: "testpass",
	}

	encoded, err := internal.EncodeAuthConfig(authConfig)
	assert.NoError(t, err)
	expected := "eyJ1c2VybmFtZSI6InRlc3R1c2VyIiwicGFzc3dvcmQiOiJ0ZXN0cGFzcyJ9"
	assert.Equal(t, expected, encoded)
}

func TestBuildAndPushImage(t *testing.T) {
	// Set up mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Use the generated mock
	mockClient := internalmocks.NewMockDockerClient(ctrl)

	// Mock context
	ctx := context.Background()

	// Mock the progress bar with a total value of 100
	bar := progressbar.New(100)

	// Create a temporary Dockerfile for testing
	tempDir, err := os.MkdirTemp("", "dockerfile-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	dockerfilePath := tempDir + "/Dockerfile"
	err = os.WriteFile(dockerfilePath, []byte("FROM scratch\n"), 0644)
	assert.NoError(t, err)

	// Change the working directory to the temporary directory
	originalDir, err := os.Getwd()
	assert.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	// Mock Docker build response
	buildResponse := io.NopCloser(bytes.NewReader([]byte(`{"stream":"Step 1/1 : FROM scratch\n"}`)))
	mockClient.EXPECT().ImageBuild(gomock.Any(), gomock.Any(), gomock.Any()).Return(types.ImageBuildResponse{Body: buildResponse}, nil)

	// Mock Docker push response
	pushResponse := io.NopCloser(bytes.NewReader([]byte(`{"status":"Pushed"}`)))
	mockClient.EXPECT().ImagePush(gomock.Any(), gomock.Any(), gomock.Any()).Return(pushResponse, nil)

	// Run the BuildAndPushImage function
	err = internal.BuildAndPushImage(mockClient, ctx, "library/go", "1.23.0-arm64", "registry.scottyc.work", "testuser", "testpass", bar)
	assert.NoError(t, err)

	// Validate that the progress bar reached 100%
	progress := bar.State().CurrentPercent
	assert.Equal(t, 1.0, progress, "Expected the progress bar to reach 100%%, but it reached %f", progress)
}
