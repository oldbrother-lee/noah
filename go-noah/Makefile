.PHONY: init
init:
	go install github.com/golang/mock/mockgen@latest
	go install github.com/swaggo/swag/cmd/swag@latest

.PHONY: bootstrap
bootstrap:
	cd ./deploy/docker-compose && docker compose up -d && cd ../../
	go run ./cmd/migration
	nunu run ./cmd/server

.PHONY: mock
mock:
	mockgen -source=internal/service/user.go -destination test/mocks/service/user.go
	mockgen -source=internal/repository/user.go -destination test/mocks/repository/user.go
	mockgen -source=internal/repository/repository.go -destination test/mocks/repository/repository.go

.PHONY: test
test:
	go test -coverpkg=./internal/handler,./internal/service,./internal/repository -coverprofile=./coverage.out ./test/server/...
	go tool cover -html=./coverage.out -o coverage.html

.PHONY: build
build:
	@echo "Building for current platform..."
	@mkdir -p ./bin
	go build -ldflags="-s -w" -o ./bin/server ./cmd/server
	@echo "Build completed: ./bin/server"

.PHONY: build-all
build-all: build-all-server build-all-task build-all-migration
	@echo "Build completed for all components and platforms!"

.PHONY: build-all-server
build-all-server:
	@echo "Building server for all platforms..."
	@mkdir -p ./bin
	@echo "Building server for linux/amd64..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/server-linux-amd64 ./cmd/server
	@echo "Building server for linux/arm64..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o ./bin/server-linux-arm64 ./cmd/server
	@echo "Building server for darwin/amd64..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/server-darwin-amd64 ./cmd/server
	@echo "Building server for darwin/arm64..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o ./bin/server-darwin-arm64 ./cmd/server
	@echo "Building server for windows/amd64..."
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/server-windows-amd64.exe ./cmd/server
	@echo "Building server for windows/arm64..."
	@CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o ./bin/server-windows-arm64.exe ./cmd/server
	@echo "Server build completed for all platforms!"

.PHONY: build-all-task
build-all-task:
	@echo "Building task for Linux platforms (task service typically runs on Linux servers)..."
	@mkdir -p ./bin
	@echo "Building task for linux/amd64..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/task-linux-amd64 ./cmd/task
	@echo "Building task for linux/arm64..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o ./bin/task-linux-arm64 ./cmd/task
	@echo "Task build completed for Linux platforms!"
	@echo "Note: darwin and windows builds are skipped due to CGO dependencies in gosigar package"

.PHONY: build-all-migration
build-all-migration:
	@echo "Building migration for all platforms..."
	@mkdir -p ./bin
	@echo "Building migration for linux/amd64..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/migration-linux-amd64 ./cmd/migration
	@echo "Building migration for linux/arm64..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o ./bin/migration-linux-arm64 ./cmd/migration
	@echo "Building migration for darwin/amd64..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/migration-darwin-amd64 ./cmd/migration
	@echo "Building migration for darwin/arm64..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o ./bin/migration-darwin-arm64 ./cmd/migration
	@echo "Building migration for windows/amd64..."
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/migration-windows-amd64.exe ./cmd/migration
	@echo "Migration build completed for all platforms!"

.PHONY: build-linux
build-linux:
	@echo "Building for linux/amd64..."
	@mkdir -p ./bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/server-linux-amd64 ./cmd/server
	@echo "Build completed: ./bin/server-linux-amd64"

.PHONY: build-darwin
build-darwin:
	@echo "Building for darwin/amd64..."
	@mkdir -p ./bin
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/server-darwin-amd64 ./cmd/server
	@echo "Build completed: ./bin/server-darwin-amd64"

.PHONY: build-windows
build-windows:
	@echo "Building for windows/amd64..."
	@mkdir -p ./bin
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/server-windows-amd64.exe ./cmd/server
	@echo "Build completed: ./bin/server-windows-amd64.exe"

.PHONY: build-task
build-task:
	@echo "Building task for current platform..."
	@mkdir -p ./bin
	go build -ldflags="-s -w" -o ./bin/task ./cmd/task
	@echo "Build completed: ./bin/task"

.PHONY: build-task-all
build-task-all: build-all-task

.PHONY: build-migration
build-migration:
	@echo "Building migration for current platform..."
	@mkdir -p ./bin
	go build -ldflags="-s -w" -o ./bin/migration ./cmd/migration
	@echo "Build completed: ./bin/migration"

.PHONY: build-migration-all
build-migration-all: build-all-migration

.PHONY: docker
docker:
	docker build -f deploy/build/Dockerfile --build-arg APP_RELATIVE_PATH=./cmd/task -t 1.1.1.1:5000/demo-task:v1 .
	docker run --rm -i 1.1.1.1:5000/demo-task:v1

.PHONY: swag
swag:
	swag init  -g cmd/server/main.go -o ./docs --parseDependency
