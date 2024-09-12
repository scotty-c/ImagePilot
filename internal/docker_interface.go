package internal

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// DockerClient is an interface representing methods used from the Docker SDK client.
type DockerClient interface {
	ImageBuild(ctx context.Context, context io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error)
	ImagePush(ctx context.Context, image string, options types.ImagePushOptions) (io.ReadCloser, error)
}

// Ensure the actual Docker client implements this interface
var _ DockerClient = (*client.Client)(nil)
