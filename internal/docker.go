package internal

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/docker/docker/api/types"
	"github.com/schollz/progressbar/v3"
)

// DockerConfig holds the configuration for Docker authentication
type DockerConfig struct {
	Auths map[string]AuthConfig `json:"auths"`
}

// AuthConfig holds the authentication data for a Docker registry
type AuthConfig struct {
	Auth string `json:"auth"` // This is the base64 encoded username:password
}

// DockerfileData holds the fields for the Dockerfile template
type DockerfileData struct {
	BaseImage string
	Args      []string
}

// GetDockerCredentials checks if the registry credentials are stored locally and decodes them.
func GetDockerCredentials(registryURL string) (string, string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	dockerConfigPath := filepath.Join(homeDir, ".docker", "config.json")
	configData, err := os.ReadFile(dockerConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", fmt.Errorf("Docker config file not found")
		}
		return "", "", fmt.Errorf("failed to read Docker config file: %w", err)
	}

	var dockerConfig DockerConfig
	if err := json.Unmarshal(configData, &dockerConfig); err != nil {
		return "", "", fmt.Errorf("failed to parse Docker config file: %w", err)
	}

	// Normalize the registry URL (strip "https://" or "http://")
	normalizedRegistryURL := strings.TrimPrefix(strings.TrimPrefix(registryURL, "https://"), "http://")

	if authConfig, ok := dockerConfig.Auths[normalizedRegistryURL]; ok {
		// Decode the base64 encoded auth string
		decodedAuth, err := base64.StdEncoding.DecodeString(authConfig.Auth)
		if err != nil {
			return "", "", fmt.Errorf("failed to decode auth for %s: %w", normalizedRegistryURL, err)
		}

		// The decoded auth string is in the format "username:password"
		authParts := strings.SplitN(string(decodedAuth), ":", 2)
		if len(authParts) != 2 {
			return "", "", fmt.Errorf("invalid auth format for %s", normalizedRegistryURL)
		}
		fmt.Println()
		fmt.Printf("✅ Credentials found for %s: Username: %s\n", normalizedRegistryURL, authParts[0])
		return authParts[0], authParts[1], nil
	}

	fmt.Printf("No credentials found for %s in Docker config\n", normalizedRegistryURL)
	return "", "", fmt.Errorf("credentials for registry %s not found in Docker config", registryURL)
}

// BuildAndPushImage builds and pushes a Docker image to the specified registry
func BuildAndPushImage(cli DockerClient, ctx context.Context, imageName, imageTag, registryURL, username, password string, bar *progressbar.ProgressBar) error {
	// Check for local Docker credentials first
	localUsername, localPassword, err := GetDockerCredentials(registryURL)
	if err == nil {
		username = localUsername
		password = localPassword
		fmt.Println("✅ Using local Docker credentials")
	} else {
		fmt.Println("⚠️  Local Docker credentials not found or not usable, using provided credentials")
	}

	// Ensure we have credentials to proceed
	if username == "" || password == "" {
		return fmt.Errorf("❌ Docker credentials not provided or found for registry: %s", registryURL)
	}

	buildContextDir := "."
	dockerfilePath := filepath.Join(buildContextDir, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		return fmt.Errorf("❌ Dockerfile not found at path: %s", dockerfilePath)
	}

	tarReader, err := createTarFromDir(buildContextDir)
	if err != nil {
		return fmt.Errorf("❌ failed to create tar archive: %w", err)
	}

	fullImageName := fmt.Sprintf("%s/%s:%s", registryURL, imageName, imageTag)
	buildOptions := types.ImageBuildOptions{
		Tags:       []string{fullImageName},
		Dockerfile: "Dockerfile",
		Remove:     true,
	}
	fmt.Println("🔨 Building Docker image...")
	fmt.Println("📜 Build logs:")
	buildResponse, err := cli.ImageBuild(ctx, tarReader, buildOptions)
	if err != nil {
		return fmt.Errorf("❌ failed to build image: %w", err)
	}
	defer buildResponse.Body.Close()

	// Stream the build response for better readability
	if err := streamDockerBuildOutput(buildResponse.Body); err != nil {
		return fmt.Errorf("❌ failed to stream build response: %w", err)
	}

	// Adjust progress by 50% after build completion
	bar.Add(bar.GetMax() / 2)
	fmt.Println()
	fmt.Println("✅ Docker image built successfully.")

	authConfig := types.AuthConfig{
		Username: username,
		Password: password,
	}
	encodedAuth, err := EncodeAuthConfig(authConfig)
	if err != nil {
		return fmt.Errorf("❌ failed to encode auth config: %w", err)
	}

	pushOptions := types.ImagePushOptions{
		RegistryAuth: encodedAuth,
	}
	fmt.Println("📤 Pushing Docker image...")
	fmt.Println("📜 Push logs:")
	pushResponse, err := cli.ImagePush(ctx, fullImageName, pushOptions)
	if err != nil {
		return fmt.Errorf("❌ failed to push image: %w", err)
	}
	defer pushResponse.Close()

	// Stream the push response for better readability
	if err := streamDockerPushOutput(pushResponse); err != nil {
		return fmt.Errorf("❌ failed to stream push response: %w", err)
	}

	// Adjust progress to complete 100%
	bar.Add(bar.GetMax() / 2)
	fmt.Println()
	fmt.Println("✅ Docker image pushed successfully.")

	return nil
}

// streamDockerBuildOutput streams and logs Docker build output for better readability
func streamDockerBuildOutput(reader io.ReadCloser) error {
	decoder := json.NewDecoder(reader)
	for {
		var message map[string]interface{}
		if err := decoder.Decode(&message); err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if errMessage, ok := message["error"]; ok {
			fmt.Printf("❌ Build error: %v\n", errMessage)
		} else if stream, ok := message["stream"]; ok {
			fmt.Print(stream)
		}
	}
	return nil
}

// streamDockerPushOutput streams and logs Docker push output for better readability
func streamDockerPushOutput(reader io.ReadCloser) error {
	decoder := json.NewDecoder(reader)
	for {
		var message map[string]interface{}
		if err := decoder.Decode(&message); err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if errMessage, ok := message["error"]; ok {
			fmt.Printf("❌ Push error: %v\n", errMessage)
		} else if status, ok := message["status"]; ok {
			fmt.Println(status)
		}
	}
	return nil
}

// EncodeAuthConfig encodes Docker authentication configuration as a base64 JSON string.
func EncodeAuthConfig(authConfig types.AuthConfig) (string, error) {
	authJSON, err := json.Marshal(authConfig)
	if err != nil {
		return "", fmt.Errorf("❌ failed to marshal auth config: %w", err)
	}

	encodedAuth := base64.URLEncoding.EncodeToString(authJSON)
	return encodedAuth, nil
}

// createTarFromDir creates a tar archive from the specified directory.
func createTarFromDir(dirPath string) (io.Reader, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	defer tw.Close()

	err := filepath.WalkDir(dirPath, func(filePath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		fileInfo, err := d.Info()
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(fileInfo, filePath)
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(filePath)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !fileInfo.IsDir() {
			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(tw, file); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &buf, nil
}

// CreateDockerfileTemplate generates a Dockerfile using Go templating
func CreateDockerfileTemplate(baseImage string, additional []string) error {
	const dockerfileTemplate = `FROM {{.BaseImage}}
{{- range .Args }}
{{ . }}
{{- end }}
`

	// Prepare the Dockerfile data
	data := DockerfileData{
		BaseImage: baseImage,
		Args:      additional,
	}

	// Create a new template and parse the Dockerfile template
	tmpl, err := template.New("Dockerfile").Parse(dockerfileTemplate)
	if err != nil {
		return fmt.Errorf("❌ failed to parse Dockerfile template: %w", err)
	}

	// Create or overwrite the Dockerfile
	dockerfile, err := os.Create("Dockerfile")
	if err != nil {
		return fmt.Errorf("❌ failed to create Dockerfile: %w", err)
	}
	defer dockerfile.Close()

	// Execute the template with the provided data
	if err := tmpl.Execute(dockerfile, data); err != nil {
		return fmt.Errorf("❌ failed to execute Dockerfile template: %w", err)
	}

	fmt.Println("✅ Dockerfile created successfully.")
	return nil
}
