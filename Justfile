vendor:
		go mod tidy
		go mod vendor

build:
		go build -o bin/local/devxy/hetznerrobot/4.0.0/darwin_arm64/terraform-provider-hetznerrobot_v4.0.0
