.PHONY: build clean test release

build:
	@echo "Building ImagePilot..."
	@go build -o ImagePilot main.go

clean:
	@echo "Cleaning up..."
	@rm -f imagepilot
	@rm -f ./bin/*

test:
	@echo "Running tests..."
	go test -v ./...

fmt:
	@echo "Formatting code..."
	go fmt ./...

release: fmt test
	@echo "Building release..."
	@VERSION=$$(git tag --sort=-v:refname | head -n 1  ); \
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/ImagePilot-$$VERSION-linux-amd64 main.go; \
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ./bin/ImagePilot-$$VERSION-linux-arm64 main.go