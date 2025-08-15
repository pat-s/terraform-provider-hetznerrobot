vendor:
		go mod tidy
		go mod vendor

build:
		go build -o bin/local/devxy/hetznerrobot/4.0.0/darwin_arm64/terraform-provider-hetznerrobot_v4.0.0

build_release:
		go build -o terraform-provider-hetznerrobot

build-all:
		@echo "Building for all platforms..."
		@mkdir -p dist
		GOOS=linux GOARCH=amd64 go build -o dist/terraform-provider-hetznerrobot_$(git tag -l | tail -n 1 | cut -c2-)_linux_amd64
		GOOS=linux GOARCH=arm64 go build -o dist/terraform-provider-hetznerrobot_$(git tag -l | tail -n 1 | cut -c2-)_linux_arm64
		GOOS=linux GOARCH=386 go build -o dist/terraform-provider-hetznerrobot_$(git tag -l | tail -n 1 | cut -c2-)_linux_386
		GOOS=darwin GOARCH=amd64 go build -o dist/terraform-provider-hetznerrobot_$(git tag -l | tail -n 1 | cut -c2-)_darwin_amd64
		GOOS=darwin GOARCH=arm64 go build -o dist/terraform-provider-hetznerrobot_$(git tag -l | tail -n 1 | cut -c2-)_darwin_arm64
		GOOS=windows GOARCH=amd64 go build -o dist/terraform-provider-hetznerrobot_$(git tag -l | tail -n 1 | cut -c2-)_windows_amd64.exe
		GOOS=windows GOARCH=386 go build -o dist/terraform-provider-hetznerrobot_$(git tag -l | tail -n 1 | cut -c2-)_windows_386.exe
		GOOS=windows GOARCH=arm64 go build -o dist/terraform-provider-hetznerrobot_$(git tag -l | tail -n 1 | cut -c2-)_windows_arm64.exe
		GOOS=freebsd GOARCH=amd64 go build -o dist/terraform-provider-hetznerrobot_$(git tag -l | tail -n 1 | cut -c2-)_freebsd_amd64
		GOOS=freebsd GOARCH=arm64 go build -o dist/terraform-provider-hetznerrobot_$(git tag -l | tail -n 1 | cut -c2-)_freebsd_arm64
		@echo "Build complete! Binaries are in ./dist/"
