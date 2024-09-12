# ImagePilot

ImagePilot is a powerful CLI tool designed to automate Docker image creation and deployment. It simplifies the process of generating Dockerfiles, building Docker images, and pushing them to a specified Docker registry. With ImagePilot, you can streamline your container workflows with ease.

## Features

- **Automated Dockerfile Generation**: Create Dockerfiles on the fly by specifying the base image.
- **Seamless Image Building**: Build Docker images directly from the command line.
- **Easy Image Deployment**: Push your Docker images to any registry with minimal configuration.
- **Progress Tracking**: Visual feedback during image build and push operations with progress bars and terminal graphics.

## Installation

1. Clone the repository:
    ```bash
    git clone https://github.com/scotty-c/ImagePilot.git
    cd ImagePilot
    ```

2. Build the binary:
    ```bash
    make build
    ```

3. Run ImagePilot:
    ```bash
    ./ImagePilot
    ```

## Usage

ImagePilot is highly configurable via command-line flags. Below is an example of how to use it:  
Without your docker daemon logged into a registry

```bash
./ImagePilot --name myapp --tag v1.0.0 --registry registry.example.com --username myuser --password mypass --from golang:1.20-alpine --add "RUN=apk update && apk add make git" 
```

With your docker daemon logged in
```bash
./ImagePilot --name myapp --tag v1.0.0 --registry registry.example.com --from golang:1.20-alpine --add "RUN=apk update && apk add make git" 
```
### Flags

- `--name, -n`: The name of the Docker image.
- `--tag, -t`: The tag of the Docker image (default: `latest`).
- `--registry, -r`: The URL of the Docker registry.
- `--username, -u`: Your Docker registry username.
- `--password, -p`: Your Docker registry password.
- `--from, -f`: The base image to use in the Dockerfile (default: `golang:1.20-alpine`).
- `--add` : Additional Dockerfile instructions in key=value format (e.g., --add 'RUN=apk add git') to use multiples see the example `--add 'RUN=apk add git' --add 'CMD=/myapp'` 

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please submit pull requests or open issues for any features or bug fixes.

## Authors

- [Scott Coulton](https://github.com/scotty-c)

---

Enjoy using ImagePilot! 🚀
