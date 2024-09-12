package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scotty-c/ImagePilot/internal"

	"github.com/docker/docker/client"
)

var (
	imageName   string
	imageTag    string
	registryURL string
	username    string
	password    string
	fromImage   string
	additional  []string
)

func init() {
	rootCmd.Flags().StringVarP(&imageName, "name", "n", "", "Name of the Docker image")
	rootCmd.Flags().StringVarP(&imageTag, "tag", "t", "latest", "Tag for the Docker image")
	rootCmd.Flags().StringVarP(&registryURL, "registry", "r", "", "Docker registry URL")
	rootCmd.Flags().StringVarP(&username, "username", "u", "", "Username for Docker registry")
	rootCmd.Flags().StringVarP(&password, "password", "p", "", "Password for Docker registry")
	rootCmd.Flags().StringVarP(&fromImage, "from", "f", "golang:1.20-alpine", "Base image for the Dockerfile")
	rootCmd.Flags().StringArrayVar(&additional, "add", []string{}, "Additional Dockerfile instructions in key=value format (e.g., --add 'RUN=apk add git')")
	viper.BindPFlag("name", rootCmd.Flags().Lookup("name"))
	viper.BindPFlag("tag", rootCmd.Flags().Lookup("tag"))
	viper.BindPFlag("registry", rootCmd.Flags().Lookup("registry"))
	viper.BindPFlag("username", rootCmd.Flags().Lookup("username"))
	viper.BindPFlag("password", rootCmd.Flags().Lookup("password"))
	viper.BindPFlag("from", rootCmd.Flags().Lookup("from"))
	viper.BindPFlag("add", rootCmd.Flags().Lookup("add"))

	rootCmd.MarkFlagRequired("name")
	rootCmd.MarkFlagRequired("registry")
}

var rootCmd = &cobra.Command{
	Use:   "imagepilot",
	Short: "ImagePilot automates Docker image creation and pushing",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🚀 Starting ImagePilot...")
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			log.Fatalf("❌ Failed to create Docker client: %v", err)
		}

		additionalPairs := make([]string, 0, len(additional))
		for _, item := range additional {
			// Handle splitting of key=value pairs
			keyValue := strings.SplitN(item, "=", 2)
			if len(keyValue) == 2 {
				additionalPairs = append(additionalPairs, fmt.Sprintf("%s %s", keyValue[0], keyValue[1]))
			} else {
				log.Fatalf("❌ Invalid format for --add flag: %s. Expected key=value", item)
			}
		}

		// Pass additionalPairs to the Dockerfile template
		if err := internal.CreateDockerfileTemplate(fromImage, additionalPairs); err != nil {
			log.Fatalf("❌ Failed to create Dockerfile: %v", err)
		}

		// Create a new context for Docker operations
		ctx := context.Background()

		// Initialize the progress bar
		bar := progressbar.NewOptions(100,
			progressbar.OptionSetDescription("Building and pushing Docker image..."),
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionShowCount(),
		)

		// Build and push the Docker image
		if err := internal.BuildAndPushImage(cli, ctx, imageName, imageTag, registryURL, username, password, bar); err != nil {
			log.Fatalf("❌ Error building or pushing image: %v", err)
		}
		fmt.Println("✅ Docker image built and pushed successfully.")
	},
}

func Execute() error {
	return rootCmd.Execute()
}
